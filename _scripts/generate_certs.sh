#!/usr/bin/env bash

# RSA & Certficate
openssl genrsa -out server.key 2048
openssl req -new -x509 -sha256 -key server.key \
  -out server.crt -days 3650

# Certificate Signing Request
openssl req -new -sha256 -key server.key -out server.csr
openssl x509 -req -sha256 -in server.csr -signkey server.key \
  -out server.crt -days 3650 \
