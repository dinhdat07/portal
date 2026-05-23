# Infrastructure & Deployment Engineering

This document explains the technical configurations of the backing infrastructure powering the Portal Backend, specifically focusing on the data streaming pipeline.

## 🗄️ PostgreSQL Configuration for CDC

To support Change Data Capture (CDC) via Debezium, the primary PostgreSQL instance must be configured for logical replication.

**Docker Compose Configuration:**
```yaml
command:
  - "postgres"
  - "-c"
  - "wal_level=logical"         # Crucial: Enables logical decoding
  - "-c"
  - "max_replication_slots=10"  # Allows Debezium to create a slot
  - "-c"
  - "max_wal_senders=10"
```

- **WHY**: Standard Postgres Write-Ahead Logs (WAL) only contain binary data for crash recovery. `wal_level=logical` instructs Postgres to add row-level information to the WAL, making it readable by replication plugins like `pgoutput`.

## 🐿️ Debezium Kafka Connect

Debezium is deployed as a Kafka Connect connector. It reads the logical WAL and streams mutations to Kafka.

### Connector Configuration (`postgres-cdc-connector.json`)
- **Plugin**: `pgoutput` (Native Postgres logical decoding plugin).
- **Slot Name**: `portal_cdc_slot`.
- **Publication**: `portal_cdc_publication`.
- **Table Include List**: `"public.users, public.action_logs"`. We strictly whitelist tables to avoid spamming Kafka with internal/unnecessary table changes.

**Data Flow**: 
A change in `public.users` will automatically be published to the Kafka topic: `portal.public.users`.

## 📨 Kafka Configuration

Kafka operates in **KRaft Mode** (ZooKeeper-less) for simplicity and performance.

- **Topics Created on Startup**:
  - `notification.requested` (Application domain events via Outbox).
  - `portal.public.users` (CDC events).
  - `portal.public.action_logs` (CDC events).
- **Partitioning**: Currently set to `1` partition per topic in dev. For production, topics should be partitioned based on throughput needs (e.g., partitioning `notification.requested` by `UserID` to ensure ordered delivery per user).

## 🔍 Elasticsearch

Elasticsearch is used as the read-heavy query engine for User searches and Audit Log analytics.
- **Indices**: `portal_users`, `portal_action_logs`.
- The `CDC Consumer` acts as the bridge between Kafka and Elasticsearch, parsing the Debezium payload structure (`before`/`after` states) and mapping it to ES documents.
