# 🚀 **Ads Service — Микросервисная платформа для управления объявлениями**

## 📋 **О проекте**

**Ads Service** — это высоконагруженная микросервисная платформа для управления объявлениями, реализованная в соответствии с техническим заданием. Проект построен на принципах **Clean Architecture** и **Domain-Driven Design (DDD)** с четким разделением на независимые, масштабируемые сервисы.

### ✨ **Реализованный функционал**

✅ **Auth Service v2** — регистрация, логин, логаут, refresh-сессии, проверка JWT, назначение ролей, блокировка и удаление аккаунтов, подтверждение email и отправка verification-токенов  
✅ **User Service** — управление профилями пользователей и обновление данных профиля  
✅ **Ad Service** — полный жизненный цикл объявлений: создание, обновление, публикация, отклонение, удаление и админские операции  
✅ **Search Service** — индексирование и поиск объявлений по тексту, категории, цене и сортировке  
✅ **Elasticsearch** — полнотекстовый поиск объявлений в отдельном поисковом сервисе  
✅ **GraphQL Gateway** — единая точка входа с агрегацией gRPC и контролем доступа  
✅ **RabbitMQ** — асинхронное событийное взаимодействие между сервисами  
✅ **Redis** — хранение verification-токенов и временного auth-состояния  
✅ **Graceful Shutdown** — корректное завершение всех сервисов  
✅ **Чистая архитектура** — слои: handler → usecase → domain → repository  
✅ **DDD** — выделенные bounded contexts, entity, value objects  
✅ **Unit + Integration + E2E тесты** — unit-тесты, интеграционные проверки с реальной инфраструктурой и полные end-to-end сценарии  
✅ **Docker** — контейнеризация всех сервисов  
✅ **gRPC** — эффективное межсервисное взаимодействие

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

## 🛠 **Технологический стек**

### **Backend**
| Технология   | Назначение |
|--------------|------------|
| **Go 1.24+** | Основной язык разработки |
| **gRPC**     | Межсервисное взаимодействие |
| **GraphQL**  | API Gateway |
| **RabbitMQ** | Асинхронные события |
| **JWT**      | Аутентификация и авторизация |
| **Redis**    | Verification-токены и временное кэширование |
| **Elasticsearch** | Полнотекстовый поиск объявлений |

### **Хранилища**
| Технология | Сервис | Назначение |
|------------|--------|------------|
| **PostgreSQL** | Auth, User, Ad | Основные данные |
| **MongoDB** | Ad Service | Медиафайлы, вложения |
| **Redis** | Auth Service | Verification-токены и временные auth-данные |
| **Elasticsearch** | Search Service | Индекс поиска объявлений |

### **Инфраструктура**
| Технология | Назначение |
|------------|------------|
| **Docker** | Контейнеризация |
| **Docker Compose** | Локальная разработка и запуск стека |
| **SMTP** | Отправка email-подтверждений |
| **gRPC Gateway** | Агрегация внешнего API |
| **sqlc** | Генерация типизированных SQL-запросов для PostgreSQL |
| **mockery** | Генерация mock-объектов для интерфейсов и портов |
| **fakes** | Лёгкие тестовые заглушки для адаптеров и внешних зависимостей |

---

## 🚀 **Начало работы**

### **Предварительные требования**
- Go 1.24+
- Docker & Docker Compose

### **Быстрый старт**

```bash
# 1. Клонировать репозиторий
git clone https://github.com/maket12/ads-service.git
cd ads-service

# 2. Настроить окружение
# Корневой .env файл является основным конфигом для docker-compose и задаёт порты и адреса сервисов.
# У каждого микросервиса также есть собственный .env.example для локального и сервисного запуска.
# Для стандартного запуска через Docker Compose основной источник данных — это корневой .env.
cp .env.example .env 2>/dev/null || true
cp backend/authservice/.env.example backend/authservice/.env
cp backend/userservice/.env.example backend/userservice/.env
cp backend/adservice/.env.example backend/adservice/.env
cp backend/searchservice/.env.example backend/searchservice/.env
cp backend/gateway/.env.example backend/gateway/.env

# 3. Запустить все сервисы (включая миграции)
docker compose up --build

# 4. Открыть GraphQL playground
http://localhost:6060/graphql
```

> В проекте есть корневой файл окружения для Docker Compose, который задаёт порты и адреса сервисов в стеке. Отдельные `.env.example` файлы в каждом микросервисе (`authservice`, `userservice`, `adservice`, `searchservice`, `gateway`) используются для локальной и сервисной настройки.

> Для стандартного запуска основной источник правды — корневой `.env`, а сервисные env-файлы остаются дополнительными помощниками для локальной отладки и отдельных запусков.


---

## 🔌 **API Endpoints**

### **GraphQL Gateway** (порт `6060`)

```graphql
# Примеры
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
        title: "Квартира в аренду",
        description: "Двухкомнатная квартира в центре города",
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
        text: "квартира",
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

| Сервис | Порт | Основные методы |
|--------|------|-----------------|
| Auth Service | 50051 | `Register`, `Login`, `RefreshSession`, `ValidateAccessToken`, `AssignRole`, `SendVerification`, `VerifyEmail`, `BlockAccount`, `DeleteAccount` |
| User Service | 50052 | `GetProfile`, `UpdateProfile` |
| Ad Service | 50053 | `CreateAd`, `GetAd`, `UpdateAd`, `PublishAd`, `RejectAd`, `DeleteAd`, `DeleteAllAds`, `ListAds`, `ListAllAds` |
| Search Service | 50054 | `SearchAds` |

---

## 🧪 **Тестирование**

Проект включает полноценную автоматизированную проверку на всех уровнях:

### **Unit tests**
- Тесты уровня use case для бизнес-логики
- Покрытие репозиториев и мапперов
- Моки через `mockery` для интерфейсов и портов
- Лёгкие `fakes` используются для адаптеров и внешних зависимостей

### **Integration tests**
- Проверки взаимодействия сервисов с реальной инфраструктурой
- PostgreSQL, MongoDB, Redis и RabbitMQ используются в тестовых сценариях
- Проверяются persistence, messaging и интеграция компонентов

### **E2E tests**
- Полные сценарии работы сервисов
- Проверяются регистрация, логин, подтверждение email, назначение ролей, модерация и жизненный цикл объявлений
- Search service проверяется в сценариях событийной интеграции
- Проходят end-to-end проверки всей backend-логики

### **SQLC**
- Доступ к PostgreSQL генерируется через `sqlc` для типизированных запросов и моделей БД
- Используется в auth, user и ad сервисах для уменьшения ручного кода и повышения типизации

---

## 🐳 **Docker containerisation**

Все сервисы имеют собственные Docker-контейнеры:

```bash
# Собрать и запустить
docker-compose up --build

# Доступные сервисы:
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

Каждый сервис корректно обрабатывает завершение:

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

### **Публикуемые события**
- `account.created` — событие регистрации пользователя
- `account.deleted` — удаление аккаунта
- `ad.published` — публикация объявления
- `ad.updated` — обновление объявления
- `ad.rejected` — отклонение объявления модератором
- `ad.deleted` — удаление объявления

### **Подписки**
- User Service реагирует на события аккаунтов
- Search Service подписан на события объявлений и обновляет Elasticsearch
- RabbitMQ используется как основной слой асинхронной интеграции между сервисами

---

## 📄 **Лицензия**

Проект распространяется под лицензией Apache-2.0. См. файл [LICENSE](LICENSE).

---

## ✅ **Статус реализации**

| Компонент | Статус | Примечание |
|-----------|--------|------------|
| **Auth Service** | ✅ Готов | JWT, refresh, роли, verification, админские действия |
| **User Service** | ✅ Готов | Профили и интеграция с аккаунтами |
| **Ad Service** | ✅ Готов | CRUD, модерация, publish/reject и MongoDB |
| **Search Service** | ✅ Готов | Индексация и поиск в Elasticsearch |
| **GraphQL Gateway** | ✅ Готов | Агрегация, авторизация и интеграция поиска |
| **RabbitMQ** | ✅ Интегрировано | События жизненного цикла аккаунтов и объявлений |
| **Redis** | ✅ Интегрирован | Verification-токены и временное хранение auth-состояния |
| **Docker** | ✅ Контейнеризация | Все сервисы |
| **Graceful Shutdown** | ✅ Реализован | gRPC, БД, очереди |
| **Чистая архитектура** | ✅ Реализована | Слои, DDD |
| **Тесты (unit/integration/e2e)** | ✅ Есть | Полная автоматизированная проверка на всех уровнях |
| **CI/CD** | ✅ Готово | GitHub Actions: линтинг, unit, integration и e2e тесты |
| **Kubernetes** | ⏳ В планах | Helm charts |
| **Мониторинг** | ⏳ В планах | Prometheus metrics |

---

**Готово к продакшену!** 🚀

## **📄 Лицензия**

Этот проект распространяется под лицензией Apache-2.0.