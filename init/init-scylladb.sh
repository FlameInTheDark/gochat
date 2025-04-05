#!/bin/bash
SCYLLA_HOST="scylla"

# Wait for ScyllaDB to be ready
until cqlsh "$SCYLLA_HOST" -e "DESCRIBE KEYSPACES"; do
    echo "Waiting for ScyllaDB to start..."
    sleep 5
done

# Create the keyspace
cqlsh "$SCYLLA_HOST" -e "CREATE KEYSPACE IF NOT EXISTS gochat WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}"