---
sidebar_position: 2
---

# Process Manager

The **Process Manager** interface allows you to monitor all currently running processes on your server. You can search, view detailed command information, and terminate (kill) individual processes directly from the OpenPanel interface.

Processes are **sorted by CPU usage**, making it easy to identify and act on resource-intensive tasks.

## Key Features

- 🔍 **Search:** Quickly filter processes by container name, PID, or command.
- 🧾 **Detailed Info:** View key process metadata such as:
  - Container name
  - UID / PID / PPID
  - CPU %
  - Start Time (STIME)
  - TTY (Terminal)
  - Total Execution Time
  - Full Command (expandable)
- 🛑 **Kill Processes:** Force-stop any non-critical process.

:::danger
⚠️ **Warning:** Stopping core services like `MySQL`, `PHP-FPM`, or `Nginx/Apache` will cause your websites to go offline. Only terminate processes you are certain about.
:::

---

## How to Use

1. **Go to** `Advanced > Process Manager` in the OpenPanel sidebar.
2. Use the **search box** to find a specific process by PID, command, or container name.
3. Click **"View full command"** to expand long-running command strings.
4. Click the **Kill** button to immediately terminate the process.

## Interface Details

Each row in the table provides:

| Column | Description |
|--------|-------------|
| **Container** | The container/service where the process is running |
| **UID** | User ID of the process owner |
| **PID** | Unique Process ID |
| **PPID** | Parent Process ID |
| **CPU %** | CPU usage percentage |
| **STIME** | Process start time |
| **TTY** | Associated terminal (`?` means detached/background) |
| **TIME** | Total CPU time consumed |
| **CMD** | The command being executed |
| **Action** | Button to kill the process |

---

## Kill Process Behavior

When you click **Kill**, the following happens:

1. A notification appears: _“Terminating PID: xxxx...”_
2. A `POST` request is sent to the backend with the `PID` to terminate.
3. Upon success or failure, you’ll receive a follow-up toast with the result.

> This interface only shows filtered user-level processes. Internal or system-level maintenance commands (like `/etc/entrypoint.sh` or `ps -eo`) are excluded automatically.

---

## Tips

- For suspicious or high CPU usage processes, inspect the **full command** before taking action.
- Use the **Refresh Processes** button to reload the process list at any time.
- Killing a parent process (PPID) may also terminate its child processes.

---

Still have questions? Reach out to your server administrator or consult the system logs for deeper insights into recurring processes.
