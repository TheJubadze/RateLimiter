@echo off
REM Set environment variables for PostgreSQL
set PGUSER=root
set PGPASSWORD=123
set POSTGRES_DB=rate-limiter

REM Pull PostgreSQL Docker image
docker pull postgres:16

REM Run PostgreSQL container with specified environment variables
docker run --name db -e POSTGRES_USER=%PGUSER% -e POSTGRES_PASSWORD=%PGPASSWORD% -e POSTGRES_DB=%POSTGRES_DB% -p 5432:5432 -d postgres:16

REM Wait for PostgreSQL to be ready
echo Waiting for PostgreSQL to start...
timeout /t 3

REM Check if container is running
docker ps | findstr db

if %ERRORLEVEL% == 0 (
    echo PostgreSQL container started successfully.
) else (
    echo Failed to start PostgreSQL container.
    exit /b 1
)

REM Connect to PostgreSQL and create the rate_limiter database if it does not exist
docker exec -it db psql -U %PGUSER% -d postgres -c "DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = '%POSTGRES_DB%') THEN CREATE DATABASE %POSTGRES_DB%; END IF; END $$;"

REM Notify the user
echo PostgreSQL installation and setup complete!
