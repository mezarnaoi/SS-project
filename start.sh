#!/bin/bash
mkdir -p uploads
./scripts/gen-env.sh
./scripts/gen-db-key.sh
./scripts/gen-certs.sh

openssl pkcs12 -export \
  -in secrets/web.crt \
  -inkey secrets/web.key \
  -certfile secrets/ca.crt \
  -out android_client.p12 \
  -name android-client \
  -passout pass:changeit

mkdir -p android-client/app/src/main/res/raw
mv android_client.p12 android-client/app/src/main/res/raw/android_client.p12
cp secrets/ca.crt android-client/app/src/main/res/raw/ca.crt

docker compose up --build -d

cd client
yarn install
yarn dev

