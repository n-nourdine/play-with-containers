#!/bin/bash
set -e

# Set RabbitMQ bin path for Alpine
export PATH="/usr/lib/rabbitmq/bin:${PATH}"

# Set environment variables with defaults
export RABBITMQ_USER="${RABBITMQ_USER:-nasdev}"
export RABBITMQ_PASSWORD="${RABBITMQ_PASSWORD:-passer}"
export RABBITMQ_QUEUE_NAME="${RABBITMQ_QUEUE_NAME:-billing_queue}"

# Create necessary directories
mkdir -p /var/lib/rabbitmq /var/log/rabbitmq
chown -R rabbitmq:rabbitmq /var/lib/rabbitmq /var/log/rabbitmq

# Set HOME for rabbitmq user (needed for .erlang.cookie)
export HOME=/var/lib/rabbitmq

# Generate and set Erlang cookie BEFORE starting RabbitMQ
ERLANG_COOKIE="mysecretcookie"
echo "$ERLANG_COOKIE" > /var/lib/rabbitmq/.erlang.cookie
chmod 600 /var/lib/rabbitmq/.erlang.cookie
chown rabbitmq:rabbitmq /var/lib/rabbitmq/.erlang.cookie

# Copy cookie to root's home for CLI commands
mkdir -p /root
cp /var/lib/rabbitmq/.erlang.cookie /root/.erlang.cookie
chmod 600 /root/.erlang.cookie

echo "Starting RabbitMQ server..."

# Start RabbitMQ server as rabbitmq user
su rabbitmq -s /bin/bash -c "HOME=/var/lib/rabbitmq rabbitmq-server" &
RABBITMQ_PID=$!

# Wait for RabbitMQ to be ready
echo "Waiting for RabbitMQ to be ready..."
sleep 5  # Initial wait for startup

for i in {1..30}; do
    if HOME=/var/lib/rabbitmq rabbitmq-diagnostics ping >/dev/null 2>&1; then
        echo "RabbitMQ is ready!"
        break
    fi
    echo "Waiting for RabbitMQ... ($i/30)"
    sleep 2
done

# Check if RabbitMQ is running
if ! HOME=/var/lib/rabbitmq rabbitmq-diagnostics ping >/dev/null 2>&1; then
    echo "RabbitMQ failed to start properly"
    exit 1
fi

# Enable management plugin
echo "Enabling management plugin..."
HOME=/var/lib/rabbitmq rabbitmq-plugins enable rabbitmq_management

# Wait for management plugin to start
echo "Waiting for management plugin to start..."
sleep 10

# Setup users
if [ "$RABBITMQ_USER" = "guest" ]; then
    echo "Using default guest user"
    # Ensure guest can connect from any host
    HOME=/var/lib/rabbitmq rabbitmqctl set_permissions -p / guest ".*" ".*" ".*"
else
    echo "Setting up custom user: $RABBITMQ_USER"
    
    # Create new user
    HOME=/var/lib/rabbitmq rabbitmqctl add_user "$RABBITMQ_USER" "$RABBITMQ_PASSWORD" 2>/dev/null || echo "User may already exist"
    HOME=/var/lib/rabbitmq rabbitmqctl set_user_tags "$RABBITMQ_USER" administrator
    HOME=/var/lib/rabbitmq rabbitmqctl set_permissions -p / "$RABBITMQ_USER" ".*" ".*" ".*"
    
    # Remove default guest user for security
    echo "Removing default guest user for security..."
    HOME=/var/lib/rabbitmq rabbitmqctl delete_user guest 2>/dev/null || echo "Guest user may not exist"
fi

# Create the billing queue
echo "Creating queue: $RABBITMQ_QUEUE_NAME"
HOME=/var/lib/rabbitmq rabbitmqctl eval "
rabbit_amqqueue:declare(
    {resource, <<\"/\">>, queue, <<\"$RABBITMQ_QUEUE_NAME\">>},
    true,  % durable
    false, % auto_delete
    [],    % arguments
    none   % acting_user
)." || echo "Queue creation completed"

echo "RabbitMQ setup completed successfully!"
echo "Management UI available at: http://localhost:15672"
echo "Login with: $RABBITMQ_USER / [password]"

# Keep the server running in foreground
wait $RABBITMQ_PID