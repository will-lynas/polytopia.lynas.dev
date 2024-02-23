#!/bin/bash
set -e

IMAGE_NAME="polytopia_lynas_dev_image"
CONTAINER_NAME="polytopia_lynas_dev_container"

docker stop $CONTAINER_NAME || true
docker rm $CONTAINER_NAME || true

docker build -t $IMAGE_NAME .
docker run -d --name $CONTAINER_NAME -p 443:8443 $IMAGE_NAME
