package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/goburrow/modbus"
)

func getSystemUptimeSeconds() (float64, error) {
	switch runtime.GOOS {
	case "linux":
		return getUptimeLinux()
	case "windows":
		return getUptimeWindows()
	default:
		return 0, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func getUptimeLinux() (float64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}
	parts := strings.Fields(string(data))
	return strconv.ParseFloat(parts[0], 64)
}

func getUptimeWindows() (float64, error) {
	// GetTickCount64 returns milliseconds since system boot
	mod := syscall.NewLazyDLL("kernel32.dll")
	proc := mod.NewProc("GetTickCount64")
	ret, _, err := proc.Call()
	if err != syscall.Errno(0) {
		return 0, err
	}
	uptimeMillis := uint64(ret)
	return float64(uptimeMillis) / 1000.0, nil
}

func runColdStartMode(config *Config, client modbus.Client) bool {
	level, err := readBatteryLevel(client, config.Modbus.Register, config.Modbus.RegisterType)
	if err != nil {
		log.Printf("Error reading battery level: %v", err)
		return false
	}

	lastLevel := level

	if level <= config.Threshold {
		log.Printf("Battery level critical: %d%%", level)

		uptime, err := getSystemUptimeSeconds()
		if err != nil {
			log.Printf("Error getting OS uptime: %v", err)
			return false
		}

		if uptime < 180 {
			log.Println("System recently started. Monitoring battery for recovery...")

			start := time.Now()
			maxWait := 10 * time.Minute

			for {
				level, err = readBatteryLevel(client, config.Modbus.Register, config.Modbus.RegisterType)
				if err != nil {
					log.Printf("Error reading battery level: %v", err)
					return false
				}

				if level > lastLevel {
					log.Printf("Battery level increased. Shutdown postponed: %d%%", level)
				}
				if level < lastLevel {
					log.Printf("Battery level decreased. Proceeding to shutdown: %d%%", level)
					return false
				}
				if level > config.Threshold {
					log.Printf("Battery level back to safe: %d%%", level)
					return true
				}
				if time.Since(start) > maxWait {
					log.Println("Cold start wait time exceeded. Proceeding to shutdown.")
					return false
				}

				lastLevel = level
				time.Sleep(time.Duration(config.PollIntervalSeconds) * time.Second)
			}
		}
	}

	return false
}
