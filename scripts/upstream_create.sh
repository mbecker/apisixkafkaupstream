#!/bin/bash

#
# Files: ".env", "kafkaupstreamconfig.json"
#
# The .env-file is used to set env parameters to either use it for the kafka connection with kcat and
# to substitute the JSON config for the plugin kafkausptream in the file "kafkaupstreamconfig.json".
# The substitued JSON file "kafkaupstreamconfig.json" is used as a json string to create / update a specific Apisix route
#
# The expected http result status coide is 200.
#

# if the .env file exists, source it
# [ -f .env ] && . .env

# read the json file and "stringify" it
JSONSTRING=$(jq tostring ./kafkaupstreamconfig.json)

# read the ".env"-file line by line; split each line and use the pair key/value to substitute the json string
# TODO: The last line is not read properly
# FIX: Add a test line to the end of .env-file
while read line; do
    # reading each line of the .env-file
    #echo "Reading line: " $line
    # split the line by '="'
    kvs=($(echo "$line" | tr '="' '\n'))
    echo "Substitute: key=${kvs[0]} key=${kvs[1]}"
    # substitue the JSON string with key/value from the .env file line by line
    JSONSTRING=$(echo $JSONSTRING | sed 's/'${kvs[0]}'/'${kvs[1]}'/')
done < .env

# Create / update the route at Apisix with JSON string for the configuration substitued by the .env-file
curl --verbose --location --request PUT 'http://127.0.0.1:9080/apisix/admin/routes/kafkaupstream' \
--header 'x-api-key: edd1c9f034335f136f87ad84b625c8f1' \
--header 'Content-Type: application/json' \
--data-raw '{
    "methods": [
        "POST"
    ],
    "name": "kafkaupstream",
    "host": "localhost",
    "uri": "/kafkaupstream",
    "plugins": {
        "ext-plugin-pre-req": {
            "conf": [
                {
                    "name": "kafkaupstream",
                    "value": '"${JSONSTRING}"'
                }
            ]
        }
    }
}'
