#!/bin/bash

# Simple deployment script
echo "🚀 Starting YieldBasis Cap Monitor deployment..."

# Check if .env exists
if [ ! -f .env ]; then
    echo "❌ .env file not found!"
    echo "Please create .env with:"
    echo "INFURA_WS_URL=wss://mainnet.infura.io/ws/v3/YOUR_KEY"
    echo "TELEGRAM_BOT_TOKEN=your_bot_token"
    echo "TELEGRAM_CHAT_ID=your_chat_id"
    exit 1
fi

# Stop existing container
echo "🛑 Stopping existing containers..."
docker-compose down

# Build and start
echo "🔨 Building and starting cap-monitor..."
docker-compose up --build -d

# Show status
echo "✅ Deployment complete!"
echo "📊 Container status:"
docker-compose ps

echo "📝 View logs with:"
echo "docker-compose logs -f cap-monitor"