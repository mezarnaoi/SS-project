#!/usr/bin/env bash
# Generates the AES-256 key used to encrypt PHI fields in MongoDB
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SECRETS_DIR="$SCRIPT_DIR/../secrets"
mkdir -p "$SECRETS_DIR"

KEY_FILE="$SECRETS_DIR/db_encryption.key"

# if [ -f "$KEY_FILE" ]; then
#     echo "db_encryption.key already exists, skipping."
#     exit 0
# fi

openssl rand -base64 32 > "$KEY_FILE"
chmod 600 "$KEY_FILE"
echo "PHI encryption key generated: $KEY_FILE"
