#!/bin/bash

set -e

# Configuration
IMAGE_NAME="registry.digitalocean.com/xask00/todo-webapp"
IMAGE_TAG="v2"
FULL_IMAGE_NAME="${IMAGE_NAME}:${IMAGE_TAG}"

echo "Building Docker image: ${FULL_IMAGE_NAME}"

# Build the Docker image
docker build -t ${FULL_IMAGE_NAME} .

echo "Docker image built successfully!"

# Push the image to DigitalOcean Container Registry
echo "Pushing image to DigitalOcean Container Registry..."
docker push ${FULL_IMAGE_NAME}

echo "Image pushed successfully!"

echo "Deploy is disabled do it manually for now .."
# # Optional: Deploy to Kubernetes if kubectl is configured
# if command -v kubectl &> /dev/null; then
#     echo "Updating Kubernetes deployment..."
    
#     # Update the deployment image
#     kubectl set image deployment/todo-webapp todo-webapp=${FULL_IMAGE_NAME}
    
#     # Wait for rollout to complete
#     kubectl rollout status deployment/todo-webapp
    
#     echo "Kubernetes deployment updated successfully!"
# else
#     echo "kubectl not found. Skipping Kubernetes deployment."
#     echo "To deploy manually, run:"
#     echo "kubectl set image deployment/todo-webapp todo-webapp=${FULL_IMAGE_NAME}"
# fi

# echo "Build and deploy completed!" 