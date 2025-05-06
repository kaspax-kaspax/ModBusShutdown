# ⚡ Modbus Shutdown Monitor

A lightweight Go application that monitors an inverter or UPS via **Modbus TCP**, checks battery level, logs status, sends an alert email, and cleanly shuts down a Windows or Linux system when battery drops below a critical threshold.

---

## ✅ Features

- 📡 Connects to Modbus TCP devices
- 🔍 Reads battery level from input or holding registers
- 🔐 Supports clean shutdown on battery critical condition
- ✉️ Sends email notifications before shutdown
- 📄 Logs all events to a rotating log file
- ⚙️ Configurable via `config.yaml`

---

## 💠 Requirements

- Go 1.18+
- Access to Modbus TCP inverter/UPS
- Working SMTP account for alerts (e.g. Brevo, Mailgun, Gmail)

---

## 📁 File Structure

```
ModBusShutdown/
├── main.go                 # Application source
├── testmode.go             # Application test mode source
├── coldstart.go            # Application coldstart mode source
├── configValidation.go     # config.yaml file validation
├── config.yaml             # External config file
├── go.mod
└── README.md
```

---

## ⚙️ Configuration

### 📄 `config.yaml`

```yaml
modbus:
  ip: "192.168.1.100"       # Inverter IP address or FQDN
  port: 502                 # Modbus TCP port (default: 502)
  slave_id: 1               # Modbus Unit ID
  register: 100             # Register holding battery %
  register_type: "input"    # "input" or "holding"

threshold: 20               # Battery level shutdown threshold (%)
poll_interval: 30           # Time between polling cycles in seconds.
alertthreshold: 85          # Battery level email alert threshold (%)
log_file: "modbus-shutdown.log"

email:
  smtp_server: "smtp.example.com"
  smtp_port: 587                    #Can be empty (default settings - 25 without auth, 587 with auth)
  username: "your@email.com"        #If empty then switch to no auth mode (by default port 25)
  password: "your-smtp-password"    #If empty then switch to no auth mode (by default port 25)
  from: "your@email.com"
  to: "admin@example.com"
  subject: "CRITICAL: Battery Low - System Shutdown"
```

---

## 🚀 Build & Run

### 1. Clone or download the project:

```bash
git clone https://github.com/kaspax-kaspax/ModBusShutdown.git
cd ModBusShutdown
```

### 2. Install dependencies:

```bash
go mod tidy
```

### 3. Build the app:

```bash
go build
```

### 4. Run:

```bash
./ModBusShutdown
```

Ensure `config.yaml` is in the **same directory** as the binary.

---

## 🥪 Test the Configuration


You can start by pointing to a known test register or log responses only by setting the threshold very low (e.g., 5) and monitoring the log file.

---

Sure! Here's a clean and corrected rewrite of that README section to explain **all available startup options**, including `--test`, `--config`, and separating test scenarios clearly:

---

## 🧭 Application Start Options

You can run the `ModBusShutdown` application with different command-line flags to control behavior:

### ▶️ Default Mode (normal operation)
```bash
ModBusShutdown.exe
```
- Starts the monitoring loop
- Continuously checks battery level
- Sends alert email and shuts down system if threshold is reached

---

### 🧪 Test Mode (Modbus + Email)
```bash
ModBusShutdown.exe --test
```
- Loads configuration
- Connects to Modbus device
- Reads current battery level
- Sends a test email
- Logs and prints results to console
- **Does not shut down or loop**

---

### ⚙️ Custom Configuration File
```bash
ModBusShutdown.exe --config="C:/path/to/custom.yaml"
```
- Uses an alternate configuration file path instead of the default `config.yaml`

---

### 🔀 Combined Example
```bash
ModBusShutdown.exe --config="C:/configs/test.yaml" --test
```
- Loads config from `test.yaml`
- Runs a one-time battery check + email alert
- Useful for staging environments or isolated testing

---

## 🔁 Running as a Service (Windows)

Use [nssm](https://nssm.cc/) to install as a background Windows service:

```bash
nssm install ModbusGuard "C:\Path\To\ModBusShutdown.exe"
```

---

## 📓 Logs

All status and errors are written to the file defined in `log_file` (e.g., `modbus-shutdown.log`).

---

## 🔌 Cold Start Detection (Power Restore Protection)

When a system boots after a power outage, the battery level may still be below the shutdown threshold, even though charging has resumed. To prevent immediate shutdown during this scenario, the Cold Start Mode adds smart recovery logic:

🔍 What It Does
    Detects recent boot using system uptime (less than 3 minutes)
    Reads battery level, then waits and monitors for a short period
    Suppresses shutdown if battery level increases (indicating charging)
    Fails safely (continues to shutdown) if battery does not improve or drops further

🧠 How It Works
    On startup, if the battery is below the configured threshold:
    The app checks system uptime.
    If uptime is under 3 minutes:
    The system is likely recovering from power loss.
    The app monitors the battery level for up to 15 minutes.
    If the battery is charging (level increases), shutdown is postponed.
    If the battery does not improve or drops further, the system shuts down normally.

---

## 🔐 Notes

- Make sure your Modbus register is accessible via TCP (not RTU).
- If the register is 1-based in docs, subtract 1 for Go (0-based).
- Email delivery may require allowing "less secure apps" or app passwords (e.g., Gmail).
- For production, use an SMTP service like [Brevo](https://www.brevo.com/), [Mailgun](https://www.mailgun.com/), or [SMTP2Go](https://www.smtp2go.com/).

---

## 🛡 Disclaimer

Use at your own risk. This tool initiates system shutdowns — test carefully before production deployment.

---

## 📬 Contact

Created by **kaspax**
📧 [kaspax@gmail.com](mailto:kaspax@gmail.com)

