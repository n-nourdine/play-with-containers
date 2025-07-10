# Movie Streaming Platform - Microservices Architecture

This project implements a microservices architecture for a movie streaming platform using Docker and Docker Compose. The system consists of an API Gateway that routes requests to two main services: an Inventory API for movie management and a Billing API for payment processing.

## Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│                 │    │                  │    │                 │
│     Client      │────│   API Gateway    │────│  Inventory API  │
│                 │    │   (Port 3000)    │    │   (Port 8080)   │
└─────────────────┘    │                  │    └─────────────────┘
                       │                  │             │
                       │                  │    ┌─────────────────┐
                       │                  │    │                 │
                       │                  │    │ Inventory DB    │
                       │                  │    │ (PostgreSQL)    │
                       └─────────┬────────┘    └─────────────────┘
                                 │
                       ┌─────────▼────────┐    ┌─────────────────┐
                       │                  │    │                 │
                       │    RabbitMQ      │────│   Billing API   │
                       │  (Port 5672)     │    │   (Port 8081)   │
                       └──────────────────┘    └─────────────────┘
                                                        │
                                               ┌─────────────────┐
                                               │                 │
                                               │   Billing DB    │
                                               │  (PostgreSQL)   │
                                               └─────────────────┘
```

## Services

### 1. API Gateway (Port 3000)
- **Purpose**: Routes requests between clients and microservices
- **Technology**: Go with HTTP proxy and RabbitMQ publisher
- **Features**:
  - Proxies `/api/movies/*` requests to Inventory API
  - Sends `/api/billing` requests to RabbitMQ
  - Built-in OpenAPI documentation
  - Request logging and CORS support

### 2. Inventory API (Port 8080)
- **Purpose**: Manages movie inventory with CRUD operations
- **Technology**: Go with PostgreSQL
- **Database**: `movies_db` with `movies` table
- **Endpoints**:
  - `GET /api/movies` - List all movies (supports `?title=` filter)
  - `POST /api/movies` - Create new movie
  - `GET /api/movies/{id}` - Get movie by ID
  - `PUT /api/movies/{id}` - Update movie
  - `DELETE /api/movies/{id}` - Delete movie
  - `DELETE /api/movies` - Delete all movies (requires `Confirm-Delete: yes` header)

### 3. Billing API (Port 8081)
- **Purpose**: Processes billing orders asynchronously via RabbitMQ
- **Technology**: Go with PostgreSQL and RabbitMQ consumer
- **Database**: `billing_db` with `orders` table
- **Features**:
  - Consumes messages from `billing_queue`
  - Processes billing orders in the background
  - Automatic acknowledgment and error handling

### 4. Message Queue (RabbitMQ)
- **Purpose**: Asynchronous message processing for billing
- **Ports**: 5672 (AMQP), 15672 (Management UI)
- **Queue**: `billing_queue`
- **Features**:
  - Persistent messages
  - Automatic queue declaration
  - Management web interface

### 5. Databases
- **Inventory DB**: PostgreSQL on port 5432
- **Billing DB**: PostgreSQL on port 5433
- **Features**:
  - Automatic initialization with required tables
  - Health checks
  - Data persistence via Docker volumes

## Prerequisites

- Docker (version 20.0 or higher)
- Docker Compose (version 2.0 or higher)
- Linux virtual machine (as specified in requirements)

## Project Structure

```
play-with-containers/
├── api-gateway/
│   ├── handlers/
│   │   └── handler.go
│   ├── middleware/
│   │   └── middleware.go
│   ├── rabbitmq/
│   │   └── publisher.go
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
├── inventory-app/
│   ├── database/
│   │   └── database.go
│   ├── handlers/
│   │   └── handler.go
│   ├── model/
│   │   └── model.go
│   ├── util/
│   │   └── utils.go
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
├── billing-app/
│   ├── database/
│   │   └── database.go
│   ├── rabbitmq/
│   │   └── consumer.go
│   ├── util/
│   │   └── utils.go
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
├── docker/
│   ├── inventory_db/
│   │   ├── Dockerfile
│   │   └── init-postgres.sh
│   ├── billing_db/
│   │   ├── Dockerfile
│   │   └── init-postgres.sh
│   └── rabbitmq/
│       ├── Dockerfile
│       ├── init-rabbitmq.sh
│       └── rabbitmq.config
├── docker-compose.yaml
├── .env
└── README.md
```

## Environment Variables

All configuration is managed through the `.env` file:

```bash
# Database Configuration - Inventory
INVENTORY_DB_HOST=inventory-database
INVENTORY_DB_PORT=5432
INVENTORY_DB_USER=nasdev
INVENTORY_DB_PASSWORD=passer
INVENTORY_DB_NAME=movies_db

# Database Configuration - Billing
BILLING_DB_HOST=billing-database
BILLING_DB_PORT=5432
BILLING_DB_USER=billinguser
BILLING_DB_PASSWORD=billingpass
BILLING_DB_NAME=billing_db

# Application Ports
INVENTORY_APP_PORT=8080
BILLING_APP_PORT=8080
API_GATEWAY_PORT=3000

# Service Discovery
INVENTORY_DB_HOST=inventory-app
INVENTORY_DB_PORT=8080

# RabbitMQ Configuration
RABBITMQ_HOST=rabbitmq-queue
RABBITMQ_PORT=5672
RABBITMQ_USER=rabbituser
RABBITMQ_PASSWORD=rabbitpass
RABBITMQ_QUEUE_NAME=billing_queue
```

## Setup and Installation

### 1. Clone the Repository
```bash
git clone <repository-url>
cd play-with-containers
```

### 2. Build and Start Services
```bash
# Build and start all services
docker-compose up --build

# Or run in detached mode
docker-compose up --build -d
```

### 3. Verify Services
```bash
# Check service status
docker-compose ps

# View logs
docker-compose logs -f

# Check specific service logs
docker-compose logs -f api-gateway-app
docker-compose logs -f inventory-app
docker-compose logs -f billing-app
```

## API Usage

### Access Points
- **API Gateway**: http://localhost:3000
- **API Documentation**: http://localhost:3000 (Swagger UI)
- **OpenAPI Spec**: http://localhost:3000/api/docs
- **RabbitMQ Management**: http://localhost:15672 (guest/guest)

### Movie Management Examples

#### 1. Create a Movie
```bash
curl -X POST http://localhost:3000/api/movies \
  -H "Content-Type: application/json" \
  -d '{"title": "Inception", "description": "A mind-bending thriller"}'
```

#### 2. Get All Movies
```bash
curl http://localhost:3000/api/movies
```

#### 3. Search Movies by Title
```bash
curl "http://localhost:3000/api/movies?title=Inception"
```

#### 4. Get Specific Movie
```bash
curl http://localhost:3000/api/movies/{movie-id}
```

#### 5. Update Movie
```bash
curl -X PUT http://localhost:3000/api/movies/{movie-id} \
  -H "Content-Type: application/json" \
  -d '{"title": "Inception", "description": "Updated description"}'
```

#### 6. Delete Movie
```bash
curl -X DELETE http://localhost:3000/api/movies/{movie-id}
```

#### 7. Delete All Movies
```bash
curl -X DELETE http://localhost:3000/api/movies \
  -H "Confirm-Delete: yes"
```

### Billing Examples

#### Process Billing Order
```bash
curl -X POST http://