package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	baseDir          string
	iniPath          string

	// Persistent configuration variables loaded from INI
	apiURL           string
	exePath          string
	configPath       string
	autoConnect      bool
	keepAlive        bool
	startWithWindows bool
)

func init() {
	rawExePath, err := os.Executable()
	if err != nil {
		baseDir = "."
	} else {
		// Go compiler natively wraps the temporary binary inside a 'go-build' folder token context
		if strings.Contains(rawExePath, "go-build") {
			baseDir, _ = os.Getwd()
		} else {
			baseDir = filepath.Dir(rawExePath)
		}
	}

	iniPath = filepath.Join(baseDir, "config.ini")
}

func setupConfig() {
	loadAndSyncConfig()
	// Force update registry
	toggleWindowsStartup(startWithWindows)
}

func checkDashboardEnabled() bool {
	var targetPath string
	if filepath.IsAbs(configPath) {
		targetPath = configPath
	} else {
		targetPath = filepath.Join(baseDir, configPath)
	}

	data, err := ioutil.ReadFile(targetPath)
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, "\"external_ui\"")
}

func loadAndSyncConfig() {
	apiURL = "127.0.0.1:9090"
	exePath = "sing-box.exe"
	configPath = "config.json"
	autoConnect = true
	keepAlive = true
	startWithWindows = false

	data, err := ioutil.ReadFile(iniPath)
	if err != nil {
		saveConfig()
		return
	}

	lines := strings.Split(string(data), "\n")
	hasChanges := false
	foundKeys := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			foundKeys[key] = true

			switch key {
			case "api_url":
				apiURL = val
			case "exe_path":
				exePath = val
			case "config_path":
				configPath = val
			case "auto_connect":
				autoConnect = (val == "true")
			case "keep_alive":
				keepAlive = (val == "true")
			case "start_with_windows":
				startWithWindows = (val == "true")
			}
		}
	}

	requiredKeys := []string{"api_url", "exe_path", "config_path", "auto_connect", "keep_alive", "start_with_windows"}
	for _, k := range requiredKeys {
		if !foundKeys[k] {
			hasChanges = true
		}
	}

	if hasChanges {
		saveConfig()
	}
}

func saveConfig() {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("api_url=%s\n", apiURL))
	buf.WriteString(fmt.Sprintf("exe_path=%s\n", exePath))
	buf.WriteString(fmt.Sprintf("config_path=%s\n", configPath))
	buf.WriteString(fmt.Sprintf("auto_connect=%t\n", autoConnect))
	buf.WriteString(fmt.Sprintf("keep_alive=%t\n", keepAlive))
	buf.WriteString(fmt.Sprintf("start_with_windows=%t\n", startWithWindows))

	_ = ioutil.WriteFile(iniPath, buf.Bytes(), 0644)
}
