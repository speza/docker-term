#!/usr/bin/env bash
docker build -t docker-term .
docker stop docker-term || true && docker rm docker-term || true
docker run -p 1000:8080 -d -v /var/run/docker.sock:/run/docker.sock --name docker-term docker-term