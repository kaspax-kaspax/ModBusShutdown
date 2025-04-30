package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/goburrow/modbus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Modbus struct {
		IP           string `yaml:"ip"`
		Port         int    `yaml:"port"`
		SlaveID      byte   `yaml:"slave_id"`
		Register     uint16 `yaml:"register"`
		RegisterType string `yaml:"register_type"` // "input" or "holding"
	} `yaml:"modbus"`

	Threshold    int           `yaml:"threshold"`
	PollInterval time.Duration `yaml:"poll_interval"`
	LogFile      string        `yaml:"log_file"`

	Email struct {
		SMTPServer string `yaml:"smtp_server"`
		SMTPPort   int    `yaml:"smtp_port"`
		Username   string `yaml:"username"`
		Password   string `yaml:"password"`
		From       string `yaml:"from"`
		To         string `yaml:"to"`
		Subject    string `yaml:"subject"`
	} `yaml:"email"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func setupLogging(file string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(f)
}

func sendEmail(cfg Config, message string) {
	auth := smtp.PlainAuth("", cfg.Email.Username, cfg.Email.Password, cfg.Email.SMTPServer)
	addr := fmt.Sprintf("%s:%d", cfg.Email.SMTPServer, cfg.Email.SMTPPort)

	msg := []byte("To: " + cfg.Email.To + "\r\n" +
		"Subject: " + cfg.Email.Subject + "\r\n" +
		"\r\n" + message + "\r\n")

	err := smtp.SendMail(addr, auth, cfg.Email.From, []string{cfg.Email.To}, msg)
	if err != nil {
		log.Printf("Email failed: %v", err)
	} else {
		log.Println("Notification email sent.")
	}
}

func shutdownSystem() {
	log.Println("Battery level critical. Shutting down system.")
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("shutdown", "/s", "/t", "30")
	case "linux":
		cmd = exec.Command("shutdown", "-h", "+1")
	default:
		log.Fatalf("Unsupported OS: %s", runtime.GOOS)
	}
	if err := cmd.Start(); err != nil {
		log.Fatalf("Shutdown command failed: %v", err)
	}
}

func readBatteryLevel(client modbus.Client, register uint16, regType string) (int, error) {
	var result []byte
	var err error

	switch regType {
	case "input":
		result, err = client.ReadInputRegisters(register, 1)
	case "holding":
		result, err = client.ReadHoldingRegisters(register, 1)
	default:
		return 0, fmt.Errorf("invalid register_type: %s", regType)
	}

	if err != nil {
		return 0, err
	}

	if len(result) < 2 {
		return 0, fmt.Errorf("invalid response length")
	}

	value := binary.BigEndian.Uint16(result)
	return int(value), nil
}

func main() {
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config.yaml: %v", err)
	}

	setupLogging(config.LogFile)
	log.Println("=== Modbus Shutdown Monitor Started ===")

	handler := modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", config.Modbus.IP, config.Modbus.Port))
	handler.SlaveId = config.Modbus.SlaveID
	handler.Timeout = 5 * time.Second

	err = handler.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to Modbus device: %v", err)
	}
	defer handler.Close()

	client := modbus.NewClient(handler)

	for {
		level, err := readBatteryLevel(client, config.Modbus.Register, config.Modbus.RegisterType)
		if err != nil {
			log.Printf("Error reading battery level: %v", err)
		} else {
			log.Printf("Battery level: %d%%", level)
			if level <= config.Threshold {
				log.Printf("Battery level critical: %d%%", level)
				sendEmail(*config, fmt.Sprintf("Battery level is %d%%. System shutdown is starting.", level))
				shutdownSystem()
				break
			}
		}
		time.Sleep(config.PollInterval)
	}
}
