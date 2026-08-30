# **🚀 Ads Service — Microservice Platform for Ad Management**

[English](README.md) | [Русский](README.ru.md)

## **📋 Overview**

**Ads Service** is a high-load microservice platform for ad management. The project is built using **Clean Architecture** and **Domain-Driven Design (DDD)** principles with a clear separation into independent, scalable services.

### **✨ Key Features**

✅ **Auth Service v2** — registration, login, logout, refresh sessions, JWT validation, role assignment, account blocking/deletion, email verification and sending verification tokens<br>
✅ **User Service** — user profile management and profile updates<br>
✅ **Ad Service** — full ad lifecycle with create / update / publish / reject / delete flows, moderation and admin-level operations<br>
✅ **Search Service** — ad indexing and search by text, category, price and sorting<br>
✅ **Elasticsearch** — full-text ad search and indexing in dedicated search service<br>
✅ **GraphQL Gateway** — single entry point with gRPC aggregation and access control<br>
✅ **RabbitMQ** — asynchronous event-driven communication between services<br>
✅ **Redis** — verification token storage and auth-related temporary state<br>
✅ **Graceful Shutdown** — clean termination of all services<br>
✅ **Clean Architecture** — layers: handler → usecase → domain → repository<br>
✅ **DDD** — bounded contexts, entities, value objects<br>
✅ **Unit + Integration + E2E Tests** — business logic tests, integration checks with real infrastructure, and full end-to-end service scenarios<br>
✅ **Docker** — containerization for all services<br>
✅ **gRPC** — efficient inter-service communication<br>

---

## 🏗 **System architecture**

```
┌──────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                           Clients (Web/Mobile)                                       │
└────────────────────────────────────────────────────┬─────────────────────────────────────────────────┘
                                                     │ HTTPS
                                                     ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                       GraphQL Gateway (Port: 6060)                                   │
│                                        Aggregation, Authorization                                    │
└───────────┬─────────────────────────┬─────────────────────────┬─────────────────────────┬────────────┘
            │                         │                         │                         │
            │ gRPC                    │ gRPC                    │ gRPC                    │ gRPC
            ▼                         ▼                         ▼                         ▼
┌───────────────────────┐ ┌───────────────────────┐ ┌───────────────────────┐ ┌───────────────────────┐
│     Auth Service      │ │     User Service      │ │       Ad Service      │ │     Search Service    │
│    gRPC Port: 50051   │ │    gRPC Port: 50052   │ │    gRPC Port: 50053   │ │    gRPC Port: 50054   │
│                       │ │                       │ │                       │ │                       │
│   PostgreSQL: auth_db │ │   PostgreSQL: user_db │ │   PostgreSQL: ad_db   │ │   ElasticSearch: ads  │
│         Redis         │ │                       │ │      MongoDB: ads     │ │                       │
└────────────────┬──────┘ └───────────────────────┘ └─────────────┬─────────┘ └───────────────────────┘
                 │                  ▲                             │                  ▲
        RabbitMQ │                  │ RabbitMQ                    │ RabbitMQ         │ RabbitMQ
                 ▼                  │                             ▼                  │
           ┌────────────────────────┴────────┐            ┌──────────────────────────┴───────┐
           │          account_topic          │            │             ad_topic             │
           └─────────────────────────────────┘            └──────────────────────────────────┘
```

---

## **🛠 Tech Stack**

### **Backend**
| Technology   | Purpose                             |
|--------------|-------------------------------------|
| **Go 1.24+** | Primary language                    |
| **gRPC**     | Inter-service communication         |
| **GraphQL**  | API Gateway                         |
| **RabbitMQ** | Async events                        |
| **JWT**      | Authentication and authorization    |
| **Redis**    | Verification tokens and cache       |
| **Elasticsearch** | Full-text search indexing     |

### **Storage**
| Technology     | Service        | Purpose                               |
|----------------|----------------|---------------------------------------|
| **PostgreSQL** | Auth, User, Ad | Main business data                    |
| **MongoDB**    | Ad Service     | Media files, attachments              |
| **Redis**      | Auth Service   | Verification tokens and temporary auth data |
| **Elasticsearch** | Search Service | Ad search index and search data |

### **Infrastructure**
| Technology         | Purpose                    |
|--------------------|----------------------------|
| **Docker**         | Containerization           |
| **Docker Compose** | Local development and orchestration |
| **SMTP**           | Email verification delivery |
| **gRPC Gateway**   | External API aggregation   |
| **sqlc**           | PostgreSQL query generation and typed access |
| **mockery**        | Auto-generated mocks for interfaces and ports |
| **fakes**          | Lightweight in-memory test doubles for service adapters |


## **🚀 Getting Started**

### **Prerequisites**
- Go 1.24+
- Docker & Docker Compose

### **Quick Start**

```bash
# 1. Clone repository
git clone https://github.com/maket12/ads-service.git
cd ads-service

# 2. Configure environment
# The root .env file is the main config for docker-compose and defines service ports and addresses.
# Each microservice also has its own .env.example file for local/service-specific configuration.
# For a standard Docker Compose run, the root .env is the one that matters most.
cp .env.example .env 2>/dev/null || true
cp backend/authservice/.env.example backend/authservice/.env
cp backend/userservice/.env.example backend/userservice/.env
cp backend/adservice/.env.example backend/adservice/.env
cp backend/searchservice/.env.example backend/searchservice/.env
cp backend/gateway/.env.example backend/gateway/.env

# 3. Launch all services (including migrations)
docker compose up --build

# 4. Open GraphQL playground
http://localhost:6060/graphql
```

> The project has a root environment file for Docker Compose that defines the service ports and internal addresses used by the stack. Separate `.env.example` files in each microservice (`authservice`, `userservice`, `adservice`, `searchservice`, `gateway`) are available for local/service-specific configuration and debugging.

> In the standard setup, the root `.env` is the main source of truth for ports and inter-service networking, while service-level env files remain optional helpers for local runs.


---

# 🔌 **API Endpoints**

### **GraphQL Gateway** (port `6060`)

```graphql
# Examples
mutation Login {
    login(input: {
        email: "user@example.com",
        password: "secret123",
        ip: "127.0.0.1",
        userAgent: "Mozilla/5.0"
    }) {
        accessToken
        refreshToken
    }
}

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

mutation CreateAd {
    createAd(input: {
        title: "Apartment for rent",
        description: "2-bedroom apartment in the city center",
        price: 950.50,
        category: "real_estate",
        images: [
            "https://example.com/1.jpg",
            "https://example.com/2.jpg"
        ]
    })
}

query Search {
    search(input: {
        text: "apartment",
        category: "real_estate",
        price_from: 500,
        price_to: 1200,
        limit: 10,
        offset: 0,
        sort_by: "price_desc"
    }) {
        id
        title
        price
        category
        main_image
    }
}
```

### **gRPC Endpoints**

| Service      | Port  | Main methods                                |
|--------------|-------|---------------------------------------------|
| Auth Service | 50051 | `Register`, `Login`, `RefreshSession`, `ValidateAccessToken`, `AssignRole`, `SendVerification`, `VerifyEmail`, `BlockAccount`, `DeleteAccount` |
| User Service | 50052 | `GetProfile`, `UpdateProfile`               |
| Ad Service   | 50053 | `CreateAd`, `GetAd`, `UpdateAd`, `PublishAd`, `RejectAd`, `DeleteAd`, `DeleteAllAds`, `ListAds`, `ListAllAds` |
| Search Service | 50054 | `SearchAds` |

---

## 🧪 **Testing**

The project includes full automated validation across all layers:

### **Unit tests**
- Use case-level tests for business logic
- Repository and mapper coverage
- Mocks via `mockery` for interfaces and ports
- Lightweight `fakes` used for adapters and external dependencies

### **Integration tests**
- Real service interaction checks with infrastructure containers
- PostgreSQL, MongoDB, Redis and RabbitMQ are used in test flows
- Validation of data persistence, messaging and service integration

### **E2E tests**
- Full service flow tests with real infrastructure
- Authentication flows, email verification, role assignment, moderation, and ad lifecycle scenarios
- Search service event-driven validation
- End-to-end checks for the complete backend behavior

### **SQLC**
- PostgreSQL access is generated via `sqlc` for typed queries and DB models
- Used in auth, user and ad services to keep repository code strongly typed and generated from SQL definitions

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
# - search-service:50054
# - gateway:6060
# - auth-db:5431
# - user-db:5432
# - ad-db:5433
# - mongodb:27017
# - auth-redis:6379
# - rabbitmq:5672
# - elasticsearch:9200
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
- `account.created` — user registration event
- `account.deleted` — account removal event
- `ad.published` — ad published to the marketplace
- `ad.updated` — ad content changed
- `ad.rejected` — ad rejected by moderator
- `ad.deleted` — ad deleted

### **Subscriptions**
- User Service reacts to account events
- Search Service consumes ad-related events to index and update Elasticsearch
- RabbitMQ is used as the main async integration layer across services

---

## 📄 **License**

The project is distributed under the Apache-2.0 license.. See the [LICENSE](LICENSE).

---

## ✅ **Status of implementation**

| Component                      | Status             | Note                         |
|--------------------------------|--------------------|------------------------------|
| **Auth Service**               | ✅ Ready            | JWT, refresh, roles, verification, admin actions |
| **User Service**               | ✅ Ready            | Profiles and account integration            |
| **Ad Service**                 | ✅ Ready            | CRUD, moderation, publish/reject flows + MongoDB |
| **Search Service**             | ✅ Ready            | Elasticsearch indexing and search          |
| **GraphQL Gateway**            | ✅ Ready            | Aggregation, authorisation, search integration |
| **RabbitMQ**                   | ✅ Integrated       | Events for account and ad lifecycle         |
| **Redis**                      | ✅ Integrated       | Verification tokens and auth-related temporary storage |
| **Docker**                     | ✅ Containerisation | All services                 |
| **Graceful Shutdown**          | ✅ Realised         | gRPC, DB, queues             |
| **Clean architecture**         | ✅ Realised         | Layers, DDD                  |
| **Testing (unit/integration/e2e)** | ✅ Included   | Full automated validation across all layers |
| **CI/CD**                      | ✅ Ready            | GitHub Actions: linting, unit, integration and e2e tests |
| **Kubernetes**                 | ⏳ In plans         | Helm charts                  |
| **Monitoring**                 | ⏳ In plans         | Prometheus metrics           |

---

**Ready for production!** 🚀

## **📄 License**

This project is licensed under the Apache-2.0 License.