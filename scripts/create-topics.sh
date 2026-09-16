#!/usr/bin/env bash
set -euo pipefail

BROKER="${KAFKA_BROKERS:-localhost:9092}"

for topic in orders trades orderbook ledger settlement notifications; do
  echo "Creating topic: $topic"
  docker exec -it "$(docker ps -qf name=kafka)" kafka-topics.sh \
    --bootstrap-server "$BROKER" \
    --create --if-not-exists \
    --topic "$topic" \
    --partitions 3 \
    --replication-factor 1 || true
done

echo "Done."
