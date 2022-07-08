# Apixis Kafka Upstream (go plugin)
<p align="center">Buy me a coffee and support me :-)</p>
<p align="center">

[![paypal](https://raw.githubusercontent.com/stefan-niedermann/paypal-donate-button/master/paypal-donate-button.png)](https://paypal.me/matsbecker?country.x=DE&locale.x=de_DE)
</p>


This git repository is an [Apache Apisix](https://apisix.apache.org/) [go(lang) plugin](https://apisix.apache.org/docs/go-plugin-runner/getting-started/) to have a plugin which
- acts as an upstream service
- accepts http(s) requests with (json) body
- produces a Kafka message with the http body to a topic

The process / communication flow is as follows
```sh
http(s) request -> Apisix (-> Apisix Plugins) -> Apisix Kafka Upstream Plugin -> Kafka Borker / Topic
```

The configuration for the Kafka broker, topic, and messages can be either configured for the plugin globally and / or for each http(s) request. By sending the correct http headers (see plugin configuration for more details) the plugin uses the header attributes as the config rather the global configuration.

The kakfa upstream plugin produce messages to the topic asynchronously. The http response is sent back immediately as the Kafka message is produced. This means, the http response does not guarantee that the message is acknowledged by the Kafka broker. The returned http status 200 defines
- The http(s) request + (json) body is ok
- The global config and http header config is ok
- The Kafka connection configuration is ok and a Kakfa connection is established
-  The Kafka message with header like partition, key, topic is produced

The most simple configraution and use case is that the kafka usptream plugin takes the http body and sends it to a Kafka topic as is. By setting the the header "content-type=application/json" the plugin takes and elides the body bytes with insignificant space characters by using the standard library [json.Compact](https://pkg.go.dev/encoding/json#Compact).

## Build

The code includes the library [confluent kafka go](github.com/confluentinc/confluent-kafka-go) to prouduces messages to Kafka. The plugin is a wrapper for a C-library called [librdkafka](https://github.com/edenhill/librdkafka).

By using an golang:alpine-Image the C-library must be compiled with alpine's C-compiler (musl) and the go build command must include the tags

```sh
# Image: golang:1.18.3-alpine
go build -tags musl --ldflags "-extldflags -static" -o kafkaupstream .
```

## Dockerfile

The Dockerfile uses a multi-stage build process to
- (Stage: base) Download and build the C-library librdkafka
- (Stage: build) Set the correct Go-Build-Env and Tags + Build the go project
- (Stage: last) Use the default apisix-alpine Image and copy the go binary to it's image

> I'm using a Macbook Pro M1 and having some problems withe the C-library librdkafka and alpine's C-compiler musl and Docker to build multi-arch images. I'm using the tag "--platform=linux/amd64" in the Dockerfile to target a specific platform to compile the C-library correctly. Any ideas and hints how to build a multi-arch image would be very helpful

Use the Dockerfile and the appropriate docker-commands to build and tag the image.

```sh
./build.sh
OR
docker build . -t matsbecker/apisixkafkaupstream:v1 -t matsbecker/apisixkafkaupstream:latest
```

# Apisix / plugin deployment

The Apisix go(lang) must be used as defined in the [Apache Apisix documentation](https://apisix.apache.org/docs/go-plugin-runner/getting-started/):

- Use the image built with the Dockerfile / or copy the go(lang) binary to your Apisix-deployment/image
- Define the Apisix configuration to point to the go(lang) binary

The config for Apache Apisix to use the plugin is as follows:
```yaml
ext-plugin:
  cmd:
    - "/usr/local/apisix/plugins/kafkaupstream"
```

## Docker compose

The directory "example" includes the docker compose files / folders from the original [Apache Apisix Docker git repo](https://github.com/apache/apisix-docker/tree/master/example). The changes to use the plugin kafkaupstream in the Apisix gateway are as follows:

- The configuration for the Apisix gateway is updated as described above (see file: "examples/apisix_conf/config.yaml")
- The Docker compose file is updated to build the Docker image as described above (alternativ: use the already built image which is pushed to Docker hub)
- Scripts added to start/stop the containers

** The docker compose file is for the platform arm64

# Configuration

Plugins in Apache Apisx are configured for routes. For the general syntax and configuration requests have a look at the Apache Apisx documentation.

Important: The configuration for custom plugins are JSON strings and not JSON objects.

The follwing table defines the configuration for the plugin. The plugin must be configured globally in Apisix by enabling and configuring the plugin globally (for example Kafka connection configuration must be configured globally). Each http request may include some http headers defined below which are used for example to produce the message with a special key or partition.

The usage of global vs. http header configuration is as follows:
```sh
http header > Global
* The http header attributes overwrites the global configuration

Global(jsonkey) > Global(key)
* The global jsonkey overwrites the global key

http header(key) > Global(jsonkey)
* The http header "key" overwrites the global configuration "jsonkey"
```


| Configraution        	| Global / http header 	| Constants          	| Description                                                                                                                                                                                                                                                                                                                                                                                                                                 	|
|----------------------	|----------------------	|--------------------	|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------	|
| config: object       	| Global               	| -                  	| Defines the Kafka connection configuration.  See [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go) for all connection configurations.                                                                                                                                                                                                                                                                                	|
| headerprefix: string 	| Global               	| -                  	| For each http request the plugin takes some header attributes for producing the Kafka message (like partition, topic, key). To use a header prefix for each of the http header keys set the "headerprefix". By using for example the headerprefix "cool-" the header key "partition" becomes "cool-partition" ("cool-topic", "cool-key").                                                                                                   	|
| partition: int       	| Global + http header 	| -1: Any Partition  	| Defines the Kafka topic partition to which the message is produced. By setting the key to "-1" the message is produced with the flag "any partition".                                                                                                                                                                                                                                                                                       	|
| topic: string        	| Global + http header 	| -                  	| Defines the topic to which the message is produced.                                                                                                                                                                                                                                                                                                                                                                                         	|
| key: string          	| Global + http header 	| -                  	| Defines the key which should be used to produce the message.                                                                                                                                                                                                                                                                                                                                                                                	|
| jsonkey: string      	| Global               	| -                  	| By defining the configuration "jsonkey" the plugin gets the key from the JSON message. The plugin [github.com/tidwall/gjson](github.com/tidwall/gjson) is used to get the json value (use the syntax of the plugin)<br/><br/><strong>Important:</strong> The http header "content-type=application/json" must be set.                                                                                                                                                                                                                                                                              	|
| content-type: string 	| http header          	| "application/json" 	| Standard http-header "content-type=application/json". By sending this header the plugin uses the standard library json.Compact. This elides the JSON-encoded body with insignificant space characters. This reduces the bytes sent to Kafka (but may be slower than just sending the body bytes as is). Sending an invalid json body may result in unexpected behaviour; the error response should be a http status code 400.<br/><br/><strong>Important:</strong> This http-header must be sent to use the configration "jsonkey". 	|

## Configuration Example (global) for upstash

> Replace the cofiguration details in the following files / codes with your details

In the directory "scripts" the following files can be used to configure and test the plugin:

```sh
scripts/
├─ .env --> The local .env-file file which sets the environment parameter or substitute the key-strings in the JSON file
├─ .env.backup --> Sample .env-File; rename to ".env" to be used in the bash scripts
├─ kafkaupstreamconfig.json --> The JSON config for the Apisix plugin kafkausptream 
├─ kcat.sh --> kcat (kafkacat) to connect to Kafka brokers defined in the .env-file
├─ upstream_create.sh --> Create / Update an Apisix upstream with enabled plugin kafkausptream configured in the JSON file
├─ upstream_request.sh --> Test the Apisix upstream by requesting it with curl (hhtp) and a sample body
├─ upstream_request_*.sh --> Additional test requests with header keys or complex key objects in the JSON body
```

The bash-scripts uses the ".env"-file to either set environemnt parameters used in kcat to connect to the Kafka broker / topic and to substitute the JSON config with the key/value-pair. The idea is to have a single source of configuration to (1) connect to Kafka via kcat and (2) create / update Apisix route.

![Apisix kafkaupstream test](./scripts/kafkausptream.gif)

## Development

I'm using a Mcbook Pro M1 with an arm64-architecture. The C-library to communicate with Kafka must be compiled locally to have no go-errors in your project. This results in a manual installation of some libarries with homebrew and linking the go complile to correct libraries. See in the folder ".vscode" the settings for Visual Stude and go for my environment.

The local dependencies are installed as follows:
```sh
brew install pkg-config
brew install librdkafka
brew install openssl
```

The go project must be built or tested with dynamic tags:
```sh
go test -v utils/json_test.go utils/json.go -bench=. --tags dynamic
OR
go build -o kafkaupstream --tags dynamic
```

For more details see: [https://github.com/confluentinc/confluent-kafka-go/issues/591](https://github.com/confluentinc/confluent-kafka-go/issues/591)