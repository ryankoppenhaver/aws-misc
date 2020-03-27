This is an assortment of scripts and tools to set up some services in AWS; it's
largely an expermient / sandbox, and I'm publishing it primarily so I can
reference it as an existing invention in an employment agreement.

__Nothing in here is suitable for production use.  Seriously.__

===


General data flow is:

    Pulumi scripts + stack config => AWS Resources
      (lambda env vars, instance user data)

    CURRENT: download-config.service copies to /etc/config.json
    TODO: instance user data s/b a cloud-init script to populate config data locallly;

    Packer => Ansible => images
    Currently packer is top level, not controlled by pulumi; might change eventually

