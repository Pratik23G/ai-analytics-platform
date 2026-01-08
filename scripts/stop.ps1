Set-Location (Split-Path $PSScriptRoot -Parent)
docker compose -f .\docker\docker-compose.yml down -v
