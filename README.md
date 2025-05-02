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

## 🔢 Test Mode

You can run the app in **test mode** using the `ModBusShutdown.exe --test` ModBusShutdown.exe --testflag. This is useful for verifying:

- Your Modbus connection works
- The battery level can be read
- The email notification system is functioning

You can run the app in **test mode** using the `ModBusShutdown.exe --testmail` ModBusShutdown.exe --testflag. This is useful for verifying:

- The email notification system is functioning

You can run the app in **test mode** using the `ModBusShutdown.exe --testmodbuss` ModBusShutdown.exe --testflag. This is useful for verifying:

- Your Modbus connection works
- The battery level can be read
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

