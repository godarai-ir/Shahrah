@echo off
setlocal enabledelayedexpansion

REM Run ClickHouse migrations in order via docker compose service `clickhouse`
REM Usage:
REM   scripts\clickhouse-migrate.cmd
REM   scripts\clickhouse-migrate.cmd status

set "ACTION=%~1"
if "%ACTION%"=="" set "ACTION=up"

if /I "%ACTION%"=="status" goto :status
if /I not "%ACTION%"=="up" (
  echo Unknown action: %ACTION%
  echo Usage: %~nx0 [up^|status]
  exit /b 1
)

echo [1/3] Applying 001_gateway_model_events.sql
for /f "usebackq delims=" %%i in (`type "infra\clickhouse\migrations\001_gateway_model_events.sql"`) do echo %%i

type "infra\clickhouse\migrations\001_gateway_model_events.sql" | docker compose -f infra/compose/local.yml exec -T clickhouse clickhouse-client --multiquery
if errorlevel 1 goto :fail

echo [2/3] Applying 002_gateway_model_metrics_1m.sql
type "infra\clickhouse\migrations\002_gateway_model_metrics_1m.sql" | docker compose -f infra/compose/local.yml exec -T clickhouse clickhouse-client --multiquery
if errorlevel 1 goto :fail

echo [3/3] Applying 003_mv_gateway_model_metrics_1m.sql
type "infra\clickhouse\migrations\003_mv_gateway_model_metrics_1m.sql" | docker compose -f infra/compose/local.yml exec -T clickhouse clickhouse-client --multiquery
if errorlevel 1 goto :fail

echo Done. ClickHouse migrations applied.
goto :eof

:status
echo Checking ClickHouse tables...
docker compose -f infra/compose/local.yml exec -T clickhouse clickhouse-client -q "SHOW TABLES FROM sharah_db"
if errorlevel 1 goto :fail
goto :eof

:fail
echo Migration command failed.
exit /b 1
