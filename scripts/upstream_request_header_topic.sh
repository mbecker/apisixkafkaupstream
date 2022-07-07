#!/bin/bash
# The plugin is configured  with the global topic "kafkaupstream" 
# This request inlucdes the header "topic: kafkaupstream2"
# This request tests that the key is taken from the header
curl --location --request POST 'http://localhost:9080/kafkaupstream' \
--header 'topic: kafkaupstream2' \
--header 'Content-Type: application/json' \
--data-raw '{
    "glossary": {
        "title": "example glossary",
		"GlossDiv": {
            "title": "The Kafka Message ID configured globally by jsonkey",
			"GlossList": {
                "GlossEntry": {
                    "ID": "SGML",
					"SortAs": "SGML",
					"GlossTerm": "Standard Generalized Markup Language",
					"Acronym": "SGML",
					"Abbrev": "ISO 8879:1986",
					"GlossDef": {
                        "para": "A meta-markup language, used to create markup languages such as DocBook.",
						"GlossSeeAlso": ["GML", "XML"]
                    },
					"GlossSee": "markup"
                }
            }
        }
    }
}'