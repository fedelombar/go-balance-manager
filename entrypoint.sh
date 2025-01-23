#!/bin/sh

echo "Waiting PostgreSQL..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; do
  echo "PostgreSQL is not ready..."
  sleep 2
done
echo "PostgreSQL está listo."

if [ "$SEED" = "true" ]; then
  echo "Seeding predefined users..."
  ./go-balance-manager -addr=${APP_ADDR} -dbhost=${DB_HOST} -dbport=${DB_PORT} -dbuser=${DB_USER} -dbpass=${DB_PASS} -dbname=${DB_NAME} -seed=true
  exit 0
fi

# Start the application
echo "Starting Go Balance Manager..."
./go-balance-manager -addr=${APP_ADDR} -dbhost=${DB_HOST} -dbport=${DB_PORT} -dbuser=${DB_USER} -dbpass=${DB_PASS} -dbname=${DB_NAME}