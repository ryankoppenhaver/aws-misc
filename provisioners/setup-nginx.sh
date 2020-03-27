#!/bin/bash
set -eu

(
  GLOBIGNORE='*-mod-http-lua.conf:*-mod-http-ndk.conf'
  cd /etc/nginx/modules-enabled && rm *
)

rm /etc/nginx/sites-enabled/default

#4096 bit params, generated 2019-10-30
cat <<'END' > /etc/nginx/dhparams
-----BEGIN DH PARAMETERS-----
MIICCAKCAgEAlq38fjLiN2ElBnVdVUWdp2F8amNsrGlBR1szRoRa0MmvT+EkmEBB
Z0Zd/2atbnhINYd00p8UCQ7i2smpiGqjhN4uXy+huJ9sychhcSEt3l+I2SFwK9QZ
UOSFyeiB6lididn0DmPPsBp2pa0spJvdlC2kAFBIPl6Z2VJez1kFH5cxbURiNrRB
dSYu5ZYo28eviR0oJ07SWZv4h0N1LbqBTXIBuPCwnGegn+/nD2vjX5ALa+X2sjsh
5nlHaszBP9y/73Nx6/mUvwBOcubyGYUyCVs7alWTKu91alvOq9+kuhg/GDiHOtL6
ylwQbCMY4HU6Ek2x3Z6kwIY9h5jA2LlwXpIJfK8GnDbjqb3uTxXrj/HxgKIT8g5A
kdDNIzdTpoPIgbYJxpknvFT7aDzPDOLDGcvl51xQmSPAAWTaFY2EGOy1ZERvQpDe
ulxSqoSwM+HpyapkySrahe7Na1bzozXI91+xtD1FyHh2hyjD1pXrAys97VERvp5o
gxbjYV1lYnwcJvtGzsvBKUStcoT2yK/3e0JCZ5rKI/O8LZtGMmajq10q6ILCCSe9
ARTX0rJWPyJC0vBk/h3ArhkI9HGuntA99QUmAT1BCzJAxPQiecdR9NXzvjzeIRul
gla73uwvNDeCKEA+j/zwUhReAZ4DnLaNVPuQzWPTxIoxDv024G5kfsMCAQI=
-----END DH PARAMETERS-----
END

cat <<'END' > /etc/nginx/sites-available/www

lua_shared_dict s3_info 8k;

init_by_lua_block {
  local s3_info = ngx.shared.s3_info;
  local json = io.open("/etc/config.json"):read("*a")
  local config = cjson.decode(json)


  s3_info:set("bucket", "TKTK")
  s3_info:set("keyId", "TKTK")
  s3_info:set("TKTK","TKTK")
}


server {
  listen 80 default_server;
  #listen [::]:80 default_server;

  return 301 https://$host$request_uri;
}

server {
  listen 443 ssl default_server;
  #listen [::]:443 ssl default_server;

  ssl_certificate /etc/letsencrypt/live/bitflip.space/fullchain.pem;
  ssl_certificate_key /etc/letsencrypt/live/bitflip.space/privkey.pem;

  # taken from
  # https://ssl-config.mozilla.org/#server=nginx&server-version=1.14.0&config=intermediate&openssl-version=1.1.1
  # as of 2019-10-30
  ssl_session_timeout 1d;
  ssl_session_cache shared:MozSSL:10m;  # about 40000 sessions
  ssl_session_tickets off;
  ssl_protocols TLSv1.2 TLSv1.3;
  ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384;
  ssl_prefer_server_ciphers off;
  ssl_dhparam /etc/nginx/dhparams;

  add_header Strict-Transport-Security "max-age=63072000" always;

  #TODO add OCSP stapling

  # end mozilla config

  # temporary
  root /var/www/html;
  index index.nginx-debian.html;
  location /test {
    default_type 'text/plain';
    content_by_lua_block {
      ngx.say('hello world')
    }
  }
  location /p/ {
    #TODO - drop non-GET (will fail sig)

    #TODO figure out how to not hardcode this
    #proxy_pass https://www-data-57e41ee.s3.us-west-2.amazonaws.com/;
    proxy_pass http://localhost:9999/;

    set $upstream_date '';
    set $auth_str '';
    set_by_lua_block $auth_str {
      local key = "TKTK";

      ngx.var.upstream_date = os.date("!%a, %d %b %Y %H:%M:%S GMT ", os.time());

      local string_to_sign = ( 
        "GET\n" 
        .. "\n\n" -- no Content-MD5 or Content-Type
        .. ngx.var.upstream_date .. "\n"
        .. "" -- no amz headers
        .. ngx.var.request_uri
      )
      -- b64( hmac-sha1( key, string_to_sign ) )
      local sig = ngx.encode_base64(ngx.hmac_sha1(key, string_to_sign));

      return "Authorization: AWS TKTK:" .. sig  ;-- .. " DEBUG: " .. string_to_sign;
    }
    proxy_set_header Date $upstream_date;
    proxy_set_header Authorization $auth_str;
  }

  location / {
    try_files $uri $uri/ =404;
  }
  
}
END
ln -s /etc/nginx/sites-available/www /etc/nginx/sites-enabled/www
