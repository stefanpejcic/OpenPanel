# Hard Limits

Each user plan has fixed resource usage limits - these are **hard limits**, meaning users cannot exceed them under any circumstances.

- **Disk** and **Inode** [quotas](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/7/html/storage_administration_guide/ch-disk-quotas) are hard limits enforced by the the Linux kernel.
- **CPU** and **Memory** allocations are also enforced as hard limits using Docker. These resources are strictly capped to prevent overuse.

> **Note:** CPU usage percentages can be misleading on systems with multiple CPU cores. For example, if a service is allocated 1.5 CPU cores, its maximum usage may appear as **150%**, not **100%**.
> For further details, see: [Docker CLI Issue #2134](https://github.com/docker/cli/issues/2134)

