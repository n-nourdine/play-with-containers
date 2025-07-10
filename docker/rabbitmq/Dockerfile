# RabbitMQ with Debian
FROM debian:12-slim

# Install dependencies
RUN apt-get update -y
RUN apt-get install curl -y
RUN apt-get install gnupg -y
RUN apt-get install erlang-nox -y
    # ca-certificates \
    # xz-utils \
    # erlang-nox \
    # erlang-mnesia \
    # erlang-public-key \
    # erlang-crypto \
    # erlang-ssl \
    # erlang-sasl \
    # erlang-asn1 \
    # erlang-inets \
    # erlang-os-mon \
    # erlang-xmerl \
    # erlang-eldap && \
    # rm -rf /var/lib/apt/lists/*

# Create rabbitmq user
RUN groupadd -r rabbitmq --gid=999 && \
    useradd -r -g rabbitmq --uid=999 --home-dir=/var/lib/rabbitmq --shell=/bin/bash rabbitmq

# Set RabbitMQ version
ENV RABBITMQ_VERSION=3.12.10

# Download and install RabbitMQ
RUN curl -fsSL https://github.com/rabbitmq/rabbitmq-server/releases/download/v${RABBITMQ_VERSION}/rabbitmq-server-generic-unix-${RABBITMQ_VERSION}.tar.xz \
    -o rabbitmq-server.tar.xz && \
    tar -xf rabbitmq-server.tar.xz -C /opt && \
    mv /opt/rabbitmq_server-${RABBITMQ_VERSION} /opt/rabbitmq && \
    rm rabbitmq-server.tar.xz

# Set environment variables
ENV RABBITMQ_HOME=/opt/rabbitmq
ENV PATH=$RABBITMQ_HOME/sbin:$PATH
ENV RABBITMQ_LOGS=/var/log/rabbitmq
ENV RABBITMQ_MNESIA_BASE=/var/lib/rabbitmq/mnesia
ENV RABBITMQ_PID_FILE=/var/lib/rabbitmq/rabbitmq.pid

# Create necessary directories
RUN mkdir -p /var/lib/rabbitmq/mnesia /var/log/rabbitmq /etc/rabbitmq && \
    chown -R rabbitmq:rabbitmq /var/lib/rabbitmq /var/log/rabbitmq /etc/rabbitmq /opt/rabbitmq

# Create rabbitmq.conf
RUN echo "loopback_users.guest = false" > /etc/rabbitmq/rabbitmq.conf && \
    echo "listeners.tcp.default = 5672" >> /etc/rabbitmq/rabbitmq.conf && \
    echo "management.tcp.port = 15672" >> /etc/rabbitmq/rabbitmq.conf && \
    chown rabbitmq:rabbitmq /etc/rabbitmq/rabbitmq.conf

# Enable management plugin
USER rabbitmq
RUN rabbitmq-plugins enable rabbitmq_management

# Expose ports
EXPOSE 5672 15672 25672 4369

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD rabbitmq-diagnostics -q ping

# Volume for data persistence
VOLUME ["/var/lib/rabbitmq"]

# Start RabbitMQ
CMD ["rabbitmq-server"]