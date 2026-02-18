#!/usr/bin/env sh

set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
OUT_DIR="$SCRIPT_DIR/../secrets"

PRIVATE_KEY="$OUT_DIR/private.pem"
PUBLIC_KEY="$OUT_DIR/public.pem"

mkdir -p "$OUT_DIR"

echo "Generating RSA 2048 keys..."

# Генерация приватного ключа
openssl genrsa -out "$PRIVATE_KEY" 2048

# Генерация публичного ключа
openssl rsa -in "$PRIVATE_KEY" -pubout -out "$PUBLIC_KEY"

# Права на ключи
chmod 600 "$PRIVATE_KEY"
chmod 644 "$PUBLIC_KEY"

echo "Done."
echo "Private key: $PRIVATE_KEY"
echo "Public key:  $PUBLIC_KEY"