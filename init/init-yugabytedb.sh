#!/bin/bash
set -euo pipefail

YUGABYTE_HOST="${YUGABYTE_HOST:-yugabyte}"
YUGABYTE_PORT="${YUGABYTE_PORT:-5433}"
YUGABYTE_USER="${YUGABYTE_USER:-yugabyte}"
YUGABYTE_DB="${YUGABYTE_DB:-gochat}"
YUGABYTE_COLOCATION="${YUGABYTE_COLOCATION:-false}"
YSQLSH="${YSQLSH:-/home/yugabyte/bin/ysqlsh}"

if [[ ! "$YUGABYTE_DB" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]]; then
    echo "YUGABYTE_DB must be a simple SQL identifier" >&2
    exit 1
fi

case "${YUGABYTE_COLOCATION,,}" in
    true|1|yes|on)
        YUGABYTE_COLOCATION_SQL="true"
        ;;
    false|0|no|off)
        YUGABYTE_COLOCATION_SQL="false"
        ;;
    *)
        echo "YUGABYTE_COLOCATION must be true or false" >&2
        exit 1
        ;;
esac

until "$YSQLSH" \
    -h "$YUGABYTE_HOST" \
    -p "$YUGABYTE_PORT" \
    -U "$YUGABYTE_USER" \
    -d yugabyte \
    -c "SELECT 1" >/dev/null 2>&1; do
    sleep 1
done

if "$YSQLSH" \
    -h "$YUGABYTE_HOST" \
    -p "$YUGABYTE_PORT" \
    -U "$YUGABYTE_USER" \
    -d yugabyte \
    -Atc "SELECT 1 FROM pg_database WHERE datname = '$YUGABYTE_DB'" | grep -q '^1$'; then
    echo "YugabyteDB database '$YUGABYTE_DB' already exists"
    EXISTING_COLOCATED="$("$YSQLSH" \
        -h "$YUGABYTE_HOST" \
        -p "$YUGABYTE_PORT" \
        -U "$YUGABYTE_USER" \
        -d "$YUGABYTE_DB" \
        -Atc "SELECT yb_is_database_colocated()")"
    if [[ "$EXISTING_COLOCATED" == "t" ]]; then
        echo "YugabyteDB database '$YUGABYTE_DB' is colocated"
    else
        echo "YugabyteDB database '$YUGABYTE_DB' is not colocated"
    fi
    if [[ "$YUGABYTE_COLOCATION_SQL" == "true" && "$EXISTING_COLOCATED" != "t" ]]; then
        echo "YugabyteDB database '$YUGABYTE_DB' is not colocated; create a new target database and migrate into it to enable colocation" >&2
    elif [[ "$YUGABYTE_COLOCATION_SQL" == "false" && "$EXISTING_COLOCATED" == "t" ]]; then
        echo "YugabyteDB database '$YUGABYTE_DB' is colocated but YUGABYTE_COLOCATION=false; create a new target database and migrate into it for the high-load layout" >&2
    fi
else
    "$YSQLSH" \
        -h "$YUGABYTE_HOST" \
        -p "$YUGABYTE_PORT" \
        -U "$YUGABYTE_USER" \
        -d yugabyte \
        -v ON_ERROR_STOP=1 \
        -c "CREATE DATABASE \"$YUGABYTE_DB\" WITH COLOCATION = $YUGABYTE_COLOCATION_SQL"
    echo "Created YugabyteDB database '$YUGABYTE_DB' with colocation=$YUGABYTE_COLOCATION_SQL"
fi
