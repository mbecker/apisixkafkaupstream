#!/bin/bash
# The plugin is configured  with the jsonkey "glossary.GlossDiv.title" 
# This request tests that the key is a JSON object rather a primitive type like string, int, bool
curl --location --request POST 'http://localhost:9080/kafkaupstream' \
--header 'Content-Type: application/json' \
--data-raw '{
    "glossary": {
        "title": "example glossary",
		"GlossDiv": {
            "title": {
                "key": "Test key for jsonkey to get a complete object as a kafka message key",
                "test": "vlaue"
            },
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