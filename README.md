\# Cloud-Native AI Analytics Platform (Demo)



A demo-ready cloud-native analytics platform built with \*\*microservices\*\*:

\- \*\*Producer (Python)\*\* publishes events to \*\*Kafka\*\*

\- \*\*ML Service (Python/Flask)\*\* consumes events and runs inference

\- \*\*API Gateway (Go)\*\* exposes REST endpoints

\- \*\*Redis\*\* stores real-time analytics counters

\- Everything runs via \*\*Docker Compose\*\* (local), ready for \*\*Kubernetes + CI/CD\*\*



\## Architecture



+------------+ +--------+ +----------------+ +--------+

| Producer | ---> | Kafka | ---> | ML Service | ---> | Redis |

| (Python) | | Broker | | (Flask + Kafka) | | stats |

+------------+ +--------+ +----------------+ +--------+

^

|

+----------------+

| API Gateway |

| (Go) |

+----------------+

/predict /stats





\## Services

\- `producer/` (Python): sends events to Kafka topic `analytics-events`

\- `ml-service/` (Python Flask): consumes events, computes score/label, stores counters in Redis

\- `api-gateway/` (Go): `/predict` forwards to ml-service; `/stats` reads Redis counters

\- `docker/docker-compose.yml`: runs Zookeeper, Kafka, Redis, and services



\## How to run (Local)



\### Prerequisites

\- Docker Desktop

\- Docker Compose v2



\### Start

From repo root:



```bash

docker compose -f docker/docker-compose.yml up --build



