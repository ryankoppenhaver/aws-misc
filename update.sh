#!/bin/bash
set -eux

packer build server.packer.json
cd pulumi
PULUMI_CONFIG_PASSPHRASE="" pulumi up
