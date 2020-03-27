#!/bin/bash
set -eux

#TODO better place for this
cat <<END > /root/certbot-init.sh
#!/bin/bash
set -eu
if certbot certificates | grep -q 'Found the following certs'; then
  exit
fi
aws s3 cp --recursive s3://rlk-general-private/certbot-config /etc/letsencrypt
certbot certonly --standalone -d bitflip.space -d www.bitflip.space --non-interactive
sed -i -Ee 's/authenticator = standalone/authenticator = nginx/' bitflip.space.conf 
END
chmod u+x /root/certbot-init.sh

#TODO abstract -d args here
cat <<END > /etc/systemd/system/certbot-init.service
[Unit]
Description=Certbot initial launch
Before=nginx.service

[Service]
Type=oneshot
ExecStart=/root/certbot-init.sh

[Install]
RequiredBy=nginx.service
END
systemctl enable certbot-init
