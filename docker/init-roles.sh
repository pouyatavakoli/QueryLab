#!/bin/bash
set -e

echo "Setting QueryLab role passwords from environment variables..."

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
  ALTER ROLE querylab_admin WITH PASSWORD '$DB_ADMIN_PASSWORD';
  ALTER ROLE querylab_sandbox WITH PASSWORD '$DB_SANDBOX_PASSWORD';
EOSQL

echo "Passwords configured."
