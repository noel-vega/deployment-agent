#!/bin/bash
set -e

echo "🚀 Deploying Hubble to Production"
echo "=================================="
echo ""

# Check if .env.production exists
if [ ! -f .env.production ]; then
    echo "❌ Error: .env.production not found!"
    echo "Please create .env.production with your production configuration"
    exit 1
fi

# Copy production env to .env
echo "📋 Using production environment configuration..."
cp .env.production .env

# Check required variables
echo "🔍 Checking required variables..."
source .env

if [ -z "$HUBBLE_DOMAIN" ]; then
    echo "❌ Error: HUBBLE_DOMAIN not set in .env.production"
    exit 1
fi

if [ -z "$HUBBLE_TRAEFIK_EMAIL" ]; then
    echo "❌ Error: HUBBLE_TRAEFIK_EMAIL not set in .env.production"
    echo "This is required for Let's Encrypt HTTPS certificates"
    exit 1
fi

if [ "$HUBBLE_DOMAIN" = "localhost" ]; then
    echo "⚠️  Warning: HUBBLE_DOMAIN is set to 'localhost'"
    echo "Let's Encrypt cannot issue certificates for localhost!"
    echo "Set HUBBLE_DOMAIN to your real domain (e.g., noelvega.dev)"
    exit 1
fi

echo "✓ Domain: $HUBBLE_DOMAIN"
echo "✓ Email: $HUBBLE_TRAEFIK_EMAIL"
echo ""

# Check DNS
echo "🌐 Checking DNS records..."
HUBBLE_IP=$(dig +short hubble.$HUBBLE_DOMAIN | tail -1)
REGISTRY_IP=$(dig +short registry.$HUBBLE_DOMAIN | tail -1)

if [ -z "$HUBBLE_IP" ]; then
    echo "⚠️  Warning: hubble.$HUBBLE_DOMAIN does not resolve to an IP"
    echo "Make sure DNS is configured: A record for hubble.$HUBBLE_DOMAIN"
else
    echo "✓ hubble.$HUBBLE_DOMAIN → $HUBBLE_IP"
fi

if [ -z "$REGISTRY_IP" ]; then
    echo "⚠️  Warning: registry.$HUBBLE_DOMAIN does not resolve to an IP"
    echo "Make sure DNS is configured: A record for registry.$HUBBLE_DOMAIN"
else
    echo "✓ registry.$HUBBLE_DOMAIN → $REGISTRY_IP"
fi

echo ""

# Ask for confirmation
read -p "Deploy to production with HTTPS enabled? (yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo "Deployment cancelled"
    exit 0
fi

echo ""
echo "🐳 Stopping existing services..."
docker compose -f docker-compose.yml -f docker-compose.prod.yml down 2>/dev/null || docker compose down

echo ""
echo "🏗️  Building services..."
docker compose -f docker-compose.yml -f docker-compose.prod.yml build

echo ""
echo "🚀 Starting services with HTTPS..."
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

echo ""
echo "⏳ Waiting for services to start..."
sleep 5

echo ""
echo "📊 Service Status:"
docker compose ps

echo ""
echo "📜 Traefik Logs (checking for Let's Encrypt):"
echo "-------------------------------------------"
docker compose logs hubble-traefik | tail -20

echo ""
echo "✅ Deployment Complete!"
echo ""
echo "🌐 Access your services at:"
echo "  - Web UI:   https://hubble.$HUBBLE_DOMAIN"
echo "  - API:      https://hubble.$HUBBLE_DOMAIN/api"
echo "  - Registry: https://registry.$HUBBLE_DOMAIN"
echo ""
echo "📊 Monitor logs with:"
echo "  docker compose logs -f"
echo ""
echo "🔒 Check Let's Encrypt certificate status:"
echo "  docker compose exec hubble-traefik cat /data/acme.json | jq"
echo ""
