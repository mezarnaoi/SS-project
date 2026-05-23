#!/usr/bin/env bash
# Generates CA, broker server, and client certificates for MQTT mTLS.
# Output goes to secrets/ (gitignored). Run once before docker-compose up.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SECRETS_DIR="$SCRIPT_DIR/../secrets"
mkdir -p "$SECRETS_DIR"
cd "$SECRETS_DIR"

# 1. (CA)
echo "[1/3] Generating CA key and self-signed certificate..."
openssl genrsa -out ca.key 4096

openssl req -new -x509 -days 3650 -key ca.key -out ca.crt \
    -subj "/C=RO/ST=Romania/L=Bucharest/O=SS-Web/OU=Security/CN=SS-Web-CA"

# 2. Broker
echo "[2/3] Generating broker (server) certificate..."
openssl genrsa -out server.key 2048

cat > server_ext.cnf << EOF
[req]
req_extensions     = v3_req
distinguished_name = req_distinguished_name
prompt             = no
[req_distinguished_name]
C  = RO
ST = Romania
L  = Bucharest
O  = SS-Web
OU = Broker
CN = broker
[v3_req]
subjectAltName = @alt_names
[alt_names]
DNS.1 = broker
DNS.2 = localhost
IP.1  = 127.0.0.1
IP.2 = 10.0.2.2
EOF

openssl req -new -key server.key -out server.csr -config server_ext.cnf

openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key \
    -CAcreateserial -out server.crt \
    -extensions v3_req -extfile server_ext.cnf

rm server.csr server_ext.cnf

# 3. (Client)
echo "[3/3] Generating web-service (client) certificate..."
openssl genrsa -out web.key 2048

openssl req -new -key web.key -out web.csr \
    -subj "/C=RO/ST=Romania/L=Bucharest/O=SS-Web/OU=WebClient/CN=web"

openssl x509 -req -days 365 -in web.csr -CA ca.crt -CAkey ca.key \
    -CAcreateserial -out web.crt

rm web.csr

# Verifying certificates
echo ""
echo "Verifying certificates..."
openssl verify -CAfile ca.crt server.crt
openssl verify -CAfile ca.crt web.crt

# 4. DB PHI encryption key (AES-256, 32 random bytes, base64-encoded)
if [ ! -f db_encryption.key ]; then
    echo "[4/4] Generating database PHI encryption key..."
    openssl rand -base64 32 > db_encryption.key
else
    echo "[4/4] db_encryption.key already exists, skipping."
fi

# Private keys: readable only by owner
chmod 600 ./*.key
chmod 644 ./*.crt
