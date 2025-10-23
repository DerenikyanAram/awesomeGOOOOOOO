#!/usr/bin/env bash
set -euo pipefail

TOPIC=${KAFKA_TOPIC:-orders}
MSG=$(cat scripts/sample_order.json)

container=kafka
if ! docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
  echo "Kafka container not running (expected 'kafka')" >&2
  exit 1
fi

echo "Sending message to ${TOPIC}…"
docker exec -i ${container} bash -lc "kafka-console-producer.sh --broker-list localhost:9092 --topic ${TOPIC}" <<<"${MSG}"
echo "Done."
