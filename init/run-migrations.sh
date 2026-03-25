#!/bin/sh
set -eu

DEFAULT_COMMAND="up"
MIGRATION_SCOPE="${MIGRATION_SCOPE:-all}"
MIGRATION_COMMAND="${MIGRATION_COMMAND:-}"

print_usage() {
    cat <<'EOF'
Usage:
  run-migrations [command] [arg]

Environment:
  MIGRATION_SCOPE    all (default), postgres, pg, cassandra, or scylla
  MIGRATION_COMMAND  Optional default command when no CLI command is provided
  PG_ADDRESS         PostgreSQL connection string
  CASSANDRA_ADDRESS  Cassandra/ScyllaDB connection string

Examples:
  run-migrations
  run-migrations down 1
  MIGRATION_SCOPE=postgres run-migrations force 16
EOF
}

normalize_scope() {
    case "$1" in
        all)
            printf '%s\n' "all"
            ;;
        postgres|pg)
            printf '%s\n' "postgres"
            ;;
        cassandra|scylla)
            printf '%s\n' "cassandra"
            ;;
        *)
            echo "Unsupported MIGRATION_SCOPE: $1" >&2
            exit 1
            ;;
    esac
}

require_env() {
    env_name="$1"
    env_value="$2"

    if [ -z "$env_value" ]; then
        echo "Missing required environment variable: $env_name" >&2
        exit 1
    fi
}

run_migration() {
    name="$1"
    database_url="$2"
    path="$3"
    shift 3

    echo "Running $name migrations with command: $MIGRATION_COMMAND"
    migrate -database "$database_url" -path "$path" "$MIGRATION_COMMAND" "$@"
}

if [ $# -gt 0 ]; then
    MIGRATION_COMMAND="$1"
    shift
fi

if [ -z "$MIGRATION_COMMAND" ]; then
    MIGRATION_COMMAND="$DEFAULT_COMMAND"
fi

case "$MIGRATION_COMMAND" in
    help|-help|--help)
        print_usage
        exit 0
        ;;
esac

MIGRATION_SCOPE="$(normalize_scope "$MIGRATION_SCOPE")"

case "$MIGRATION_SCOPE" in
    all)
        require_env "PG_ADDRESS" "${PG_ADDRESS:-}"
        require_env "CASSANDRA_ADDRESS" "${CASSANDRA_ADDRESS:-}"
        run_migration "postgres" "$PG_ADDRESS" "/migrations/postgres" "$@"
        run_migration "cassandra" "$CASSANDRA_ADDRESS" "/migrations/cassandra" "$@"
        ;;
    postgres)
        require_env "PG_ADDRESS" "${PG_ADDRESS:-}"
        run_migration "postgres" "$PG_ADDRESS" "/migrations/postgres" "$@"
        ;;
    cassandra)
        require_env "CASSANDRA_ADDRESS" "${CASSANDRA_ADDRESS:-}"
        run_migration "cassandra" "$CASSANDRA_ADDRESS" "/migrations/cassandra" "$@"
        ;;
esac
