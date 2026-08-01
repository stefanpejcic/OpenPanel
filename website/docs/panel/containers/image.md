---
sidebar_position: 4
---

# Image Updates

The **Docker > Image Updates** page allows you to check if image updates are available for your images.

## Requirements

To access this feature:

- The **Docker** module must be enabled **server-wide** by an Administrator.
- Your account must have the **Docker** feature enabled.

## Checking for Updates

Click the refresh icon next to **Last checked** to scan all of your images for available updates. The check runs in the background and can take a few minutes to complete — the page will show "Check is in progress.." until results are ready.

Once results are available, a summary is shown at the top of the page:

- **Monitored Images** – Total number of images being checked.
- **Updates Available** – Images with a newer version available.
- **Minor Updates** – Images with an available minor version update.
- **Major Updates** – Images with an available major version update.
- **Up to Date** – Images already running the latest version.
- **Unknown** – Images whose update status could not be determined.

## Image Table

Below the summary, a table lists every monitored image with:

- **Image** – The image repository name (linked to its source, when available).
- **Tag** – The currently used tag.
- **Update** – Whether an update is available, the image is up to date, or the status is unknown.
- **Info** – Details about the available update (new tag, current/new version, and whether it's a minor or major update), or an error message if the check failed.
- **Actions** – Click **Delete** to permanently remove an image. An image that is currently used by a container cannot be deleted.

To change the tag used by a service once you know which version you want, use the [Change Image Tag](/docs/panel/containers/change/) page.
