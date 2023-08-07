#!/bin/bash

export PGPASSWORD=password

# Wait for the database to start
until psql -h timescaledb -U "postgres" -c 'SELECT 1' 2>/dev/null; do
  echo "waiting for up"
  sleep 1
done

# Run the SQL migration script
psql -h timescaledb -U "postgres" -f ./cpu_usage.sql

# Load CSV data into the table
psql -h timescaledb -U "postgres" -d "homework" -c "\COPY cpu_usage FROM './cpu_usage.csv' CSV HEADER"

