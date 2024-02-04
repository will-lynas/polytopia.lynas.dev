#!/bin/bash
set -e
docker build -t lynas_dev .
docker run -p 443:8443 lynas_dev
