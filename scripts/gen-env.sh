#!/usr/bin/env bash
# Generates the .env file for docker-compose.
# UID/GID are detected automatically; MongoDB password is prompted interactively.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/../.env"

read -r -s -p "Enter MongoDB root password: " MONGO_PASSWORD
echo ""

DETECTED_IP=$(hostname -I | awk '{print $1}')

cat > "$ENV_FILE" << EOF
UID=$(id -u)                               # User ID local 
GID=$(id -g)                                # Group ID local 
MONGO_INITDB_ROOT_USERNAME=admin      # Username MongoDB
MONGO_INITDB_ROOT_PASSWORD=${MONGO_PASSWORD} # Parolă MongoDB
JWT_SECRET=dev-secret                 # Secret pentru JWT
AWS_ACCESS_KEY=local-aws-access       # Opțional: pentru S3
AWS_SECRET_KEY=local-aws-secret       # Opțional: pentru S3
AWS_REGION=us-east-1                  # Opțional: pentru S3
S3_BUCKET_NAME=local-bucket           # Opțional: pentru S3
MQTT_HOST_IP=${DETECTED_IP}             # IP-ul host-ului pentru MQTT
EOF

echo ".env generat: $ENV_FILE"