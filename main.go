package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
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
		IP                   string `yaml:"ip"`
		Port                 int    `yaml:"port"`
		SlaveID              byte   `yaml:"slave_id"`
		BatteryRegister      uint16 `yaml:"battery_register"`
		RegisterType         string `yaml:"register_type"`
		InputRegister        uint16 `yaml:"input_register"`
		NotConnectedRegister uint16 `yaml:"not_connected_input_value"`
	} `yaml:"modbus"`

	Threshold           int    `yaml:"threshold"`
	PollIntervalSeconds int    `yaml:"poll_interval"`
	AlertThreshold      int    `yaml:"alert_threshold"`
	LogFile             string `yaml:"log_file"`

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

func isGridMonitored(config *Config) bool {
	if config.Modbus.InputRegister == 0 || config.Modbus.NotConnectedRegister == 0 {
		return false
	}
	return true
}

func setupLogging(file string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
}

func sendEmail(cfg Config, message string) error {
	port := cfg.Email.SMTPPort
	if port == 0 {
		if cfg.Email.Username == "" {
			port = 25
		} else {
			port = 587
		}
	}
	addr := fmt.Sprintf("%s:%d", cfg.Email.SMTPServer, port)

	var auth smtp.Auth = nil
	if cfg.Email.Username != "" && cfg.Email.Password != "" {
		auth = smtp.PlainAuth("", cfg.Email.Username, cfg.Email.Password, cfg.Email.SMTPServer)
	}

	msg := []byte("To: " + cfg.Email.To + "\r\n" +
		"Subject: " + cfg.Email.Subject + "\r\n" +
		"\r\n" + message + "\r\n")

	err := smtp.SendMail(addr, auth, cfg.Email.From, []string{cfg.Email.To}, msg)
	if err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
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

func readGridState(client modbus.Client, register uint16, regType string, NotConnectedRegister uint16) (bool, error) {
	var result []byte
	var err error

	switch regType {
	case "input":
		result, err = client.ReadInputRegisters(register, 1)
	case "holding":
		result, err = client.ReadHoldingRegisters(register, 1)
	default:
		return false, fmt.Errorf("invalid register_type: %s", regType)
	}

	if err != nil {
		return false, err
	}

	if len(result) < 2 {
		return false, fmt.Errorf("invalid response length")
	}

	value := binary.BigEndian.Uint16(result)
	if value == NotConnectedRegister {
		return false, nil
	}
	return true, nil
}

func main() {
	testMode := flag.Bool("test", false, "Run in test mode (single Modbus check and email alert)")
	testModeMail := flag.Bool("testmail", false, "Run in test mode (check email alert)")
	testModeModbus := flag.Bool("testmodbus", false, "Run in test mode (single Modbus check)")
	configPath := flag.String("config", "config.yaml", "Path to configuration file. Example: -config=\"C:\\Modbus\\config\\config.yaml\"")

	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config.yaml: %v", err)
	}

	if err := validateConfig(config); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
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

	if *testMode {
		runTestMode(config, client)
		return
	}
	if *testModeMail {
		runTestModeMail(config)
		return
	}
	if *testModeModbus {
		runTestModeModbuss(config, client)
		return
	}

	if runColdStartMode(config, client) {
		log.Println("Cold start recovery successful. Skipping shutdown loop.")
		return
	}

	var AlertSet bool
	if config.AlertThreshold == 0 {
		AlertSet = false
	} else {
		AlertSet = true
	}
	AlertSend := false
	hostname, _ := os.Hostname()

	for {
		log.Println("Checking battery level...")
		level, err := readBatteryLevel(client, config.Modbus.BatteryRegister, config.Modbus.RegisterType)
		if err != nil {
			log.Printf("Error reading battery level: %v", err)
		} else {
			log.Printf("Battery level: %d%%", level)
			if level <= config.AlertThreshold && AlertSet && !AlertSend {
				log.Printf("Battery level alert: %d%%", level)
				sendEmail(*config, fmt.Sprintf("Battery level alert: Battery level is %d%% on %s.", level, hostname))
				AlertSend = true
			}
			if level > config.AlertThreshold && AlertSet && AlertSend {
				log.Printf("Battery level alert cleared: %d%%", level)

				sendEmail(*config, fmt.Sprintf("Battery level alert cleared: Battery level is %d%% on %s.", level, hostname))
				AlertSend = false
			}
			if level <= config.Threshold {
				log.Printf("Battery level critical: %d%%", level)
				sendEmail(*config, fmt.Sprintf("Battery level is %d%%. System %s shutdown is starting.", level, hostname))
				shutdownSystem()
				break
			}
		}
		time.Sleep(time.Duration(config.PollIntervalSeconds) * time.Second)

	}
}
