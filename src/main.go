package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func checkI2C() error {
	devices, _ := filepath.Glob("/dev/i2c-*")
	if len(devices) == 0 {
		return errors.New("no /dev/i2c-* device found: load the i2c-dev module " +
			"(sudo modprobe i2c-dev, and add it to /etc/modules-load.d/ to persist)")
	}

	for _, dev := range devices {
		f, err := os.OpenFile(dev, os.O_RDWR, 0)
		if err == nil {
			f.Close()
			return nil
		}
	}

	return fmt.Errorf("no accessible /dev/i2c-* device (%d found, permission denied): "+
		"add your user to the i2c group (sudo usermod -aG i2c $USER) and log in again", len(devices))
}

func refreshMonitors(old []Monitor) []Monitor {
	if err := RedetectMonitors(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}

	fresh, err := GetMonitors()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return old
	}

	known := make(map[string]Monitor, len(old))
	for _, m := range old { known[m.Name] = m }

	for i := range fresh {
		prev, ok := known[fresh[i].Name]
		if ok {
			fresh[i].CurrentRotation = prev.CurrentRotation
		} else {
			fmt.Printf("new monitor: %s (%s)\n", fresh[i].Name, fresh[i].ModelName)
		}
		delete(known, fresh[i].Name)
	}
	for name := range known {
		fmt.Printf("monitor removed: %s\n", name)
	}

	return fresh
}

func connectedOutputs() string {
	paths, _ := filepath.Glob("/sys/class/drm/card*-*/status")
	sort.Strings(paths)

	var b strings.Builder
	for _, p := range paths {
		status, err := os.ReadFile(p)
		if err == nil && strings.TrimSpace(string(status)) == "connected" {
			b.WriteString(filepath.Base(filepath.Dir(p)))
			b.WriteByte(' ')
		}
	}
	
	return b.String()
}

const (
	pollInterval = 3 * time.Second
	maxFailures = 3
	emptyRetryInterval = 60 * time.Second
)

func main() {
	if err := checkI2C(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	monitors, err := GetMonitors()
	if err != nil { fmt.Fprintf(os.Stderr, "error: %v\n", err) }

	outputs := connectedOutputs()
	pendingRefresh := false
	failures := 0
	lastRefresh := time.Now()

	for {
		if len(monitors) == 0 && time.Since(lastRefresh) >= emptyRetryInterval {
			pendingRefresh = true
		}

		if pendingRefresh {
			monitors = refreshMonitors(monitors)
			pendingRefresh = false
			failures = 0
			lastRefresh = time.Now()
		}
		
		if cur := connectedOutputs(); cur != outputs {
			outputs = cur
			pendingRefresh = true
		}

		failed := false
		for i := range monitors {
			m := &monitors[i]
			rotation, err := m.GetOrientation()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s: %v\n", m.Name, err)
				failed = true
				continue
			}

			if m.CurrentRotation != rotation {
				err = rotate(m.Name, rotation)
				if err != nil { fmt.Fprintf(os.Stderr, "error: %v\n", err); continue }
				fmt.Printf("%s → %s %s\n", m.CurrentRotation.String(), m.Name, rotation.String())
				m.CurrentRotation = rotation
			}
		}

		if failed {
			failures++
			if failures >= maxFailures {
				fmt.Fprintf(os.Stderr, "orientation unreadable %d times in a row, redetecting monitors\n", failures)
				pendingRefresh = true
			}
		} else {
			failures = 0
		}

		time.Sleep(pollInterval)
	}
}
