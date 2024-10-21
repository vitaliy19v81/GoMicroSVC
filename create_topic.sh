#!/bin/bash

# Убедитесь, что Kafka доступен перед выполнением команды
# Замените kafka:9092 на ваш адрес Kafka, если необходимо
echo "Ожидание запуска Kafka..."
until nc -z kafka 9092; do
  sleep 1
done

# Создаем топик
echo "Создание топика messages_topic..."
kafka-topics.sh --create --bootstrap-server kafka:9092 --replication-factor 1 --partitions 1 --topic messages_topic

echo "Топик messages_topic успешно создан."