---
sidebar_position: 5
---

# Change Image Tag

The **Docker > Change Image Tag** page allows you to update the Docker image tag (version) used by your services.

## Requirements

To access this feature:

- The **Docker** module must be enabled **server-wide** by an Administrator.
- Your account must have the **Docker** feature enabled.

## Usage

To change the tag for a Docker image:

1. First, check the available tags for the image on [Docker Hub](https://hub.docker.com/).
2. In the OpenPanel menu, go to **Docker > Change Image Tag**.
3. Under **Select Service**, choose the service for which you want to change the image tag.
4. Enter the **New Image Tag** in the input field.
5. Click the **Change Tag** button to apply the change.

After confirmation:

- The service will be automatically stopped.
- The new image tag will be pulled.
- The service will be restarted with the updated image.

> ⚠️ Make sure the new tag exists and is compatible with your current configuration to avoid service disruption.
