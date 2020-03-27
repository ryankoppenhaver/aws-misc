#!/bin/bash
set -eux

cat <<'END' > /etc/systemd/system/download-config.service
[Unit]
Description=Retrieve Config from EC2 User Data
Requires=network.target
After=network.target
Before=nginx.service

[Service]
Type=oneshot
ExecStart=/usr/bin/curl http://169.254.169.254/latest/user-data -o /etc/config.json

[Install]
RequiredBy=nginx.service
END
systemctl enable download-config

exit
#### STUFF BELOW IS NOT WORKING

cat <<'END' > /root/poll-for-credentials #TKTK name differs
#!/bin/bash
while true; do
curl TKTKTKTKTKTKTKTKT
sleep 600
done
END
chmod +x /root/poll-for-credentials.rb

cat <<'END' > /etc/systemd/system/poll-for-credentials.service
[Unit]
Description=Poll for IAM creds
Requires=network.target
After=network.target
Before=nginx.service

[Service]
Type=simple
ExecStart=/root/poll-for-credentials.rb
[Install]
RequiredBy=nginx.service
END
systemctl enable poll-for-credentials
