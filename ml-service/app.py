import os, json, threading
from flask import Flask, request, jsonify
from kafka import KafkaConsumer
import redis

app = Flask(__name__)

KAFKA_BOOTSTRAP = os.getenv("KAFKA_BOOTSTRAP", "kafka:9092")
KAFKA_TOPIC = os.getenv("KAFKA_TOPIC", "analytics-events")

REDIS_HOST = os.getenv("REDIS_HOST", "redis")
REDIS_PORT = int(os.getenv("REDIS_PORT", "6379"))

rdb = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)

# Redis keys
K_TOTAL = "stats:events_total"
K_POS = "stats:positive_total"
K_SCORE_SUM = "stats:score_sum"

def record(score: float, label: int):
    # atomic-ish increments
    pipe = rdb.pipeline()
    pipe.incr(K_TOTAL, 1)
    pipe.incrbyfloat(K_SCORE_SUM, float(score))
    if label == 1:
        pipe.incr(K_POS, 1)
    pipe.execute()

def consume_events():
    try:
        consumer = KafkaConsumer(
            KAFKA_TOPIC,
            bootstrap_servers=KAFKA_BOOTSTRAP,
            value_deserializer=lambda m: json.loads(m.decode("utf-8")),
            auto_offset_reset="earliest",
            enable_auto_commit=True,
            group_id="ml-service",
        )
        print(f"[ml-service] consuming from {KAFKA_TOPIC} via {KAFKA_BOOTSTRAP}")
        for msg in consumer:
            event = msg.value
            feats = event.get("features", [])
            score = float(sum(feats)) if isinstance(feats, list) else 0.0
            label = int(score > 50)
            record(score, label)
            print(f"[ml-service] event user={event.get('user_id')} score={score:.2f} label={label}")
    except Exception as e:
        print("[ml-service] Kafka consumer error:", e)

@app.get("/health")
def health():
    return {"status": "ok"}

@app.post("/predict")
def predict():
    data = request.get_json(force=True) or {}
    features = data.get("features", [])
    score = float(sum(features)) if isinstance(features, list) else 0.0
    label = int(score > 50)
    # Also record sync predictions
    record(score, label)
    return jsonify({"score": score, "label": label})

threading.Thread(target=consume_events, daemon=True).start()

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)
