---
sidebar_position: 4
---

# Podman

*OpenAdmin > Services > Podman* gives Administrators visibility into the host's Podman installation and lets them manage the shared image store. It's only available to Administrators and Users, not Resellers. It has five tabs: **Info**, **Images**, **Volumes**, **Networks**, and **Disk Usage**.

### Info

An overview of the server's Podman: the number of running, paused and stopped rootful containers (root's own, users' rootless containers aren't counted), images, volumes and networks, how much of the container storage disk is used, and the Podman version with its OCI runtime and network backend. Below that are the host (OS, kernel, CPUs, uptime), runtime (Podman, OCI runtime, conmon, cgroups, security) and storage details. **Full podman info** opens the raw output of `podman info`.

![Podman page on the Info tab with the output of podman info](/img/openadmin-screenshots/services/podman-info.png#gh-light-mode-only)
![Podman page on the Info tab with the output of podman info](/img/openadmin-screenshots/services/podman-info_dark.png#gh-dark-mode-only)

### Images

Lists every image in the shared image store, the same store every hosting user's rootless Podman instance reads from, so an image only needs to be pulled once to be available to all users. The table shows:

- **Repository** / **Tag** / **Image ID** / **Size** / **Created**
- **Containers**: how many containers currently use the image, split into **N system** (root's own containers, e.g. the mail server) and **N user** (summed across every hosting user's own containers). An image with neither shows **Unused**.
- **Update**: **Check** compares the registry's current digest with every digest the local image has (no download, just a manifest fetch). If a newer digest is available it shows **Update available** with an **Update** button to re-pull; otherwise **Up to date** with a **Recheck** option. Pulling an update does **not** affect already-running containers, they keep using the content they started with until stopped/recreated.
- **Vulnerabilities**: filled by **Check all for vulnerabilities**, which scans the images with [Trivy](https://trivy.dev) for HIGH and CRITICAL CVEs. Each CVE counts once, even when it's in several packages. The badge shows how many have a fixed package (**fixable**) and how many have **no fix yet**, red when a critical one is fixable. Click it for the list, fixable CVEs first, the ones without a fix behind **Show … without a fix yet**. When a newer image is available it says so, pulling it usually removes most fixable CVEs. CVEs without a fix have no patched package from the distribution yet, so updating the image can't remove them.
- **Delete**: only offered for images with 0 system and 0 user containers using them.

![Podman Images tab listing container images with their tag, size and actions](/img/openadmin-screenshots/services/podman-images.png#gh-light-mode-only)
![Podman Images tab listing container images with their tag, size and actions](/img/openadmin-screenshots/services/podman-images_dark.png#gh-dark-mode-only)

The table also cross-references the compose stack used to provision new users (`/etc/openpanel/docker/compose/1.0/docker-compose.yml`). Any image that stack references but that isn't in the shared store yet shows as **Not downloaded**, with a one-click **Pull**.

Four bulk actions sit above the table:

- **Check all for updates**: runs the digest check against every downloaded image.
- **Check all for vulnerabilities**: scans every image with Trivy, installing it on first use. Scans run one image at a time at low priority and unpack images in `/var/tmp/openadmin-trivy` on disk, which is cleaned after every image.
- **Download all**: pulls every image the stack references that's currently missing.
- **Delete unused**: removes every currently-downloaded image with 0 system and 0 user containers.

Pull, Update and Download all put the image into the shared store users read from, so updated images reach their containers once they're recreated. All of these (per-image and bulk) run in the background with a progress toast, since a pull can take a while for a large image.

### Volumes / networks

Read-only listings of `podman volume ls` / `podman network ls` for the local context.

![Podman Volumes tab listing volumes](/img/openadmin-screenshots/services/podman-volumes.png#gh-light-mode-only)
![Podman Volumes tab listing volumes](/img/openadmin-screenshots/services/podman-volumes_dark.png#gh-dark-mode-only)

### Disk usage

`podman system df` output, Images reflects the whole shared store (every hosting user's images), while Containers and Volumes are root's own local Podman only, not aggregated across hosting users.

![Podman Disk Usage tab with the space used by images, containers and volumes](/img/openadmin-screenshots/services/podman-diskusage.png#gh-light-mode-only)
![Podman Disk Usage tab with the space used by images, containers and volumes](/img/openadmin-screenshots/services/podman-diskusage_dark.png#gh-dark-mode-only)

### Sorting and search

Every table column has click-to-sort arrows (frontend-only, no page reload), and each tab has a search box that filters across every visible column.
