#!/bin/bash
#  --force-recreate
rm -r apisix_log || true && mkdir apisix_log && touch apisix_log/.gitkeep && docker compose -p apisix -f docker-compose-arm64.yml up -d --build --remove-orphans --no-deps