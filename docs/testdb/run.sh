#!/bin/sh

set -eu

if [ "$#" -eq 0 ]; then
	echo "usage: $0 command [args...]" >&2
	exit 2
fi

port=${TESTDB_PORT:-55432}
container="mfd-testdb-$$"
dsn="postgres://postgres:postgres@localhost:${port}/newsportal?sslmode=disable"

cleanup() {
	docker stop "$container" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

docker run --rm --name "$container" \
	-e POSTGRES_PASSWORD=postgres \
	-e POSTGRES_DB=newsportal \
	-p "${port}:5432" \
	-d postgres:16.4 >/dev/null

attempt=0
while ! docker exec "$container" psql -U postgres -d newsportal -c 'SELECT 1' >/dev/null 2>&1; do
	attempt=$((attempt + 1))
	if [ "$attempt" -ge 30 ]; then
		echo "PostgreSQL did not become ready" >&2
		exit 1
	fi
	sleep 1
done

docker exec -i "$container" psql -U postgres -d newsportal < "$(dirname "$0")/schema.sql" >/dev/null
DB_DSN="$dsn" "$@"
