#!/bin/sh
# Initialize registry htpasswd file in Docker volume
# This runs once before the registry container starts

ADMIN_USER="${ADMIN_USERNAME:-admin}"
ADMIN_PASS="${ADMIN_PASSWORD}"

if [ -z "$ADMIN_PASS" ]; then
    echo "Error: ADMIN_PASSWORD environment variable is required"
    exit 1
fi

echo "Creating htpasswd file for registry authentication..."
docker run --rm \
    -v hubble-registry-auth:/auth \
    httpd:alpine \
    sh -c "htpasswd -Bbn $ADMIN_USER $ADMIN_PASS > /auth/htpasswd"

echo "✓ Registry auth configured for user: $ADMIN_USER"
