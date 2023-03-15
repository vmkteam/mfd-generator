db-test:
	@echo "Rebuilding test DB..."
	@dropdb --if-exists newsportal
	@createdb -E UTF-8 -O postgres -T template0 --lc-collate C --lc-ctype=ru_RU.UTF-8 newsportal
	@psql newsportal < schema.sql
