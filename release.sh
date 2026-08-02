#!/bin/bash
# Create GitHub release for lazy-organizer v1.0.0
# Usage: GITHUB_TOKEN=ghp_xxxxx ./release.sh

set -e

REPO="dwiksurya/lazy-organizer"
TAG="v1.0.0"
TOKEN="${GITHUB_TOKEN:?Set GITHUB_TOKEN first}"

echo "Creating release $TAG on $REPO..."

# Create release
RELEASE_ID=$(curl -s -X POST \
  -H "Authorization: token $TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  "https://api.github.com/repos/$REPO/releases" \
  -d "$(cat <<EOF
{
  "tag_name": "$TAG",
  "name": "lazy-organizer $TAG",
  "body": "$(cat RELEASE_NOTES.md | sed 's/"/\\"/g' | sed ':a;N;$!ba;s/\n/\\n/g')",
  "draft": false,
  "prerelease": false
}
EOF
)" | jq -r '.id')

if [ "$RELEASE_ID" = "null" ] || [ -z "$RELEASE_ID" ]; then
  echo "Failed to create release. Response:"
  curl -s -X POST \
    -H "Authorization: token $TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    "https://api.github.com/repos/$REPO/releases" \
    -d '{"tag_name":"'$TAG'","name":"lazy-organizer '$TAG'","body":"Initial release","draft":false,"prerelease":false}' | jq .
  exit 1
fi

echo "Release created: ID=$RELEASE_ID"

# Upload assets
for file in dist/*.tar.gz; do
  name=$(basename "$file")
  echo "Uploading $name..."
  curl -s -X POST \
    -H "Authorization: token $TOKEN" \
    -H "Content-Type: application/gzip" \
    "https://api.github.com/repos/$REPO/releases/$RELEASE_ID/assets?name=$name" \
    --data-binary @"$file" | jq -r '.state // .message'
done

echo "Done! Release: https://github.com/$REPO/releases/tag/$TAG"
