# Billing Service - RabbitMQ Consumer

## Overview

The Billing Service is a Go-based microservice that consumes messages from a RabbitMQ queue and stores order information in a PostgreSQL database. It's designed to process billing orders asynchronously through message queuing.

## Architecture

```
RabbitMQ (billing_queue) → Billing App → PostgreSQL (billing_db)
```

## Features

- **RabbitMQ Consumer**: Consumes messages from `billing_queue`
- **PostgreSQL Integration**: Stores order data in `billing_db` database
- **Message Acknowledgment**: Properly acknowledges processed messages
- **Error Handling**: Rejects malformed messages, retries on database errors
- **Health Checks**: HTTP endpoint for service monitoring
- **Graceful Shutdown**: Proper cleanup on termination signals

## Database Schema

The `orders` table structure:

```sql
CREATE TABLE orders (
    id TEXT PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    number_of_items VARCHAR(255) NOT NULL,
    total_amount VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Message Format

The service expects JSON messages in this format:

```json
{
  "user_id": "123",
  "number_of_items": "5", 
  "total_amount": "150.00"
}
```

## Environment Variables

Create a `.env` file with the following variables:

```bash
# Billing Database
BILLING_DB_HOST=billing-database
BILLING_DB_PORT=5432
BILLING_DB_USER=billinguser
BILLING_DB_PASSWORD=billingpass
BILLING_DB_NAME=billing_db

# RabbitMQ
RABBITMQ_HOST=rabbitmq-queue
RABBITMQ_PORT=5672
RABBITMQ_USER=admin
RABBITMQ_PASSWORD=adminpass
RABBITMQ_QUEUE_NAME=billing_queue

# Application
BILLING_APP_PORT=8080
```

## Quick Start

### 1. Build and Run with Docker Compose

```bash
# Build and start all services
docker-compose up --build

# Start in background
docker-compose up -d --build
```

### 2. Test the Service

```bash
# Check service health
curl http://localhost:8080/api/health

# View current orders
curl http://localhost:8080/api/orders
```

### 3. Send Test Messages

**Option A: Using the test script**
```bash
chmod +x test-billing.sh
./test-billing.sh
```

**Option B: Using RabbitMQ Management UI**
1. Open http://localhost:15672
2. Login with `admin` / `adminpass`
3. Go to Queues → `billing_queue`
4. Publish a test message

**Option C: Using the Go publisher**
```bash
# Build the publisher
cd publisher
go mod init publisher
go get github.com/streadway/amqp
go run publisher.go 123 5 150.00
```

## Service Endpoints

- **Health Check**: `GET /api/health`
- **View Orders**: `GET /api/orders` (debug endpoint)

## Testing Scenarios

### 1. Normal Operation
- Start billing service
- Send messages to queue
- Verify orders appear in database

### 2. Service Resilience
- Stop billing service
- Send messages to queue
- Start billing service
- Verify queued messages are processed

### 3. Invalid Messages
- Send malformed JSON
- Send messages with missing fields
- Verify messages are rejected (not requeued)

## Monitoring

### Check Logs
```bash
# View billing app logs
docker logs billing-app

# Follow logs in real-time
docker logs -f billing-app
```

### RabbitMQ Management
- URL: http://localhost:15672
- Username: `admin`
- Password: `adminpass`

### Database Access
```bash
# Connect to billing database
docker exec -it billing-database psql -U billinguser -d billing_db

# View orders
SELECT * FROM orders ORDER BY created_at DESC;
```

## Project Structure

```
billing-app/
├── main.go                 # Application entry point
├── model/
│   └── order.go           # Order data model
├── database/
│   └── database.go        # Database operations
├── rabbitmq/
│   └── consumer.go        # RabbitMQ consumer
├── util/
│   └── utils.go           # Utility functions
├── Dockerfile             # Container build instructions
├── go.mod                 # Go dependencies
└── go.sum                 # Go checksums
```

## Dependencies

- **github.com/jackc/pgx/v5**: PostgreSQL driver
- **github.com/streadway/amqp**: RabbitMQ client
- **github.com/google/uuid**: UUID generation

## Troubleshooting

### Common Issues

1. **RabbitMQ Connection Failed**
   - Check if RabbitMQ container is running
   - Verify credentials in `.env` file
   - Check network connectivity

2. **Database Connection Failed**
   - Ensure billing-db container is healthy
   - Verify database credentials
   - Check if database exists

3. **Messages Not Processing**
   - Check billing-app logs
   - Verify queue name matches
   - Ensure message format is correct

### Debug Commands

```bash
# Check container status
docker-compose ps

# View all logs
docker-compose logs

# Restart specific service
docker-compose restart billing-app

# Clean restart
docker-compose down
docker-compose up --build
```

## Performance Considerations

- **QoS Setting**: Service processes one message at a time for reliability
- **Connection Management**: Automatic reconnection on failures
- **Transaction Safety**: Database operations use transactions
- **Memory Usage**: Minimal memory footprint with Alpine Linux base

## Security Notes

- Database credentials should be in `.env` file (not committed to repo)
- RabbitMQ uses authentication
- No sensitive data in Docker images
- Health endpoints don't expose sensitive information