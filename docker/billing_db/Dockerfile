FROM alpine:3.21

# Install PostgreSQL and dependencies
RUN apk add --no-cache postgresql postgresql-contrib su-exec bash

# Create data directory and socket directory, set permissions
RUN mkdir -p /var/lib/postgresql/data /run/postgresql && \
    chown -R postgres:postgres /var/lib/postgresql /run/postgresql && \
    chmod 700 /var/lib/postgresql/data

# Copy initialization script
COPY init-postgres.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/init-postgres.sh

# Use the standard Postgres volume location
VOLUME /var/lib/postgresql/data

EXPOSE 5432

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD su-exec postgres pg_isready -U $BILLING_DB_USER -d $BILLING_DB_NAME || exit 1

# Use exec form to properly handle signals
ENTRYPOINT ["/usr/local/bin/init-postgres.sh"]