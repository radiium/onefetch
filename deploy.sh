#!/bin/sh

USERNAME="radiium"
IMAGE_NAME="onefetch"
VERSION="latest"

docker login

# Créer et utiliser le builder multi-arch (si pas déjà fait)
docker buildx create --name multiarch --driver docker-container --use 2>/dev/null || true
docker buildx use multiarch

# Builder et pousser pour AMD64 et ARM64
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t $USERNAME/$IMAGE_NAME:$VERSION \
  --push \
  .