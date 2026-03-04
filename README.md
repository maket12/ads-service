# **🚀 Ads Service — Microservice Platform for Ad Management**

[English](README.md) | [Русский](README.ru.md)

## **📋 Overview**

**Ads Service** is a high-load microservice platform for ad management. The project is built using **Clean Architecture** and **Domain-Driven Design (DDD)** principles with a clear separation into independent, scalable services.

### **✨ Key Features**

✅ **Auth Service** — full authentication & authorization (JWT, refresh tokens, roles)
✅ **User Service** — user profile management
✅ **Ad Service** — Ad CRUD with PostgreSQL \+ MongoDB
✅ **GraphQL Gateway** — single entry point with gRPC aggregation
✅ **RabbitMQ** — asynchronous event-driven communication
✅ **Graceful Shutdown** — clean termination of all services
✅ **Clean Architecture** — layers: handler → usecase → domain → repository
✅ **DDD** — bounded contexts, entities, value objects
✅ **Unit \+ Integration Tests** — coverage for key scenarios
✅ **Docker** — containerization for all services
✅ **gRPC** — efficient inter-service communication

---

## 🏗 **System architecture**

```
┌───────────────────────────────────────────────────────────────────────────┐
│                            Clients (Web/Mobile)                           │
└──────────────────────────────────┬────────────────────────────────────────┘
                                   │ HTTPS
                                   ▼
┌───────────────────────────────────────────────────────────────────────────┐
│                     GraphQL Gateway (Port: 8080)                          │
│                      Aggregation, Authorization                           │
└───────────────┬─────────────────┬─────────────────┬───────────────────────┘
                │                 │                 │
                │ gRPC            │ gRPC            │ gRPC
                ▼                 ▼                 ▼
┌───────────────────────┐ ┌───────────────────────┐ ┌───────────────────────┐
│    Auth Service       │ │    User Service       │ │     Ad Service        │
│    gRPC Port: 50051   │ │    gRPC Port: 50052   │ │    gRPC Port: 50053   │
│                       │ │                       │ │                       │
│   PostgreSQL: auth_db │ │   PostgreSQL: user_db │ │   PostgreSQL: ad_db   │
└───────────────────────┘ └───────────────────────┘ └──────────┬────────────┘
                 │                  ▲                          │
        RabbitMQ │                  │ RabbitMQ                 │ MongoDB
                 ▼                  │                          ▼
           ┌─────────────────────────────────┐       ┌────────────────────┐
           │          account_topic          │       │   MongoDB: media   │
           └─────────────────────────────────┘       └────────────────────┘
```

---

## **🛠 Tech Stack**

### **Backend**
| Technology   | Purpose                     |
|--------------|-----------------------------|
| **Go 1.24+** | Primary language            |
| **gRPC**     | Inter-service communication |
| **GraphQL**  | API Gateway                 |
| **RabbitMQ** | Async events                |
| **JWT**      | Authentication              |

### **Storage**
| Technology     | Service        | Purpose                  |
|----------------|----------------|--------------------------|
| **PostgreSQL** | Auth, User, Ad | Main data                |
| **MongoDB**    | Ad Service     | Media files, attachments |

### **Infrastructure**
| Technology         | Purpose           |
|--------------------|-------------------|
| **Docker**         | Containerization  |
| **Docker Compose** | Local development |

## **🚀 Getting Started**

### **Prerequisites**
- Go 1.24+
- Docker & Docker Compose
- Protocol Buffers (protoc)

### **Quick Start**

```bash
# 1. Clone repository
git clone https://github.com/maket12/ads-service.git
cd ads-service

# 2. Copy env
cp .env.example .env

# 3. Launch all services (including migrations)
docker compose up --build

# 4. Open GraphQL playground
http://localhost:8080/graphql
```

---

# 🔌 **API Endpoints**

### **GraphQL Gateway** (порт `8080`)

```graphql
# Examples
query GetProfile {
    me {
        id
        role
        firstName
        lastName
        phone
        avatarUrl
        bio
        updatedAt
    }
}

mutation UpdateProfile {
    updateProfile(
        firstName: "Jane",
        lastName: "Smith",
        phone: "+9876543210",
        avatarUrl: "https://storage.example.com/avatars/new.jpg",
        bio: "Updated bio"
    )
}
```

### **gRPC Endpoints**

| Service      | Port  | Main methods                                |
|--------------|-------|---------------------------------------------|
| Auth Service | 50051 | `ValidateAccessToken`, `Login`, `Register`  |
| User Service | 50052 | `GetProfile`, `UpdateProfile`               |
| Ad Service   | 50053 | `CreateAd`, `UpdateAd`, `DeleteAd`, `GetAd` |

---

## 🧪 **Testing**

Project has full test coverage:

### **Unit tests**
- Mocks via `mockery` for each port
- Isolated testing of use cases
- Coverage: **~85%**

### **Integrational tests**
- Real DB (PostgreSQL, MongoDB)

---

## 🐳 **Docker containerisation**

All services have their own docker containers:

```bash
# Build and launch
docker-compose up --build

# Available services:
# - auth-service:50051
# - user-service:50052
# - ad-service:50053
# - gateway:8080
# - postgres:5432
# - mongodb:27017
# - rabbitmq:5672
```

---

## ⚡ **Graceful Shutdown**

Each service correctly handles termination gracefully: 

```go
// Graceful shutdown
select {
    case <-ctx.Done():
        logger.InfoContext(
            ctx, "received shutdown signal, stopping grpc server...",
        )
        gRPCServer.GracefulStop()
        return nil
    case err := <-errChan:
        return fmt.Errorf("grpc server failed: %w", err)
}
```

---

## 🔄 **RabbitMQ Events**

### **Published events**
- `account.created` — while registration of user

### **Subscriptions**
- User Service subscribed on `account.created`

---

## 📄 **License**

The project is distributed under the Apache-2.0 license.. See the [LICENSE](LICENSE).

---

## ✅ **Status of implementation**

| Component                      | Status             | Note                         |
|--------------------------------|--------------------|------------------------------|
| **Auth Service**               | ✅ Ready            | JWT, refresh, roles          |
| **User Service**               | ✅ Ready            | Profiles, settings           |
| **Ad Service**                 | ✅ Ready            | CRUD + MongoDB               |
| **GraphQL Gateway**            | ✅ Ready            | Aggregation, authorisation   |
| **RabbitMQ**                   | ✅ Integrated       | Events `account.created`     |
| **Docker**                     | ✅ Containerisation | All services                 |
| **Graceful Shutdown**          | ✅ Realised         | gRPC, DB, queues             |
| **Clean architecture**         | ✅ Realised         | Layers, DDD                  |
| **Testing (unit/integration)** | ✅ Included         | Coverage ~80%                |
| **CI/CD**                      | ⏳ In plans         | Linting, tests, installation |
| **Search Service**             | ⏳ In plans         | Elasticsearch                |
| **Kubernetes**                 | ⏳ In plans         | Helm charts                  |
| **Monitoring**                 | ⏳ In plans         | Prometheus metrics           |

---

**Ready for production!** 🚀

## **📄 License**

This project is licensed under the Apache-2.0 License.