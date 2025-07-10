#!/bin/bash

# Test script for Billing API
echo "=== Testing Billing API ==="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
RABBITMQ_HOST="localhost"
RABBITMQ_PORT="5672"
RABBITMQ_USER="admin"
RABBITMQ_PASSWORD="adminpass"
QUEUE_NAME="billing_queue"
BILLING_API_URL="http://localhost:8081"

echo -e "${YELLOW}1. Testing Billing API Health Check${NC}"
curl -s "$BILLING_API_URL/api/health" && echo -e "\n${GREEN}✓ Health check passed${NC}" || echo -e "\n${RED}✗ Health check failed${NC}"

echo -e "\n${YELLOW}2. Checking current orders in database${NC}"
echo "Current orders:"
curl -s "$BILLING_API_URL/api/orders" | jq . 2>/dev/null || echo "No orders found or jq not installed"

echo -e "\n${YELLOW}3. Publishing test messages to RabbitMQ${NC}"

# Test message 1
TEST_MESSAGE_1='{"user_id": "123", "number_of_items": "5", "total_amount": "150.00"}'
echo -e "Publishing message 1: ${GREEN}$TEST_MESSAGE_1${NC}"

# Using rabbitmqadmin (if available) or curl to RabbitMQ management API
if command -v rabbitmqadmin &> /dev/null; then
    echo "$TEST_MESSAGE_1" | rabbitmqadmin publish routing_key="$QUEUE_NAME" payload_encoding=string
else
    # Alternative: Use curl with RabbitMQ Management API
    curl -u "$RABBITMQ_USER:$RABBITMQ_PASSWORD" \
         -H "Content-Type: application/json" \
         -X POST \
         "http://$RABBITMQ_HOST:15672/api/exchanges/%2F/amq.default/publish" \
         -d "{\"properties\":{},\"routing_key\":\"$QUEUE_NAME\",\"payload\":\"$TEST_MESSAGE_1\",\"payload_encoding\":\"string\"}"
fi

# Test message 2
TEST_MESSAGE_2='{"user_id": "456", "number_of_items": "3", "total_amount": "89.99"}'
echo -e "Publishing message 2: ${GREEN}$TEST_MESSAGE_2${NC}"

if command -v rabbitmqadmin &> /dev/null; then
    echo "$TEST_MESSAGE_2" | rabbitmqadmin publish routing_key="$QUEUE_NAME" payload_encoding=string
else
    curl -u "$RABBITMQ_USER:$RABBITMQ_PASSWORD" \
         -H "Content-Type: application/json" \
         -X POST \
         "http://$RABBITMQ_HOST:15672/api/exchanges/%2F/amq.default/publish" \
         -d "{\"properties\":{},\"routing_key\":\"$QUEUE_NAME\",\"payload\":\"$TEST_MESSAGE_2\",\"payload_encoding\":\"string\"}"
fi

# Test message 3
TEST_MESSAGE_3='{"user_id": "789", "number_of_items": "1", "total_amount": "25.50"}'
echo -e "Publishing message 3: ${GREEN}$TEST_MESSAGE_3${NC}"

if command -v rabbitmqadmin &> /dev/null; then
    echo "$TEST_MESSAGE_3" | rabbitmqadmin publish routing_key="$QUEUE_NAME" payload_encoding=string
else
    curl -u "$RABBITMQ_USER:$RABBITMQ_PASSWORD" \
         -H "Content-Type: application/json" \
         -X POST \
         "http://$RABBITMQ_HOST:15672/api/exchanges/%2F/amq.default/publish" \
         -d "{\"properties\":{},\"routing_key\":\"$QUEUE_NAME\",\"payload\":\"$TEST_MESSAGE_3\",\"payload_encoding\":\"string\"}"
fi

echo -e "\n${YELLOW}4. Waiting for messages to be processed...${NC}"
sleep 3

echo -e "\n${YELLOW}5. Checking orders after processing${NC}"
echo "Orders after processing:"
curl -s "$BILLING_API_URL/api/orders" | jq . 2>/dev/null || curl -s "$BILLING_API_URL/api/orders"

echo -e "\n${YELLOW}6. Testing invalid message (should be rejected)${NC}"
INVALID_MESSAGE='{"user_id": "999", "invalid_field": "test"}'
echo -e "Publishing invalid message: ${RED}$INVALID_MESSAGE${NC}"

if command -v rabbitmqadmin &> /dev/null; then
    echo "$INVALID_MESSAGE" | rabbitmqadmin publish routing_key="$QUEUE_NAME" payload_encoding=string
else
    curl -u "$RABBITMQ_USER:$RABBITMQ_PASSWORD" \
         -H "Content-Type: application/json" \
         -X POST \
         "http://$RABBITMQ_HOST:15672/api/exchanges/%2F/amq.default/publish" \
         -d "{\"properties\":{},\"routing_key\":\"$QUEUE_NAME\",\"payload\":\"$INVALID_MESSAGE\",\"payload_encoding\":\"string\"}"
fi

echo -e "\n${GREEN}=== Test completed ===${NC}"
echo -e "Check the billing-app logs with: ${YELLOW}docker logs billing-app${NC}"
echo -e "Access RabbitMQ Management UI at: ${YELLOW}http://localhost:15672${NC}"
echo -e "Username: admin, Password: adminpass"