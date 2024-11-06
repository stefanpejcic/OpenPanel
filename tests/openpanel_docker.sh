#!/bin/bash

IMAGE_NAME="openpanel"
REPOSITORY_IMG_NAME="openpanel/openpanel"
TAG="latest"

echo "Stopping existing OpenPanel container"
cd /root && docker compose down openpanel

echo "Deleting existing image "
docker image rm $IMAGE_NAME:$TAG

echo "Building the image..."
cd /root/2083/
cp -r /usr/local/admin/scripts /root/2083/scripts
docker build -t $IMAGE_NAME:$TAG .

if [ $? -ne 0 ]; then
    echo "Error: Docker image build failed."
    exit 1
fi

echo "Docker image built successfully."

echo "Starting container for testing..."
cd /root && docker compose up -d openpanel

sleep 5

echo "Testing the Flask app..."
curl -f http://localhost:2083 || {
    echo "Error: Flask app did not start correctly."
    cd /root && docker compose down openpanel
    exit 1
}

echo "Flask app is running successfully in Docker."
