# 🚀 Portal Backend: Enterprise Identity & Access Management API

A Backend API service for the Portal system. This project is built using modern **Go** architecture utilizing **gRPC** and **gRPC-Gateway**, featuring **Change Data Capture (CDC)** pipeline via **Kafka, Debezium, and Elasticsearch** for full-text search.

---

## 🌟 Key Features

* **Authentication & Authorization:** Complete JWT-based auth flows (Register, Login, Email Verification, Password Reset).
* **Dynamic RBAC:** Flexible Role-Based Access Control allowing granular custom policies, system roles, and permissions mapping.
* **API Gateway Pattern:** Unified `gRPC-Gateway` translating REST HTTP/JSON requests into high-performance gRPC calls.
* **Full-Text Search & CQRS:** Implements CDC with **Debezium** to stream PostgreSQL row changes into **Kafka**, which are consumed by a background worker and indexed into **Elasticsearch** for low-latency searching of Users and Audit Logs.
* **Security & Reliability:**
  * Distributed rate-limiting via **Redis**.
  * gRPC Interceptors for Auth, RBAC validation, and Input validation (`buf.build/go/protovalidate`).
* **Audit Logging:** Comprehensive tracking of security-relevant system actions.

---

## 📚 Engineering Documentation

The repository contains detailed, production-grade engineering documentation, including architectural deep-dives, sequence diagrams, and infrastructure guides.

* **[Core System Documentation](docs/README.md)**: Entry point for the core monolith architecture and infrastructure.
* **[System Architecture (HLD)](docs/architecture.md)**: Details on CQRS, Outbox Pattern, and Service Boundaries.
* **[Infrastructure Deep Dive](docs/infrastructure.md)**: Guides on PostgreSQL logical replication, Debezium, and Kafka.
* **[Notification Service](services/notification-service/docs/README.md)**: Dedicated docs for the resilient, idempotent notification microservice.
* **[Architecture Diagrams](docs/diagrams/)**: Visual Mermaid flows for Authentication, Outbox Registration, CDC Sync, and Email Processing.

---

## 🛠️ Technology Stack

**Core API:**
* **Language:** Go 1.22+
* **Framework:** gRPC & gRPC-Gateway
* **Contract/Validation:** Protocol Buffers (v3), Buf (`bufbuild/buf`), Protovalidate

**Data & Infrastructure:**
* **Primary Database:** PostgreSQL 15 (with GORM)
* **Caching & Rate Limiting:** Redis 7
* **CDC & Messaging:** Apache Kafka & Debezium (Postgres Connector)
* **Search Engine:** Elasticsearch 8.13 + Kibana

**DevOps & Tooling:**
* Docker & Docker Compose
* MailHog (Local SMTP testing)
* Air (Live reloading)

---

## 📂 Project Structure

```text
portal_backend/
├── cmd/
│   ├── api/main.go               # API Server entry point
│   └── cdc-consumer/main.go      # CDC Consumer entry point
├── proto/                        # Protobuf definitions (Contracts)
├── gen/                          # Auto-generated Go stubs from Protobuf
├── config/                       # Environment and service configuration loaders
├── internal/                     # Private application code
│   ├── app/                      # Application bootstrap, DI, and server assembly
│   ├── domain/                   # Business enums and constants
│   ├── handler/                  # Protocol boundary
│   │   ├── grpcserver/           # gRPC service implementations
│   │   └── gateway/              # HTTP REST gateway mux configuration
│   ├── infrastructure/           # Infrastructure & external adapters
│   │   ├── cdc/                  # Kafka consumer for Change Data Capture
│   │   ├── debezium/             # Debezium connector configs
│   │   ├── elasticsearch/        # ES indexing and querying logic
│   │   ├── kafka/                # Kafka producer/consumer setup
│   │   ├── metrics/              # Prometheus metrics
│   │   ├── ratelimit/            # Redis-based token bucket rate limiter
│   │   └── security/             # JWT Authenticator & RBAC Authorizer Interceptors
│   ├── model/                    # GORM database entities
│   ├── repository/               # Postgres Data Access Layer (GORM)
│   ├── service/                  # Core Business Logic & Orchestration
│   └── worker/                   # Background workers (Outbox Publisher)
├── shared/                       # Shared event schemas
├── services/                     # Satellite microservices
│   └── notification-service/     # Async email notification worker
├── buf.yaml & buf.gen.yaml       # Protobuf generation configurations
└── docker-compose.yml            # Local infrastructure orchestration
```

---

## 🏛️ Architecture & Data Flow

### 1. API Request Flow (REST / gRPC)
```mermaid
graph LR
    A[Client] -->|HTTP/JSON| B(gRPC-Gateway)
    B -->|gRPC| C{Interceptors}
    C -->|Validate, Auth, Limit| D[gRPC Server]
    D --> E[Service Layer]
    E --> F[Repository]
    F --> G[(PostgreSQL)]
```

### 2. CQRS & CDC Search Indexing Flow
To ensure search performance without overloading the primary operational database, we implement a Change Data Capture pipeline:
```mermaid
graph LR
    A[(PostgreSQL)] -->|WAL Logical Replication| B(Debezium Connector)
    B -->|Row level changes| C[Apache Kafka]
    C -->|Consume| D[CDC Consumer Worker]
    D -->|Index Document| E[(Elasticsearch)]
```

---

## 🚀 Getting Started

### Prerequisites
* Go 1.22+
* Docker & Docker Compose
* [Buf CLI](https://buf.build/docs/installation) (For modifying API contracts)

### 1. Environment Setup
Create a `.env` file in the root directory (copy from defaults if available):
```env
DB_URL=postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable
REDIS_ENABLED=true
REDIS_ADDR=localhost:6379
JWT_SECRET=super-secret-key-change-in-production
JWT_ACCESS_TTL=3600
JWT_REFRESH_TTL=604800
HTTP_PORT=8000
GRPC_PORT=50051
ENV=development

# Admin Bootstrap
ADMIN_EMAIL=admin@portal.local
ADMIN_PASSWORD=Admin@123456

# Elasticsearch & Kafka
KAFKA_BROKERS=localhost:9092
ELASTICSEARCH_URL=http://localhost:9200
```

### 2. Spin up Infrastructure
Start PostgreSQL, Redis, Kafka, Debezium, Elasticsearch, and MailHog:
```bash
docker compose up -d
```

### 3. Initialize Debezium CDC Connector
Once Kafka and Debezium are healthy, register the Postgres connector to start listening to WAL changes:
```bash
curl -i -X POST -H "Accept:application/json" -H "Content-Type:application/json" \
  http://localhost:8083/connectors/ -d @internal/infrastructure/debezium/postgres-cdc-connector.json
```

### 4. Run the API Server
Install dependencies and run the main application:
```bash
go mod download
go run ./cmd/api/main.go
```
*(Optional: Use `go run github.com/air-verse/air@latest` for hot-reloading during development).*

---

## 🧪 Testing & Code Generation

**Updating Protobuf Contracts:**
If you make changes to files inside the `proto/` directory, regenerate the Go stubs:
```bash
buf generate
```

**Running Tests:**
Execute unit tests across all packages:
```bash
go test -v ./...
```

---

## 📡 Endpoints Overview

* **REST API Gateway:** `http://localhost:8000/api/v1`
* **gRPC Server:** `localhost:50051`
* **Elasticsearch Search API:** `http://localhost:9200`
* **Kibana UI:** `http://localhost:5601`
* **MailHog UI (Emails):** `http://localhost:8025`

