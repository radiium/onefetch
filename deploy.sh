#!/bin/sh
set -e

USERNAME="radiium"
IMAGE_NAME="onefetch"
BUMP="${1:-patch}"   # major | minor | patch

# Lecture de la version courante
CURRENT=$(cat VERSION)
MAJOR=$(echo $CURRENT | cut -d. -f1)
MINOR=$(echo $CURRENT | cut -d. -f2)
PATCH=$(echo $CURRENT | cut -d. -f3)

# Incrément
case "$BUMP" in
  major) MAJOR=$((MAJOR+1)); MINOR=0; PATCH=0 ;;
  minor) MINOR=$((MINOR+1)); PATCH=0 ;;
  patch) PATCH=$((PATCH+1)) ;;
  *) echo "Usage: $0 [major|minor|patch]"; exit 1 ;;
esac

NEW_VERSION="$MAJOR.$MINOR.$PATCH"
echo "Bumping $CURRENT → $NEW_VERSION"

# Mise à jour des fichiers
echo "$NEW_VERSION" > VERSION
sed -i '' "s/\"version\": \".*\"/\"version\": \"$NEW_VERSION\"/" frontend/package.json

# Git commit + tag
git add VERSION frontend/package.json
git commit -m "chore: bump version to v$NEW_VERSION"
git tag -a "v$NEW_VERSION" -m "Release v$NEW_VERSION"
git push && git push --tags

# Docker build + push multi-arch
docker login

docker buildx create --name multiarch --driver docker-container --use 2>/dev/null || true
docker buildx use multiarch

docker buildx build \
  --build-arg APP_VERSION="$NEW_VERSION" \
  --platform linux/amd64,linux/arm64 \
  -t $USERNAME/$IMAGE_NAME:$NEW_VERSION \
  -t $USERNAME/$IMAGE_NAME:latest \
  --push \
  .

echo "✓ Released v$NEW_VERSION"
