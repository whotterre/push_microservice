# 📨 Push Notification Microservice

### Part of: Stage 4 – Distributed Notification System (HNG 13 Backend Track)

This microservice is responsible for sending **mobile and web push notifications** asynchronously via a message queue, or synchronously through its REST API.
It works as part of the distributed notification architecture that also includes the API Gateway, User, Email, and Template services.

---

## 🚀 Overview

The **Push Service** listens to the `push.send.queue` and `push.tokens.queue` from RabbitMQ and handles all outgoing push notifications.
It integrates with **OneSignal** for multi-platform push notification delivery (web, iOS, Android).

Core workflow:

1. **Asynchronous**: The **API Gateway** routes notification requests to `push.send.queue` and device registrations to `push.tokens.queue`.
2. The **Push Service** consumes those messages and validates the OneSignal Player IDs.
3. Sends the message to OneSignal API.
4. Logs status, updates retry counts, and handles failures gracefully.
5. **Synchronous**: Direct REST API endpoints available for immediate push notification sending with instant feedback.

---

## 🧩 Responsibilities

* Consume messages from RabbitMQ (`push.send.queue` and `push.tokens.queue`)
* Register and manage user devices (map User IDs to OneSignal Player IDs)
* Send push notifications via OneSignal REST API
* Support both synchronous (REST API) and asynchronous (RabbitMQ) delivery modes
* Provide health monitoring and dependency status
* Handle device registration, updates, and deactivation

---

## ⚙️ Architecture

```
                ┌─────────────────────────────┐
                │     API Gateway Service     │
                │ (routes push notifications) │
                └─────────────┬───────────────┘
                              │
                     notifications.direct
                              │
                        ┌─────┴──────┐
                        │ push.queue │
                        └─────┬──────┘
                              │
                    ┌─────────▼──────────┐
                    │   Push Service     │
                    │ (this repository)  │
                    ├────────────────────┤
                    │ Validate tokens    │
                    │ Send via FCM/OneS  │
                    │ Retry + DLQ logic  │
                    │ Track delivery     │
                    └─────────┬──────────┘
                              │
                          ┌───▼───┐
                          │ Users │
                          │  Apps │
                          └───────┘
```

---

## 🛠️ Tech Stack

* **Language:** Go 
* **Message Queue:** RabbitMQ
* **Database:** PostgreSQL (for tokens, message logs)
* **Cache:** Redis (for rate limiting, retry tracking)
* **Containerization:** Docker
* **Docs:** OpenAPI (Swagger)

---

## 🧠 Message Queue Setup

| Queue              | Purpose                                           |
| ------------------ | ------------------------------------------------- |
| `push.send.queue`  | Push notification delivery requests               |
| `push.tokens.queue`| Device registration and token updates             |

---

## 💬 REST API Endpoints

Base URL:

```
http://localhost:8003/api
```

### **1. Send Push Notification (Synchronous)**

**POST** `/push/send`

Send a push notification immediately to a user (bypasses queue).
Looks up the user's registered devices and sends via OneSignal.

#### Request Body

```json
{
  "user_id": "user123",
  "title": "Order Update",
  "message": "Your order has been shipped",
  "data": {
    "order_id": "234",
    "deep_link": "myapp://orders/234"
  }
}
```

#### Response

```json
{
  "success": true,
  "message": "Notification sent successfully",
  "onesignal_id": "c28b2fce-27a3-4e93-8c3e-98b1d1a2fd8f",
  "recipients": 2,
  "devices_found": 2
}
```

---

### **2. Register/Update Device**

**POST** `/push/register`

Register or update a user's device with their OneSignal Player ID.

#### Request Body

```json
{
  "user_id": "user123",
  "onesignal_player_id": "abc123-def456-ghi789",
  "platform": "web"
}
```

#### Response

```json
{
  "success": true,
  "message": "Device registered successfully"
}
```

---

### **3. Health Check**

**GET** `/health`

Checks the service’s dependencies and uptime.

#### Example Response

```json
{
  "status": "healthy",
  "timestamp": "2025-11-10T12:10:00Z",
  "service": "Push Notifications Service",
  "dependencies": {
    "rabbitmq": "connected",
    "postgresql": "connected"
  }
}
```

---

## 🧪 Testing

A test page is available for browser-based OneSignal subscription testing:

**Static Test Page**: `http://localhost:4000/static/test-subscriber.html`

**Note**: The `static/` folder contains test files with the OneSignal App ID embedded. These files should **not be committed to version control**. Add `static/` to `.gitignore` to prevent accidentally pushing secrets.

### Test Endpoints

**POST** `/test/push` - Send a test notification to specific Player IDs
**GET** `/test/players` - Retrieve all registered OneSignal players

See `TEST_GUIDE.md` for detailed testing instructions.

---

## 📂 Example Folder Structure

```
push-service/
├── cmd/
│   └── main.go
├── internal/
|   |── dto/
│   ├── handlers/
## 🧰 Environment Variables

| Variable              | Description                                  |
| --------------------- | -------------------------------------------- |
| `RABBITMQ_URL`        | RabbitMQ connection string                   |
| `ONESIGNAL_APP_ID`    | OneSignal App ID                             |
| `ONESIGNAL_REST_KEY`  | OneSignal REST API key                       |
| `DB_HOST`             | PostgreSQL host                              |
| `DB_PORT`             | PostgreSQL port (default: 5432)              |
| `DB_USER`             | PostgreSQL username                          |
| `DB_PASSWORD`         | PostgreSQL password                          |
| `DB_NAME`             | PostgreSQL database name                     |
| `PORT`                | Service port (default: 4000)                 |
└── README.md
```

---

## 🧰 Environment Variables

| Variable        | Description                      |
| --------------- | -------------------------------- |
| `RABBITMQ_URL`  | RabbitMQ connection string       |
| `FCM_API_KEY`   | Firebase API key                 |
| `ONESIGNAL_KEY` | OneSignal REST key               |
| `POSTGRES_URL`  | Connection string for PostgreSQL |
| `REDIS_URL`     | Redis URL (optional)             |
| `PORT`          | Service port (default: 8003)     |
| `SERVICE_NAME`  | Name for service discovery       |

---

## �️ Database Schema

### `user_devices` Table

Stores the mapping between User IDs and OneSignal Player IDs:

| Column       | Type      | Description                              |
| ------------ | --------- | ---------------------------------------- |
| `id`         | UUID      | Primary key                              |
| `user_id`    | String    | User identifier (indexed)                |
| `player_id`  | String    | OneSignal Player ID (unique)             |
| `platform`   | String    | Platform: web, ios, android              |
| `is_active`  | Boolean   | Whether device is active                 |
| `created_at` | Timestamp | Device registration time                 |
| `updated_at` | Timestamp | Last update time                         |

---

## 🔁 Error Handling & DLQ

* **Synchronous Mode**: Returns immediate error response to caller with details
* **Asynchronous Mode**: Failed messages should be routed to Dead Letter Queue (DLQ) for retry/inspection (pending implementation)

---

## 🔄 CI/CD Workflow

GitHub Actions workflow:

* Runs tests and linters on every push
* Builds and pushes Docker image
* Deploys to assigned server (`/request-server`)
* Restarts container via `docker-compose` on deployment

---

## 📊 Monitoring & Logs

Metrics tracked:

* Message throughput (msg/min)
* Queue depth
* Success vs failure rate
* External API latency

Logs include correlation IDs for tracing each notification end-to-end:

```
[prefix:push-service] [notification_id:c28b2fce] status=sent provider=fcm
```

---

## 🧾 Response Format (Standard)

```json
{
  "success": true,
  "data": {},
  "message": "operation successful",
  "meta": {
    "total": 1,
    "limit": 10,
    "page": 1,
    "total_pages": 1,
    "has_next": false,
    "has_previous": false
  }
}
```

---

## 🧠 Learning Outcomes

This service demonstrates:

* Event-driven push notification delivery
* Asynchronous message processing
* Circuit breaker and retry mechanisms
* Microservice isolation and scalability
* Real-world CI/CD setup with monitoring

---

## 👥 Maintainers

Push Notification Service – **Group 8**
HNG13 Backend Cohort — *Stage 4: Microservices & Message Queues*
