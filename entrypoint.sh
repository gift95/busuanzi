#!/bin/sh

set -x

cd /app || exit

if [ ! -f "/app/expose/install.lock" ];then
  # busuanzi js API address
  if [ -n "$API_SERVER" ];then
    sed -i "s/http:\/\/127.0.0.1:8080/$API_SERVER/g" dist/busuanzi.js
    sed -i "s/https:\/\/bsz.hnlyx.top/$API_SERVER/g" dist/index.html
  fi
  mv dist /app/expose/dist
  mv config.yaml /app/expose/config.yaml
  touch /app/expose/install.lock
fi

exec /app/busuanzi -c /app/expose/config.yaml -d /app/expose/dist