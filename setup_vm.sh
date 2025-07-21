#!/bin/bash

# Setup script for GCP Compute Engine Ubuntu instance
echo "🚀 Setting up CPIMP Scanner on Compute Engine..."

# Update system
sudo apt update && sudo apt upgrade -y

# Install Go
echo "📦 Installing Go..."
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
export PATH=$PATH:/usr/local/go/bin

# Install Git (if not already installed)
sudo apt install -y git

# Install screen for background execution
sudo apt install -y screen

# Verify installations
echo "✅ Go version: $(go version)"
echo "✅ Git version: $(git --version)"

echo "🔧 Setup complete! Now you can:"
echo "1. Clone your repo: git clone <your-repo-url>"
echo "2. cd into the project directory"
echo "3. Run: go mod tidy"
echo "4. Start scanning: screen -S scanner -dm go run ."
echo "5. Monitor with: screen -r scanner"
echo "6. Detach with: Ctrl+A, then D" 