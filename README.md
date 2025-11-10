# 📨 Push Notification Microservice

### Part of: Stage 4 – Distributed Notification System (HNG 13 Backend Track)

This microservice is responsible for sending **mobile and web push notifications** asynchronously via a message queue, or synchronously through its REST API.
It works as part of the distributed notification architecture that also includes the API Gateway, User, Email, and Template services.

---

## 🚀 Overview

The **Push Service** listens to the `push.queue` from `notifications.direct` exchange and handles all outgoing push notifications.
It integrates with push providers like **Firebase Cloud Messaging (FCM)**, **OneSignal**, or **Web Push (VAPID)**.

Core workflow:

1. The **API Gateway** routes notification requests to `push.queue`.
2. The **Push Service** consumes those messages and validates the push tokens.
3. Sends the message to the correct provider (FCM, OneSignal, etc.).
4. Logs status, updates retry counts, and handles failures gracefully.

---

## 🧩 Responsibilities

* Consume messages from RabbitMQ (`push.queue`)
* Validate and store device tokens
* Send notifications with title, body, image, and deep link
* Retry failed deliveries with exponential backoff
* Expose synchronous API endpoints for direct sends and status checks
* Provide health monitoring and dependency status

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

| Exchange               | Queue          | Routing Key | Purpose                            |
| ---------------------- | -------------- | ----------- | ---------------------------------- |
| `notifications.direct` | `push.queue`   | `push`      | Push notification delivery         |
| `notifications.direct` | `failed.queue` | `failed`    | Stores permanently failed messages |

---

## 💬 REST API Endpoints

Base URL:

```
http://localhost:8003/api
```

### **1. Send Push Notification (Direct)**

**POST** `/push/send`

Send a push notification immediately (bypasses queue).
Useful for admin panels or testing.

#### Request Body

```json
{
  "push_token": "device_fcm_token",
  "title": "Order Update",
  "body": "Your order has been shipped",
  "image_url": "https://cdn.example.com/order.png",
  "deep_link": "myapp://orders/234",
  "platform": "android",
  "data": {
    "order_id": "234"
  }
}
```

#### Response

```json
{
  "success": true,
  "data": {
    "message_id": "c28b2fce-27a3-4e93-8c3e-98b1d1a2fd8f",
    "status": "sent",
    "timestamp": "2025-11-10T12:03:00Z"
  },
  "message": "Push notification sent successfully"
}
```

---

### **2. Get Notification Status**

**GET** `/push/status/{message_id}`

Retrieve the current delivery status of a specific push notification.

#### Example Response

```json
{
  "success": true,
  "data": {
    "message_id": "c28b2fce-27a3-4e93-8c3e-98b1d1a2fd8f",
    "status": "delivered",
    "recipient_token": "device_fcm_token",
    "platform": "android",
    "sent_at": "2025-11-10T12:03:00Z",
    "delivery_info": {
      "fcm_message_id": "0:1731234567890%abc12345",
      "retry_count": 0
    }
  },
  "message": "Status retrieved"
}
```

---

### **3. Update User Push Token**

**PUT** `/push/tokens/{user_id}`

Register or update a user's push token (FCM, OneSignal, or web push key).

#### Request Body

```json
{
  "push_token": "new_fcm_token",
  "platform": "android",
  "app_version": "3.1.0"
}
```

#### Response

```json
{
  "message": "Token updated successfully"
}
```

---

### **4. Health Check**

**GET** `/health`

Checks the service’s dependencies and uptime.

#### Example Response

```json
{
  "status": "healthy",
  "timestamp": "2025-11-10T12:10:00Z",
  "service": "push-service",
  "dependencies": {
    "rabbitmq": "connected",
    "fcm": "connected",
    "template_service": "connected"
  }
}
```

---

## 📂 Example Folder Structure

```
push-service/
├── cmd/
│   └── main.go
├── internal/
|   |── dto/
│   ├── handlers/
    |── initializers/
│   ├── services/
│   ├── queue/
│   ├── repository/
│   └── config/
├── pkg/
│   └── utils/
├── specs/
├── Dockerfile
├── docker-compose.yml
├── .env.example
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

## 🔁 Retry, Circuit Breaker & DLQ

* **Retries**: Uses exponential backoff for failed deliveries.
* **Circuit Breaker**: Temporarily halts external calls if FCM or OneSignal repeatedly fail.
* **Dead Letter Queue**: Permanently failed messages are moved to `failed.queue` for later inspection.

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
