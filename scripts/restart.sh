#!/usr/bin/env bash

sleep 2

cd /opt/armeria

./build/armeria \
    -data=/opt/armeria/data \
    -public=/opt/armeria/client/dist \
    -scripts=/opt/armeria/scripts \
    >> /var/log/armeria.log 2>&1 &