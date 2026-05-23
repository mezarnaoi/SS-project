#!/bin/bash
mkdir -p uploads
./scripts/gen-env.sh
./scripts/gen-db-key.sh
./scripts/gen-certs.sh

docker compose up --build -d

cd client
yarn install
yarn dev

