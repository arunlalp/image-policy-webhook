#!/bin/bash
set -euo pipefail

# Default paths for certificates and keys
CERT_DIR="./certs"
KEY_DIR="./keys"
SERVICE_NAME=${SERVICE_NAME:-"image-policy-webhook"}
NAMESPACE=${NAMESPACE:-"default"}
DAYS_VALID=${DAYS_VALID:-"365"}

# Create directories if they don't exist
mkdir -p ${CERT_DIR} ${KEY_DIR}

echo "Generating certificates for imagePolicyWebhook admission controller..."

# Generate CA key and certificate
echo "Generating CA..."
openssl genrsa -out ${KEY_DIR}/ca.key 2048
openssl req -x509 -new -nodes -key ${KEY_DIR}/ca.key -days ${DAYS_VALID} -out ${CERT_DIR}/ca.crt -subj "/CN=webhook-ca"

# Generate server key and CSR
echo "Generating server certificates..."
openssl genrsa -out ${KEY_DIR}/webhook-server.key 2048
openssl req -new -key ${KEY_DIR}/webhook-server.key -out ${KEY_DIR}/webhook-server.csr -subj "/CN=${SERVICE_NAME}.${NAMESPACE}.svc" -config <(
cat <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${SERVICE_NAME}
DNS.2 = ${SERVICE_NAME}.${NAMESPACE}
DNS.3 = ${SERVICE_NAME}.${NAMESPACE}.svc
DNS.4 = ${SERVICE_NAME}.${NAMESPACE}.svc.cluster.local
EOF
)

# Sign the certificate with our CA
openssl x509 -req -in ${KEY_DIR}/webhook-server.csr -CA ${CERT_DIR}/ca.crt -CAkey ${KEY_DIR}/ca.key -CAcreateserial -out ${CERT_DIR}/webhook-server.crt -days ${DAYS_VALID} -extensions v3_req -extfile <(
cat <<EOF
[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${SERVICE_NAME}
DNS.2 = ${SERVICE_NAME}.${NAMESPACE}
DNS.3 = ${SERVICE_NAME}.${NAMESPACE}.svc
DNS.4 = ${SERVICE_NAME}.${NAMESPACE}.svc.cluster.local
EOF
)

# Encode the CA certificate for the Kubernetes configuration
BASE64_CA=$(cat ${CERT_DIR}/ca.crt | base64 | tr -d '\n')

echo "Certificates generated successfully!"
echo "CA Certificate (base64 encoded): ${BASE64_CA}"
echo ""
echo "You can use this CA in your Kubernetes webhook configuration."
echo "Make sure to create a Kubernetes secret with the server certificates:"
echo "kubectl create secret tls webhook-server-tls --cert=${CERT_DIR}/webhook-server.crt --key=${KEY_DIR}/webhook-server.key -n ${NAMESPACE}"