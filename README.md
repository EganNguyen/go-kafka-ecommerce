# E-commerce App — Go + Kafka

A simple, full-stack event-driven e-commerce application demonstrating direct Kafka interaction in Go.

## Architecture

```
┌────────────┐      REST       ┌─────────────┐      Kafka       ┌──────────────┐
│   React    │ ───────────────→│   Go API    │ ───────────────→│    Kafka     │
│  Frontend  │  /api/products  │  (net/http)  │  orders.commands │    Topics    │
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
1. User places an order via the React frontend.
2. The Go API publishes a `PlaceOrder` command to the `orders.commands` Kafka topic.
3. A background consumer reads the command, saves the order to PostgreSQL, and publishes an `OrderPlaced` event to `orders.placed`.
4. A downstream consumer reads `OrderPlaced` and updates the order status to `confirmed`.

## Prerequisites

- Docker & Docker Compose
- Go 1.25+
- Node.js 18+

## Getting Started

### 1. Start Infrastructure

```bash
docker-compose up -d
```

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

The React app starts on `http://localhost:3000`.

## Tech Stack

| Layer       | Technology                          |
|-------------|-------------------------------------|
| Frontend    | React 19, Vite 6, Vanilla CSS       |
| Backend     | Go, net/http                        |
| Messaging   | Kafka (via segmentio/kafka-go)      |
| Database    | PostgreSQL 16                       |
| Infra       | Docker Compose                      |
