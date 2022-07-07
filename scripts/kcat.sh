# if the .env file exists, source it
[ -f .env ] && . .env
# connect to kafka via kcat with the environemnt parameters
kcat -b ${KAFKA_BOOTSTRAP_SERVERS} \
-X security.protocol=${KAFKA_SECURITY_PROTOCOL} -X sasl.mechanisms=${KAFKA_SASL_MECHANISMS} \
-X sasl.username=${KAFKA_SALS_USERNAME}  -X sasl.password=${KAFKA_SASL_PASSWORD}  \
-L -t ${KAFKA_TOPIC} -C -f '\nKey (%K bytes): %k\t\nValue (%S bytes): %s\nTimestamp: %T\tPartition: %p\tOffset: %o\n--\n'