#!/bin/bash
set -eux

#todo- common helper lib?
a() {
  DEBIAN_FRONTEND=noninteractive apt-get --yes --quiet "$@";
}

snap remove amazon-ssm-agent
rm -rf /var/cache/snapd/aux # purge fails on this dir

a remove --purge apport apport-symptoms byobu command-not-found \
command-not-found-data cryptsetup cryptsetup-bin dosfstools eject \
nano parted snapd telnet

a autoremove
