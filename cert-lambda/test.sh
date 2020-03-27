#!/bin/bash
set -e
go build
export DEBUG=1
export SERVER_CERT=/home/rlk/go/src/github.com/letsencrypt/pebble/test/certs/localhost/cert.pem
# "https://acme-staging-v02.api.letsencrypt.org/directory"
export DIRECTORY_URL="https://localhost:14000/dir"
#"https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf" 
export TOS_URL="data:text/plain,Do%20what%20thou%20wilt"
./cert-lambda 

