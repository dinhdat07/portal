# Portal Backend - Core System Documentation

Welcome to the internal engineering documentation for the Portal Backend. This repository houses a **Monolithic Core with Event-Driven Satellite Services** architecture written in **Go**, designed for high scalability, eventual consistency, and robustness.

## 📚 Documentation Index

1. **[System Architecture](architecture.md)**: High-Level Design (HLD), domains, boundaries, and macro-level patterns.
2. **[Infrastructure & Deployment](infrastructure.md)**: Details on PostgreSQL logical replication, Debezium, Kafka, and Elasticsearch.
3. **[Notification Service Docs](../services/notification-service/docs/README.md)**: Dedicated documentation for the async notification worker.
4. **[Diagrams Library](diagrams/)**: Visual representation of core workflows (Mermaid).

---

## 🚀 Quick Start (Development)

### Prerequisites
- Docker & Docker Compose
- Go 1.22+ (if running locally outside Docker)

### Running the Infrastructure
The system relies heavily on backing services. To spin up PostgreSQL, Redis, Kafka, Elasticsearch, and Mailhog:

```bash
# Start all backing services (DB, Kafka, Elastic, etc.)
docker compose up -d db redis kafka kafka-init kafka-connect elasticsearch kibana mailhog
```

### Environment Variables
Ensure you have a `.env` file at the root. Key variables required:
```ini
DB_URL=postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable
KAFKA_BROKERS=localhost:9092
ELASTICSEARCH_URL=http://localhost:9200
```

### Running the Core Services
Once infrastructure is healthy, you can run the core components:

```bash
# 1. Run the Portal Core API (gRPC/REST Gateway)
go run ./cmd/api

# 2. Run the CDC Indexing Consumer
go run ./cmd/cdc-consumer

# 3. Run the Notification Service (from its own directory)
cd services/notification-service && go run ./cmd
```

## 🛠 Project Structure

- `cmd/`: Application entrypoints (`api`, `cdc-consumer`).
- `internal/`: Core monolith logic.
  - `app/`: Dependency injection and bootstrap.
  - `domain/`: Business enums and constants.
  - `handler/`: gRPC and REST Gateway controllers.
  - `infrastructure/`: Adapters for external systems (Kafka, CDC, ES, DB).
  - `repository/`: Data access layer.
  - `service/`: Core business logic (Auth, User, Admin).
  - `worker/`: Background workers (e.g., Outbox Publisher).
- `services/`: Microservices extracted from the core (e.g., `notification-service`).
- `docs/`: Engineering documentation.
