#!/bin/bash

# Test script to verify multi-platform Docker build support
# This script tests that the Docker image can be built for both amd64 and arm64 platforms

set -e

echo "🔧 Testing multi-platform Docker build support for GSM..."

# Check if Docker buildx is available
if ! docker buildx version >/dev/null 2>&1; then
    echo "❌ Docker buildx is not available. Please install Docker Desktop or update Docker to support buildx."
    exit 1
fi

echo "✅ Docker buildx is available"

# List available builders and platforms
echo "📋 Available buildx builders and platforms:"
docker buildx ls

# Test multi-platform build
echo "🚀 Building multi-platform Docker image..."
docker buildx build --platform linux/amd64,linux/arm64 -t gsm-multiplatform:test .

echo "✅ Multi-platform build completed successfully!"

# Test docker-compose build
echo "🐳 Testing docker-compose build..."
docker-compose build secret-manager-emulator

echo "✅ Docker-compose build completed successfully!"

# Test running the image
echo "🧪 Testing the built image..."
docker run -d --name gsm-multiplatform-test -p 8087:8085 gsm-multiplatform:test

# Wait for the service to start
echo "⏳ Waiting for service to start..."
sleep 5

# Test health endpoint
if curl -f http://localhost:8087/health > /dev/null 2>&1; then
    echo "✅ Health check passed!"
else
    echo "❌ Health check failed!"
    docker logs gsm-multiplatform-test
    docker stop gsm-multiplatform-test
    docker rm gsm-multiplatform-test
    exit 1
fi

# Cleanup
docker stop gsm-multiplatform-test
docker rm gsm-multiplatform-test

echo "🎉 All multi-platform tests passed successfully!"
echo ""
echo "📖 Usage Instructions:"
echo "  • For local development: docker-compose up secret-manager-emulator"
echo "  • For pre-built image: docker-compose up gsm-emulator"
echo "  • For manual build: docker buildx build --platform linux/amd64,linux/arm64 -t your-tag ."
echo ""
echo "💡 The Docker image now supports both linux/amd64 and linux/arm64 platforms,"
echo "   which means it will work on both Intel/AMD and Apple Silicon (ARM) machines."