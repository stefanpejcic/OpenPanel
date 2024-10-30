#!/bin/bash

IMAGE_NAME="openpanel:openpanel"
TAG="latest"


echo "Building the Docker image..."
cd /root/2083/
docker build -t $IMAGE_NAME:$TAG .

if [ $? -ne 0 ]; then
    echo "Error: Docker image build failed."
    exit 1
fi

echo "Docker image built successfully."

echo "Starting container for testing..."
docker run -d --name test_openpanel_container -p 2083:2083 $IMAGE_NAME:$TAG

sleep 5

echo "Testing the Flask app..."
curl -f http://localhost:2083 || {
    echo "Error: Flask app did not start correctly."
    docker rm -f test_openpanel_container
    exit 1
}

echo "Flask app is running successfully in Docker."

# Clean up
docker rm -f test_openpanel_container
echo "Cleaned up test container."
