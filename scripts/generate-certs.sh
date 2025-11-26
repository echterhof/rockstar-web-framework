#!/bin/bash
# Generate self-signed TLS certificates for QUIC server example
# These certificates are for development/testing purposes only

set -e

CERT_FILE="${1:-cert.pem}"
KEY_FILE="${2:-key.pem}"

echo "🔐 Generating TLS certificates for QUIC server..."
echo "   Certificate: $CERT_FILE"
echo "   Private Key: $KEY_FILE"
echo ""

# Generate self-signed certificate valid for 365 days
openssl req -x509 \
    -newkey rsa:4096 \
    -keyout "$KEY_FILE" \
    -out "$CERT_FILE" \
    -days 365 \
    -nodes \
    -subj '/CN=localhost' \
    -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

echo ""
echo "✅ Certificates generated successfully!"
echo ""
echo "⚠️  These are self-signed certificates for development only."
echo "   Do NOT use in production!"
echo ""
