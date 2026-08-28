# 📋 Technical Specification (TS): Search Service

## 1. Introduction & Purpose

The **Search Service** is a dedicated microservice designed for high-performance full-text search, multi-criteria filtering, sorting, and autocomplete suggestions over the advertisement catalog.

### Core Objectives:
* **High Performance & Low Latency:** Provide instant search and filtering results with minimal overhead on the primary transactional database (`Ad Service`).
* **Event-Driven Denormalization:** Act as a read-heavy consumer of events. The service is **not** the Source of Truth; it ingests data asynchronously from `RabbitMQ` when ads are created, updated, or deleted in `Ad Service`.
* **Single-Hop Feed Delivery:** Store denormalized preview data (IDs, title, price, primary images) directly in Elasticsearch so the `GraphQL Gateway` can retrieve a ready-to-render feed in a single gRPC request without reaching back into `Ad Service`.
* **Clean Architecture & DDD:** Built using layered architecture with strict separation between `domain`, `usecase`, `delivery` (gRPC / RabbitMQ consumer), and `repository` (Elasticsearch).

---

## 2. Tech Stack

| Component | Technology | Description |
| :--- | :--- | :--- |
| **Language** | **Go 1.24+** | Primary programming language |
| **Search Engine** | **Elasticsearch 8.x** | Index storage, full-text search, geo-queries, and aggregations |
| **Inter-service Protocol** | **gRPC / Protocol Buffers v3** | High-speed communication with `GraphQL Gateway` |
| **Message Broker** | **RabbitMQ (AMQP 0.9.1)** | Async event consumer for `AdCreated`, `AdUpdated`, `AdDeleted` |
| **Testing** | **Go `testing`, `stretchr/testify`, `mockery`** | Unit testing for use cases, repository mocking |
| **Containerization** | **Docker & Docker Compose** | Containerized deployment alongside Elasticsearch |

---

## 3. Architecture & Data Flow

### 3.1 Interaction Diagram

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                            GraphQL Gateway                                  │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ gRPC: SearchAds / GetSuggestions
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Search Service                                   │
│  ┌────────────────────────┐  ┌──────────────────────┐  ┌─────────────────┐  │
│  │   gRPC Delivery Layer  │  │  AMQP Consumer Layer │  │  Usecase Layer  │  │
│  └───────────┬────────────┘  └──────────┬───────────┘  └────────┬────────┘  │
└──────────────┼──────────────────────────┼───────────────────────┼───────────┘
               │                          │                       │
               │                          │ Consume Events        │ Query / Index
               │                          ▼                       ▼
       ┌───────┴──────┐         ┌───────────────────┐    ┌─────────────────┐
       │ Client (Web) │         │     RabbitMQ      │    │  Elasticsearch  │
       └──────────────┘         └─────────▲─────────┘    └─────────────────┘
                                          │ Publish Events
                                  ┌───────┴─────────┐
                                  │   Ad Service    │
                                  └─────────────────┘
```

### 3.2 Key Data Flow Pipelines

1. **Write Path (Asynchronous Indexing):**
   * User creates or updates an ad via `Ad Service`.
   * `Ad Service` saves state to PostgreSQL/MongoDB and publishes an event (e.g., `ad.created` / `ad.updated`) to `RabbitMQ`.
   * `Search Service` consumes the message, maps the payload into the `AdIndex` domain model, and writes/updates the document in Elasticsearch.

2. **Read Path (Synchronous Search):**
   * User searches or filters items on the frontend.
   * `GraphQL Gateway` calls `SearchAds` on `Search Service` via gRPC.
   * `Search Service` executes an Elasticsearch query and returns a payload of `SearchAdHit` objects (containing ID, title, price, and image URLs).
   * Upon clicking a specific item, the client fetches the full details directly from `Ad Service` via `GetAd(id)`.

---

## 4. Domain Entities & Data Models

### 4.1 Elasticsearch Document Schema (`AdIndex`)

JSON document indexed under `ads_index`:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "seller_id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Apple MacBook Pro 16 M1 Max",
  "description": "Mint condition, 32GB RAM, 1TB SSD. Comes with full box.",
  "price": 185000,
  "status": "published",
  "images": [
    "https://storage.example.com/ads/mbp16_1.jpg",
    "https://storage.example.com/ads/mbp16_2.jpg"
  ],
  "category_id": "electronics-laptops",
  "location": {
    "lat": 55.7558,
    "lon": 37.6173
  },
  "created_at": "2026-08-25T08:30:00Z",
  "updated_at": "2026-08-25T08:30:00Z"
}
```

### 4.2 Index Mapping Definition (`ads_index_mapping.json`)

```json
{
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "seller_id": { "type": "keyword" },
      "title": { 
        "type": "text", 
        "analyzer": "standard",
        "fields": {
          "suggest": {
            "type": "completion"
          }
        }
      },
      "description": { "type": "text", "analyzer": "standard" },
      "price": { "type": "long" },
      "status": { "type": "keyword" },
      "images": { "type": "keyword", "index": false },
      "category_id": { "type": "keyword" },
      "location": { "type": "geo_point" },
      "created_at": { "type": "date" },
      "updated_at": { "type": "date" }
    }
  }
}
```

---

## 5. gRPC Interfaces (Protobuf Contract)

File: `api/proto/search/v1/search.proto`

```protobuf
syntax = "proto3";

package search.v1;

option go_package = "github.com/maket12/ads-service/backend/adservice/pkg/generated/search_v1;search_v1";

import "google/protobuf/timestamp.proto";

// Service for searching and auto-completing ad entries
service SearchService {
  // Search Ads with full-text search, filters, sorting, and pagination
  rpc SearchAds(SearchAdsRequest) returns (SearchAdsResponse);
  
  // Get search suggestions (autocomplete) based on user input
  rpc GetSuggestions(GetSuggestionsRequest) returns (GetSuggestionsResponse);
}

enum SortOption {
  SORT_OPTION_RELEVANCE_UNSPECIFIED = 0;
  SORT_OPTION_DATE_DESC = 1;
  SORT_OPTION_PRICE_ASC = 2;
  SORT_OPTION_PRICE_DESC = 3;
}

message GeoPoint {
  double latitude = 1;
  double longitude = 2;
}

message GeoFilter {
  GeoPoint center = 1;
  double radius_km = 2;
}

message SearchAdsRequest {
  optional string query = 1;
  optional int64 price_from = 2;
  optional int64 price_to = 3;
  optional string category_id = 4;
  optional GeoFilter geo_filter = 5;
  
  int32 limit = 6;
  int32 offset = 7;
  SortOption sort_by = 8;
}

message SearchAdHit {
  string id = 1;
  string seller_id = 2;
  string title = 3;
  int64 price = 4;
  repeated string images = 5;
  string status = 6;
  google.protobuf.Timestamp created_at = 7;
}

message SearchAdsResponse {
  repeated SearchAdHit ads = 1;
  int64 total = 2;
}

message GetSuggestionsRequest {
  string query = 1;
  int32 limit = 2;
}

message GetSuggestionsResponse {
  repeated string suggestions = 1;
}
```

---

## 6. Use Cases

### UC-1: Index New Advertisement (`IndexAd`)
* **Trigger:** RabbitMQ `ad.created` event.
* **Input Payload:** `AdCreatedEvent` (ID, Title, Description, Price, Status, Images, SellerID, Location).
* **Flow:**
  1. AMQP Consumer parses and validates incoming message.
  2. Passes domain payload to the Usecase layer.
  3. Usecase maps event payload to `AdIndex` entity.
  4. Repository writes document into Elasticsearch via `IndexDocument`.
  5. Message is manually acknowledged (`ACK`).
* **Error Handling:** Invalid messages move to Dead Letter Queue (`DLQ`). If Elasticsearch is temporarily unreachable, message is `NACK`ed with requeue.

### UC-2: Update Advertisement Index (`UpdateAdIndex`)
* **Trigger:** RabbitMQ `ad.updated` or `ad.status_changed` events.
* **Flow:**
  1. Consumer processes event payload.
  2. If status is `published`, updates the corresponding document fields in Elasticsearch.
  3. If status transitions to `rejected`, `draft`, or `deleted`, triggers document deletion from Elasticsearch.

### UC-3: Remove Advertisement from Index (`DeleteAdIndex`)
* **Trigger:** RabbitMQ `ad.deleted` event.
* **Flow:**
  1. Consumer extracts `ad_id`.
  2. Repository invokes `DeleteDocument(ad_id)` in Elasticsearch.

### UC-4: Full-Text Search and Filtering (`SearchAds`)
* **Trigger:** Inbound gRPC `SearchAds` request from `GraphQL Gateway`.
* **Flow:**
  1. Handler validates pagination bounds and arguments.
  2. Usecase constructs Elasticsearch DSL query:
     * Applies `multi_match` on `title^3` and `description` if `query` is present.
     * Enforces `term` filter for `status: "published"`.
     * Applies optional range filters for price and geo-distance queries.
     * Configures sorting and offset/limit pagination.
  3. Elasticsearch returns hits and total count.
  4. Result mapped to `SearchAdsResponse` and returned via gRPC.

### UC-5: Autocomplete Suggestions (`GetSuggestions`)
* **Trigger:** Inbound gRPC `GetSuggestions` request.
* **Flow:**
  1. Queries Elasticsearch `completion` suggester on `title.suggest`.
  2. Returns top N matching terms for auto-complete UI.

---

## 7. Project Layout (Clean Architecture / Go 1.24+)

```text
search-service/
├── api/
│   └── proto/
│       └── search/
│           └── v1/
│               └── search.proto
├── cmd/
│   └── searchservice/
│       └── main.go
├── config/
│   └── config.go
├── internal/
│   ├── delivery/
│   │   ├── grpc/
│   │   │   ├── handler.go
│   │   │   └── mapper.go
│   │   └── amqp/
│   │       ├── consumer.go
│   │       └── event_handler.go
│   ├── domain/
│   │   ├── ad_index.go
│   │   └── errors.go
│   ├── repository/
│   │   └── elasticsearch/
│   │       ├── client.go
│   │       ├── ad_repository.go
│   │       └── mapping.go
│   └── usecase/
│       ├── search_usecase.go
│       └── index_usecase.go
├── pkg/
│   └── generated/
│       └── search_v1/
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

---

## 8. Non-Functional Requirements & Acceptance Criteria

1. **Performance:** `SearchAds` gRPC latency must be ≤ 50 ms (p95) for catalogs up to 1,000,000 documents.
2. **Go Version:** Module defined with `go 1.24+` in `go.mod`.
3. **Reliable Ingestion:** Manual AMQP Acknowledgments (`ACK`/`NACK`) to prevent data loss.
4. **Graceful Shutdown:** Clean disconnection from RabbitMQ, flush in-flight ES requests, and graceful stop for gRPC server upon `SIGINT`/`SIGTERM`.
5. **Test Coverage:** ≥ 80% unit test coverage on Usecase and Delivery layers using mocks.
