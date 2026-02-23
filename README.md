# E-commerce App — Watermill + Kafka + React

A full-stack event-driven e-commerce application demonstrating the power of [Watermill](https://watermill.io/) with Apache Kafka.

## Architecture

```
┌────────────┐      REST       ┌─────────────┐      Kafka       ┌──────────────┐
│   React    │ ───────────────→│   Go API    │ ───────────────→│  Watermill   │
│  Frontend  │  /api/products  │  (net/http)  │  orders.commands │   Router     │
│  (Vite)    │  /api/orders    │             │                 │              │
└────────────┘                 └─────────────┘                 │  Command →   │
                                      ↑                        │  Event flow  │
                                      │                        └──────────────┘
                                      │                              │
                               ┌──────┴──────┐                      │
                               │ PostgreSQL  │←─────────────────────┘
                               │  (Products, │   DB writes from
                               │   Orders)   │   handlers
                               └─────────────┘
```

**Flow:**
1. User adds products to cart and places an order via the React frontend.
2. The Go API publishes a `PlaceOrder` **command** to the `orders.commands` Kafka topic.
3. Watermill's router consumes the command, processes it (inserts into PostgreSQL, decrements stock), and publishes an `OrderPlaced` **event** to `orders.placed`.
4. A downstream handler consumes `OrderPlaced`, confirms the order, and publishes `OrderConfirmed` to `orders.confirmed`.

## Prerequisites

- Docker & Docker Compose
- Go 1.21+
- Node.js 18+

## Getting Started

### 1. Start Infrastructure

```bash
docker-compose up -d
```

This starts:
- **Kafka** (port 9092)
- **Zookeeper** (port 2181)
- **PostgreSQL** (port 5432)

### 2. Start Backend

```bash
cd backend
go mod tidy
go run .
```

The API server starts on `http://localhost:8080`.

### 3. Start Frontend

```bash
cd frontend
npm install
npm run dev
```

The React app starts on `http://localhost:3000` with API proxy to the Go backend.

## API Endpoints

| Method | Endpoint        | Description           |
|--------|-----------------|-----------------------|
| GET    | /api/products   | List all products     |
| POST   | /api/orders     | Place a new order     |
| GET    | /api/orders     | List recent orders    |

## Tech Stack

| Layer       | Technology                          |
|-------------|-------------------------------------|
| Frontend    | React 19, Vite 6, Vanilla CSS      |
| Backend     | Go, net/http                        |
| Messaging   | Watermill, Kafka (via Sarama)       |
| Database    | PostgreSQL 16                       |
| Infra       | Docker Compose                      |
