#!/bin/bash
set -eu
if certbot certificates | grep -q 'Found the following certs'; then
  exit
fi
aws s3 cp --recursive s3://rlk-general-private/certbot-config /etc/letsencrypt
certbot certonly --standalone -d bitflip.space -d www.bitflip.space --non-interactive
sed -i -Ee 's/authenticator = standalone/authenticator = nginx/' bitflip.space.conf
