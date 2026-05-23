# System Architecture (High-Level Design)

## Overview
The Portal Backend is designed around a **Monolithic Core with Event-Driven Satellite Services**. It applies principles of **CQRS** (Command Query Responsibility Segregation) and **Eventual Consistency** to achieve high availability and loose coupling.

## 🏗 Core Modules & Service Boundaries

### 1. Portal Core API (The Monolith)
The primary entry point for all client commands. 
- **Responsibilities**: Authentication, Authorization, User Management, Admin Operations, and Audit Logging.
- **Data Store**: PostgreSQL (Primary Source of Truth).
- **Integration**: Emits integration events using the **Transactional Outbox Pattern** to ensure absolute consistency between business state and messaging state.

### 2. Notification Service
A completely decoupled microservice handling asynchronous communication with users.
- **Responsibilities**: Sending emails (Welcome, OTP, Password Reset).
- **Mechanics**: Consumes events from Kafka. It is designed to be highly resilient, featuring an internal database to track delivery states (`NotificationDelivery`), enabling **Idempotent** processing and independent **Retry** loops.
- **Docs**: [Notification Service Architecture](../services/notification-service/docs/architecture.md)

### 3. CDC Consumer (Search Indexer)
A background worker acting as the "Query" side of our CQRS implementation.
- **Responsibilities**: Keeping the Elasticsearch index synchronized with the primary PostgreSQL database.
- **Mechanics**: Consumes row-level database changes emitted by Debezium via Kafka.

---

## ⚙️ Key Architecture Patterns

### 1. Transactional Outbox Pattern
Used for publishing business events (e.g., `NotificationRequested`).
- **WHY**: Without an outbox, a system might save a user to the database but crash before sending the Kafka message (or vice versa). 
- **HOW**: The Core Service wraps the business entity insertion and the `outbox_events` insertion in a **single PostgreSQL transaction**. A background `Outbox Worker` polls this table and publishes to Kafka.

### 2. Log-Based Change Data Capture (CDC)
Used for replicating state to Elasticsearch for advanced querying.
- **WHY**: Dual-writing to both Postgres and Elasticsearch from the API layer is fragile and leads to inconsistencies if one fails.
- **HOW**: The API only writes to Postgres. **Debezium** attaches to the Postgres Write-Ahead Log (WAL), detects row changes (`public.users`, `public.action_logs`), and streams them to Kafka. The CDC Consumer reads these streams and updates Elasticsearch.

### 3. Idempotent Consumers
- **WHY**: Kafka guarantees "At-Least-Once" delivery. In failure scenarios, messages may be redelivered.
- **HOW**: Consumers (like the Notification Service) use database constraints (e.g., unique `EventID`) to detect duplicate messages. If a duplicate is detected, the operation is skipped, but the Kafka offset is still committed to move the partition forward.

---

## 🔒 Authentication Flow
The system uses a robust JWT-based authentication mechanism.
1. **Access Tokens**: Short-lived JWTs containing user claims and roles. Signed symmetrically/asymmetrically.
2. **Refresh Tokens**: Opaque, long-lived strings stored in the database (`refresh_tokens` table) alongside an `auth_sessions` record.
3. **Security**: Storing refresh tokens enables forced session revocation. The system also detects **Refresh Token Reuse** (a sign of token theft) and automatically invalidates the entire session tree if detected.

---

## 📚 Diagrams
Refer to the `diagrams/` folder for visual representations of these architectures:
- [System Architecture](diagrams/system_architecture.md)
- [User Registration (Outbox Flow)](diagrams/user_registration_outbox_flow.md)
- [CDC Search Sync Flow](diagrams/cdc_search_sync_flow.md)
