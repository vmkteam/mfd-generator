#!/bin/sh

set -eu

if [ "$#" -eq 0 ]; then
	echo "usage: $0 command [args...]" >&2
	exit 2
fi

database="mfd_test_$(date +%s)_$$"
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
created=0

cleanup() {
	if [ "$created" -eq 1 ]; then
		PGHOST=localhost PGPORT=5432 PGUSER=topscan \
			dropdb --if-exists --maintenance-db=postgres "$database" >/dev/null 2>&1 || true
	fi
}
trap cleanup EXIT INT TERM

PGHOST=localhost PGPORT=5432 PGUSER=topscan \
	createdb --maintenance-db=postgres --owner=topscan "$database"
created=1

PGHOST=localhost PGPORT=5432 PGUSER=topscan \
	psql --dbname="$database" --file="$script_dir/schema.sql" >/dev/null

dsn="postgres://topscan@localhost:5432/${database}?sslmode=disable"
# go-pg does not read ~/.pgpass, so keep the password in the child process environment.
PGPASSWORD="${PGPASSWORD:-topscan}" DB_DSN="$dsn" "$@"
