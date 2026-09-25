#!/bin/sh
# Usage:
#   ./run.sh <env>   Levanta docker compose (docker-compose.<env>.yaml)
#
# Ejemplo:
#   ./run.sh local

if [ -z "$1" ]; then
  echo "Usage: ./run.sh <env>   (ej: ./run.sh local)"
  exit 1
fi

docker compose -f "docker-compose.$1.yaml" up --build
