#!/bin/bash
set -e

export PGDATA="/var/lib/postgresql/data"
export DB_NAME="${INVENTORY_DB_NAME:-movies_db}"
export DB_USER="${INVENTORY_DB_USER:-nasdev}"
export DB_PASSWORD="${INVENTORY_DB_PASSWORD:-passer}"

# Create socket directory
mkdir -p /run/postgresql
chown postgres:postgres /run/postgresql

# Fonction pour créer les tables
init_database() {
    # Initialize if no existing data
    if [ ! -s "$PGDATA/PG_VERSION" ]; then
        echo "Initializing PostgreSQL with user $DB_USER..."
        
        # Initialize with trust auth first, then change it
        su-exec postgres initdb \
            --pgdata="$PGDATA" \
            --username="$DB_USER" \
            --encoding=UTF8 \
            --locale=C.UTF-8 \
            --auth=trust \
            --data-checksums
        
        # Start PostgreSQL temporarily with trust auth
        su-exec postgres pg_ctl start -D "$PGDATA" -w
        
        # Set the password for the superuser
        psql_cmd="psql -v ON_ERROR_STOP=1 --username $DB_USER --dbname postgres"
        $psql_cmd <<EOSQL
ALTER USER "$DB_USER" WITH PASSWORD '$DB_PASSWORD';
EOSQL
        
        # Stop PostgreSQL
        su-exec postgres pg_ctl stop -D "$PGDATA" -m fast
    fi

    # Start PostgreSQL with proper auth
    su-exec postgres pg_ctl start -D "$PGDATA" -w

    # Using $DB_USER (nasdev) as superuser, connect to postgres database
    echo "Setting up database as $DB_USER..."
    export PGPASSWORD="$DB_PASSWORD"
    psql_cmd="psql -v ON_ERROR_STOP=1 --username $DB_USER --dbname postgres"

    # Create database if not exists
    if ! $psql_cmd -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
        echo "Creating database $DB_NAME..."
        $psql_cmd <<EOSQL
CREATE DATABASE "$DB_NAME";
GRANT ALL PRIVILEGES ON DATABASE "$DB_NAME" TO "$DB_USER";
EOSQL

    # Now connect to the new database to create tables
    echo "Creating tables in $DB_NAME..."
    psql_movies_cmd="psql -v ON_ERROR_STOP=1 --username $DB_USER --dbname $DB_NAME"
    $psql_movies_cmd <<EOSQL
CREATE TABLE movies (
    id TEXT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description VARCHAR(255)
);
EOSQL
    fi

    unset PGPASSWORD
    su-exec postgres pg_ctl stop -D "$PGDATA" -m fast
}

init_database

# Configure PostgreSQL
cat >> "$PGDATA/postgresql.conf" <<EOSQL
listen_addresses = '*'
port = ${INVENTORY_DB_PORT:-5432}
EOSQL

# Configure authentication (now using scram-sha-256 everywhere)
cat > "$PGDATA/pg_hba.conf" <<EOSQL
local   all             all                                     scram-sha-256
host    all             all             127.0.0.1/32            scram-sha-256
host    all             all             ::1/128                 scram-sha-256
host    all             all             0.0.0.0/0               scram-sha-256
EOSQL

exec su-exec postgres postgres -D "$PGDATA"