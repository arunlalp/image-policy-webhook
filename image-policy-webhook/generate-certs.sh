#!/bin/bash
set -euo pipefail

SERVICE_NAME="image-policy-webhook"
NAMESPACE="default"
CERT_DIR="./certs"
KEY_DIR="./certs"
DAYS_VALID=365

mkdir -p ${CERT_DIR} ${KEY_DIR}

echo "[*] Generating CA..."
openssl genrsa -out ${KEY_DIR}/ca.key 2048
openssl req -x509 -new -nodes -key ${KEY_DIR}/ca.key -days ${DAYS_VALID} -out ${CERT_DIR}/ca.crt -subj "/CN=Webhook-CA"

echo "[*] Generating Server Certificate..."
openssl genrsa -out ${KEY_DIR}/tls.key 2048
openssl req -new -key ${KEY_DIR}/tls.key -out ${KEY_DIR}/server.csr -subj "/CN=${SERVICE_NAME}.${NAMESPACE}.svc"

cat <<EOF > ${KEY_DIR}/csr.conf
[req]
distinguished_name=req
[ v3_ext ]
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = ${SERVICE_NAME}
DNS.2 = ${SERVICE_NAME}.${NAMESPACE}
DNS.3 = ${SERVICE_NAME}.${NAMESPACE}.svc
EOF

openssl x509 -req -in ${KEY_DIR}/server.csr -CA ${CERT_DIR}/ca.crt -CAkey ${KEY_DIR}/ca.key -CAcreateserial \
-out ${CERT_DIR}/tls.crt -days ${DAYS_VALID} -extensions v3_ext -extfile ${KEY_DIR}/csr.conf

echo "[✔] Certificates generated in certs/"
