We use confluentinc docker images because, the founders of the this company have developed the kafka it self, and that had their own version of kafka images, which are pretty solid as they are exists for many years. But after few years, we can actually switch to apache/kafka images, which actaully will work without zookeeper as thay are documented in the official website, and configurations will be pretty the same, so if you understand this one you can easily switch to another one in the future. 

So we need three things:
- Zookeeper - helper tool for kafka borker to manage partitions and etc.
- Apache Kafka - the core tool for out pub/sub
- Kafka UI by provectuslabs - the guys made cool open source tool to manage your topics and message visually

So first we start by setting zookeeper:

#### Zookeeper

```yml
version: '3.8'

services:
  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    container_name: zookeeper
    ports:
      - "2181:2181"
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    volumes:
      - ./zookeeper-data:/var/lib/zookeeper/data
    networks:
      - default

networks:
  default:
    driver: bridge
```

Image `confluentinc/cp-zookeeper:latest` 
ZOOKEEPER_CLIENT_PORT - specify the port Zookeeper will run
ZOOKEEPER_TICK_TIME - basic time unit for any operation by Zookeeper 

#### Kafka

```yml
  kafka:
    image: confluentinc/cp-kafka:latest
    container_name: kafka
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1

      KAFKA_LISTENERS: EXTERNAL://0.0.0.0:9092, INTERNAL://0.0.0.0:29092
      KAFKA_ADVERTISED_LISTENERS: EXTERNAL://localhost:9092, INTERNAL://kafka:29092
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: EXTERNAL:PLAINTEXT, INTERNAL:PLAINTEXT
      KAFKA_INTER_BROKER_LISTENER_NAME: INTERNAL
    volumes:
      - ./kafka-data:/var/lib/kafka/data
    networks:
      - default

```

Image: confluentinc/cp-kafka:latest
**KAFKA_BROKER_ID**: 1 - we can have many borker so we need assign id (even if we have only one)
**KAFKA_ZOOKEEPER_CONNECT**: zookeeper:2181
**KAFKA_AUTO_CREATE_TOPICS_ENABLE**: "true" - autocreate the topic
**KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR**: 1 - tells how many relocation of client message to hold, which is equal to how many brokers you have.  

In Kafka you can create as many listeners as possible, using **KAFKA_LISTENERS** by labeling them in the following syntext:

`LABEL_NAME://0.0.0.0:9092, LABEL_NAME2://0.0.0.0:9092`

After you do it, you must register this listeners in advertiser, using **KAFKA_ADVERTISED_LISTENERS**,  which tells the clients what address to connect for each listener:

`LABEL_NAME://{host, kafka_container_name}:{port}, LABEL_NAME2://{host}:{port}`

Basically use container name and port for internal clients (e.g. kafka ui) in docker network, and ip address and port for external clients. If you application in docker it will be enough one listener that serves internal communication.

The most important moment to map your labels to protocol types, using  **KAFKA_LISTENER_SECURITY_PROTOCOL_MAP**, such as:

- PLAINTEXT - insecure communication (for dev or in private network cluster)
- SLL - encrypted and authenticated communication via cert
- SASL_PLAINTEXT - Authenticated, but not encrypted communication
- SASL_SSL - Fully secure and authenticated communication 

`EXTERNAL:PLAINTEXT, INTERNAL:PLAINTEXT`

After using **KAFKA_INTER_BROKER_LISTENER_NAME** we must specify the label of listener used by kafka to communicate between brokers (even if we have only a broker)

#### Kafka UI

```yml
  kafka-ui:
    image: provectuslabs/kafka-ui:latest
    container_name: kafka-ui
    ports:
      - "9080:8080"
    depends_on:
      - kafka
    environment:
      KAFKA_CLUSTERS_0_NAME: local
      KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS: kafka:29092
      KAFKA_CLUSTERS_0_ZOOKEEPER: zookeeper:2181
    networks:
      - default

```

**KAFKA_CLUSTERS_0_NAME** - cluster name shown in web
**KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS** - specify kafka address (internal one cuz we are in docker network)
**KAFKA_CLUSTERS_0_ZOOKEEPER** - specify zookeeper address (internal one)

>[!Note]
> Using container names as host in docker network totally fine, docker network has Internal DNS to resolve them easily
>   

Lets talk overall how kafka works, and why it is not only pub/sub, and should not be used as pub/sub





