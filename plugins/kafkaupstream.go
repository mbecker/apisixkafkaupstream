/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugins

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	kafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/elgs/gojq"
)

const HeaderKey string = "key"
const HeaderPartition string = "partition"
const HeaderTopic string = "topic"
const HeaderContentType string = "content-type"
const HeaderContentTypeApplicationJson = "application/json"
const HeaderPartitionAny int32 = -1

var kProducer *kafka.Producer
var deliveryChan chan kafka.Event
var kConfig kafka.ConfigMap

type KafkaUpstream struct {
	// Embed the default plugin here,
	// so that we don't need to reimplement all the methods.
	plugin.Plugin
}

type KafkaUpstreamResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type KafkaUpstreamConf struct {
	Conf         map[string]interface{} `json:"config"`
	JsonKey      string                 `json:"jsonkey"`
	Topic        string                 `json:"topic"`
	Partition    int32                  `json:"partition"`
	Key          string                 `json:"key"`
	HeaderPrefix string                 `json:"headerprefix"`
}

type KafkaMessageAttr struct {
	partition int32
	topic     string
	key       string
}

func getHeaderWithPrefix(headerPrefix, headerKey string) string {
	return fmt.Sprintf("%s%s", headerPrefix, headerKey)
}

func getHeaderAttrs(header pkgHTTP.Header, conf KafkaUpstreamConf) KafkaMessageAttr {
	var attr KafkaMessageAttr

	// Kafka Topic
	attr.topic = conf.Topic
	hTopic := header.Get(getHeaderWithPrefix(conf.HeaderPrefix, HeaderTopic))
	if len(hTopic) > 0 {
		attr.topic = hTopic
		log.Infof("Header value found: key=%s value=%s", HeaderTopic, attr.topic)
	}

	// Kafka Partition
	attr.partition = conf.Partition
	hPartition, err := strconv.ParseInt(header.Get(getHeaderWithPrefix(conf.HeaderPrefix, HeaderPartition)), 10, 32)
	if err == nil {
		attr.partition = int32(hPartition)
		log.Infof("Header value found: key=%s value=%d", HeaderPartition, attr.partition)
	}
	if attr.partition == HeaderPartitionAny {
		attr.partition = kafka.PartitionAny
		log.Infof("Header '%s' is kafka.PartitionAny=%d", HeaderPartition, attr.partition)
	}

	// Kafka Message key
	attr.key = conf.Key
	hKey := header.Get(getHeaderWithPrefix(conf.HeaderPrefix, HeaderKey))
	if len(hKey) > 0 {
		attr.key = hKey
		log.Infof("Header value found: key=%s value=%d", HeaderKey, hKey)
	}

	return attr

}

func init() {
	log.Infof("KafkaUpstream init()")
	err := plugin.RegisterPlugin(&KafkaUpstream{})
	if err != nil {
		log.Fatalf("failed to register plugin KafkaUpstream: %s", err)
	}

	// Initialize delivery report handler for kafka's produced messages
	deliveryChan = make(chan kafka.Event)
	go func() {
		for e := range deliveryChan {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Errorf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					log.Infof("Produced event: topic=%s partition=%d key=%-10s value=%s timestamp=%s\n",
						*ev.TopicPartition.Topic, ev.TopicPartition.Partition, string(ev.Key), string(ev.Value), ev.Timestamp)
				}
			}
		}
	}()
}

func (p *KafkaUpstream) Name() string {
	log.Infof("KafkaUpstream Name()")
	return "kafkaupstream"
}

func (p *KafkaUpstream) ParseConf(in []byte) (interface{}, error) {
	log.Infof("KafkaUpstream ParseConf()")
	conf := KafkaUpstreamConf{}
	err := json.Unmarshal(in, &conf)

	if err != nil {
		log.Errorf("Error unmarshaling plugin config: %s", err)
		return conf, err
	}
	log.Infof("Plugin json config: %s", conf)

	// Initialite kafka configmap
	kConfig = kafka.ConfigMap{}
	// Range the give JSON config of the plugin and set the kafka configmap
	// Only the following types are allowd: string,bool,int,ConfigMap
	// Cast any number to type int
	// TODO: Cast JSON map/object to ConfigMap
	for k, v := range conf.Conf {
		log.Infof("Set Kafka ConfigMap: value=%s key=%s", k, v)
		switch nv := v.(type) {
		case float32:
			kConfig.SetKey(k, int(nv))
		case float64:
			kConfig.SetKey(k, int(nv))
		case int8:
			kConfig.SetKey(k, int(nv))
		case int16:
			kConfig.SetKey(k, int(nv))
		case int32:
			kConfig.SetKey(k, int(nv))
		case int64:
			kConfig.SetKey(k, int(nv))
		default:
			kConfig.SetKey(k, v)
		}
	}

	kProducer, err = kafka.NewProducer(&kConfig)
	if err != nil {
		log.Errorf("Error connecting to Kafka broker: %s", err)
		return conf, err
	}
	// defer conf.KProducer.Close()

	return conf, err
}

func (p *KafkaUpstream) Filter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {
	log.Infof("KafkaUpstream Filter()")
	var err error

	if kProducer == nil {
		log.Infof("Kafka producer is nil; connecting to kafka broker")
		kProducer, err = kafka.NewProducer(&kConfig)
		if err != nil {
			log.Errorf("Error connecting to Kafka broker: %s", err)
			p.writeMessage(w, 400, "Error connecting to Kafka broker")
			return
		}
	}

	// Request Header attributes (partition, key, topic)
	attrs := getHeaderAttrs(r.Header(), conf.(KafkaUpstreamConf))

	// Request body ("application/json")
	body, err := r.Body()
	if err != nil {
		log.Errorf("Error fetching request body: %s", err)
		p.writeMessage(w, 400, "Error fetching request body")
		return
	}
	// The header "content-type=application/json" is set explicetly; (un)marshals the body bytes'
	if r.Header().Get(HeaderContentType) == HeaderContentTypeApplicationJson {
		var parser *gojq.JQ
		var i interface{}
		err = json.Unmarshal(body, &i)
		if err != nil {
			log.Errorf("Error unmarshaling request body: %s", err)
			p.writeMessage(w, 400, "Error unmarshaling request body")
			return
		}
		body, err = json.Marshal(i)
		if err != nil {
			log.Errorf("Error marshaling request body: %s", err)
			p.writeMessage(w, 400, "Error marshaling request body")
			return
		}
		// Check that http header "key" is not set and that the  global conf "jsonkey" is set; if yes get the key from the JSON body
		if len(attrs.key) == 0 && len(conf.(KafkaUpstreamConf).JsonKey) > 0 {
			parser = gojq.NewQuery(i)
			key, err := parser.QueryToString(conf.(KafkaUpstreamConf).JsonKey)
			if err != nil {
				log.Errorf("Error getting json key with JQ querytostring")
			} else {
				attrs.key = key
			}
		}

	}

	// Produce messages to topic (asynchronously)
	log.Infof("Sending Kafka message ::: partition=%d - topic=%s - key=%s - message=%s", attrs.partition, attrs.topic, attrs.key, string(body))
	kProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &attrs.topic, Partition: attrs.partition},
		Key:            []byte(attrs.key),
		Value:          body,
	}, deliveryChan)

	p.writeMessage(w, 200, "All good!")
	return

}

func (p *KafkaUpstream) writeMessage(w http.ResponseWriter, status int, message string) {
	log.Infof("Sending / writing response message to client: %s", message)
	e := KafkaUpstreamResponse{
		Message: message,
		Code:    status,
	}
	b, err := json.Marshal(e)
	if err != nil {
		log.Errorf("failed to marshal error message: %s", err)
		return
	}
	// Write response header status code and body
	w.WriteHeader(e.Code)
	w.Header().Add(HeaderContentType, HeaderContentTypeApplicationJson)
	_, err = w.Write(b)
	if err != nil {
		log.Errorf("failed to write: %s", err)
	}
	return
}

// isNumeric checks if a given value is a number (float64)
func isNumeric(s interface{}) bool {
	_, err := strconv.ParseFloat(fmt.Sprintf("%f", s), 64)
	return err == nil
}
