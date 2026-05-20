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
  MIGRATION_SCOPE    all (default), yugabyte, yb, ysql, citus, postgres, pg, cassandra, or scylla
  MIGRATION_COMMAND  Optional default command when no CLI command is provided
  YUGABYTE_ADDRESS   YugabyteDB YSQL connection string. Falls back to PG_ADDRESS.
  CITUS_ADDRESS      Legacy Citus connection string. Falls back to PG_ADDRESS.
  PG_ADDRESS         Backward-compatible PostgreSQL connection string
  CASSANDRA_ADDRESS  Cassandra/ScyllaDB connection string

Examples:
  run-migrations
  run-migrations down 1
  MIGRATION_SCOPE=yugabyte run-migrations force 16
  MIGRATION_SCOPE=citus run-migrations force 16
EOF
}

normalize_scope() {
    case "$1" in
        all)
            printf '%s\n' "all"
            ;;
        yugabyte|yb|ysql)
            printf '%s\n' "yugabyte"
            ;;
        citus|postgres|pg)
            printf '%s\n' "citus"
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

resolve_yugabyte_address() {
    if [ -n "${YUGABYTE_ADDRESS:-}" ]; then
        printf '%s\n' "$YUGABYTE_ADDRESS"
        return
    fi
    printf '%s\n' "${PG_ADDRESS:-}"
}

resolve_citus_address() {
    if [ -n "${CITUS_ADDRESS:-}" ]; then
        printf '%s\n' "$CITUS_ADDRESS"
        return
    fi
    printf '%s\n' "${PG_ADDRESS:-}"
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
        yugabyte_address="$(resolve_yugabyte_address)"
        require_env "YUGABYTE_ADDRESS or PG_ADDRESS" "$yugabyte_address"
        require_env "CASSANDRA_ADDRESS" "${CASSANDRA_ADDRESS:-}"
        run_migration "yugabyte" "$yugabyte_address" "/migrations/yugabyte" "$@"
        run_migration "cassandra" "$CASSANDRA_ADDRESS" "/migrations/cassandra" "$@"
        ;;
    yugabyte)
        yugabyte_address="$(resolve_yugabyte_address)"
        require_env "YUGABYTE_ADDRESS or PG_ADDRESS" "$yugabyte_address"
        run_migration "yugabyte" "$yugabyte_address" "/migrations/yugabyte" "$@"
        ;;
    citus)
        citus_address="$(resolve_citus_address)"
        require_env "CITUS_ADDRESS or PG_ADDRESS" "$citus_address"
        run_migration "citus" "$citus_address" "/migrations/postgres" "$@"
        ;;
    cassandra)
        require_env "CASSANDRA_ADDRESS" "${CASSANDRA_ADDRESS:-}"
        run_migration "cassandra" "$CASSANDRA_ADDRESS" "/migrations/cassandra" "$@"
        ;;
esac
