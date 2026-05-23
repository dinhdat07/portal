# User Registration & Transactional Outbox Flow

This diagram illustrates how the system ensures Eventual Consistency when creating a user and triggering a welcome/verification email, without the risk of dual-write failures.

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant CoreAPI as Portal API
    participant DB as PostgreSQL
    participant Worker as Outbox Worker
    participant Kafka as Kafka

    User->>CoreAPI: POST /api/v1/auth/register
    
    rect rgb(240, 248, 255)
        note right of CoreAPI: Single Database Transaction
        CoreAPI->>DB: BEGIN TX
        CoreAPI->>DB: 1. INSERT public.users
        CoreAPI->>DB: 2. REVOKE old verify tokens (by user + type)
        CoreAPI->>DB: 3. INSERT public.user_tokens (Verify Email Token)
        CoreAPI->>DB: 4. INSERT outbox_events (Topic: notification.requested, Status: Pending)
        CoreAPI->>DB: 5. INSERT public.action_logs (Audit Log)
        DB-->>CoreAPI: COMMIT TX
    end

    CoreAPI-->>User: 200 OK (Registration Successful)

    Note over Worker,Kafka: Asynchronous Outbox Publishing
    
    loop Every N Seconds
        Worker->>DB: SELECT * FROM outbox_events WHERE status = 'pending' FOR UPDATE SKIP LOCKED
        DB-->>Worker: Return Pending Events
        
        alt Has Events
            Worker->>DB: UPDATE status = 'publishing'
            Worker->>Kafka: Publish Event to `notification.requested`
            
            alt Publish Success
                Kafka-->>Worker: ACK
                Worker->>DB: UPDATE status = 'published'
            else Publish Failure (e.g. Timeout)
                Kafka-->>Worker: Error
                Worker->>DB: UPDATE status = 'retry_scheduled' (Exponential Backoff)
            end
        end
    end
```
