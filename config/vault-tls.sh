#!/bin/bash

# Create CA certificate config
cat <<EOF > "/tmp/ca-conf.cnf"
[ req ]
distinguished_name  = req_distinguished_name
x509_extensions     = root_ca

[ req_distinguished_name ]
countryName             = IL
countryName_min         = 2
countryName_max         = 2
stateOrProvinceName     = IL
localityName            = IL
0.organizationName      = KeyVault
organizationalUnitName  = KeyVault
commonName              = $VAULT_EXTERNAL_ADDRESS
commonName_max          = 64
emailAddress            = support@bloxstaking.com
emailAddress_max        = 64

[ root_ca ]
basicConstraints        = critical, CA:true
EOF

# Create server certificate config
cat <<EOF > "/tmp/server-conf.cnf"
subjectAltName = @alt_names
extendedKeyUsage = serverAuth

[alt_names]
IP.1 = 127.0.0.1
IP.2 = $VAULT_EXTERNAL_ADDRESS
EOF

# Create CA certificate and private key
openssl req -x509 \
    -newkey rsa:2048 \
    -out $VAULT_CACERT \
    -outform PEM \
    -keyout /vault/config/ca.key \
    -days 10000 \
    -verbose \
    -config /tmp/ca-conf.cnf \
    -nodes -sha256 \
    -subj "/CN=KeyVault CA"

# Create private key for server certificate
openssl req \
    -newkey rsa:2048 \
    -keyout /vault/config/server.key \
    -out /vault/config/server.req \
    -subj /CN=$VAULT_EXTERNAL_ADDRESS \
    -sha256 -nodes

# Issue server certificate
openssl x509 -req \
    -CA $VAULT_CACERT \
    -CAkey /vault/config/ca.key \
    -in /vault/config/server.req \
    -out /vault/config/server.pem \
    -days 10000 \
    -extfile /tmp/server-conf.cnf \
    -sha256 \
    -set_serial 0x1111
