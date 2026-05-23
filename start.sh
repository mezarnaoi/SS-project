#!/bin/bash
# Convenience wrapper for starting the development environment

mkdir -p uploads
./scripts/gen-env.sh
./scripts/gen-db-key.sh
./scripts/gen-certs.sh

docker compose up --build -d

cd client
yarn install
yarn dev

