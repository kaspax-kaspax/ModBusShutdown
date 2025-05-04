package main

import (
	"fmt"
	"log"

	"github.com/goburrow/modbus"
)

func runTestMode(config *Config, client modbus.Client) {
	log.Println("Running in test mode...")

	level, err := readBatteryLevel(client, config.Modbus.Register, config.Modbus.RegisterType)
	if err != nil {
		log.Fatalf("Test failed: could not read battery level: %v", err)
	}

	log.Printf("Battery level: %d%%", level)

	testMsg := fmt.Sprintf("Test mode: Battery level read as %d%%. Email is being sent as a test.", level)
	err = sendEmail(*config, testMsg)
	if err != nil {
		log.Fatalf("Test failed: could not send email: %v", err)
	}

	log.Println("!!! TEST SUCCESS: Battery read and email sent successfully !!!")
}

func runTestModeMail(config *Config) {
	log.Println("Running in Email test mode...")
	testMsg := "Test mode: Email is being sent as a test."
	err := sendEmail(*config, testMsg)
	if err != nil {
		log.Fatalf("Test failed: could not send email: %v", err)
	}

	log.Println("!!! TEST SUCCESS !!!")
}

func runTestModeModbuss(config *Config, client modbus.Client) {
	log.Println("Running in test mode...")

	level, err := readBatteryLevel(client, config.Modbus.Register, config.Modbus.RegisterType)
	if err != nil {
		log.Fatalf("Test failed: could not read battery level: %v", err)
	}
	log.Printf("Battery level: %d%%", level)
	log.Println("!!! TEST SUCCESS: Battery read !!!")
}
