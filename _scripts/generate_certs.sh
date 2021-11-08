#!/usr/bin/env bash

EXPIRE_DAYS=2050
KEY_EXPIRE_DAYS=360
SCRIPT_PATH_REL="$(dirname $0)"
SCRIPT_PATH="$(readlink -f $SCRIPT_PATH_REL)"
CERT_PATH="$SCRIPT_PATH_REL/../cert"

mkdir $CERT_PATH
pushd $CERT_PATH

# 1. Generate CA's private key and self-signed certificate
openssl req -x509 -newkey rsa:4096 -days $KEY_EXPIRE_DAYS -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/CN=localhost"

echo "CA's self-signed certificate"
openssl x509 -in ca-cert.pem -noout -text

# 2. Generate web server's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/CN=localhost"

# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -req -in server-req.pem -days $EXPIRE_DAYS -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile $SCRIPT_PATH/server-ext.cnf

echo "Server's signed certificate"
openssl x509 -in server-cert.pem -noout -text

popd