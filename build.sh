#!/bin/bash
# Use the following tags for removing intermediate containers and not using cahc when building the image
# --force-rm		Always remove intermediate containers
# --no-cache		Do not use cache when building the image
docker build . -t matsbecker/apisixkafkaupstream:v1 -t matsbecker/apisixkafkaupstream:latest