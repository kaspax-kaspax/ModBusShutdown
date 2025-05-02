package main

import (
	"fmt"
	"log"
	"net"
)

func validateConfig(cfg *Config) error {
	// Check essential Modbus values
	if _, err := net.LookupHost(cfg.Modbus.IP); err != nil {
		return fmt.Errorf("modbus.ip or hostname cannot be resolved: %q (%v)", cfg.Modbus.IP, err)
	}
	if cfg.Modbus.Port == 0 {
		return fmt.Errorf("modbus.port must be set")
	}
	if cfg.Modbus.RegisterType != "input" && cfg.Modbus.RegisterType != "holding" {
		return fmt.Errorf("modbus.register_type must be 'input' or 'holding'")
	}

	// Battery threshold
	if cfg.Threshold <= 0 || cfg.Threshold > 100 {
		return fmt.Errorf("threshold must be between 1 and 100")
	}

	// Poll interval
	if cfg.PollIntervalSeconds < 1 || cfg.PollIntervalSeconds > 3600 {
		return fmt.Errorf("poll_interval_seconds must be between 1 and 3600")
	}

	// Email config (optional, but warn or fail if test/email mode is used)
	if cfg.Email.SMTPServer == "" || cfg.Email.To == "" || cfg.Email.From == "" {
		log.Println("Warning: email settings incomplete — email sending may fail.")
	}

	return nil
}
