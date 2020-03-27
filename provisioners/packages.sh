#!/bin/bash
set -eux

curl -s http://169.254.169.254/ || { echo "not on ec2, aborting"; exit 1; }

a() {
  DEBIAN_FRONTEND=noninteractive apt-get --yes "$@";
}

a update
a upgrade
a update #yes, it's necesary to update again, or apt can't find ag and awscli

# DEV AND ADMIN TOOLS
a install vim silversearcher-ag

# PROD STUFF
a install nginx nginx-extras awscli lua-cjson lua-http jq

# certbot
a install software-properties-common
add-apt-repository universe
add-apt-repository ppa:certbot/certbot
a update
a install certbot python-certbot-nginx 
