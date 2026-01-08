import os, json, time, random
from kafka import KafkaProducer

bootstrap = os.getenv("KAFKA_BOOTSTRAP", "kafka:9092")
topic = os.getenv("KAFKA_TOPIC", "analytics-events")

producer = KafkaProducer(
    bootstrap_servers=bootstrap,
    value_serializer=lambda v: json.dumps(v).encode("utf-8"),
)

print(f"[producer] sending events to {topic} via {bootstrap}")

while True:
    event = {
        "user_id": random.randint(1, 1000),
        "features": [random.random() * 30, random.random() * 40, random.random() * 50],
        "ts": time.time(),
    }
    producer.send(topic, event)
    producer.flush()
    print("[producer] sent:", event)
    time.sleep(1)
