package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

var (
	cmd        *exec.Cmd
	spawnedPid int 
)

func processCleanUp()	{
		if spawnedPid > 0 && runtime.GOOS == "windows" {
			fmt.Printf("Killing sing-box child process (PID: %d)...\n", spawnedPid)
			_ = exec.Command("taskkill", "/F", "/PID", strconv.Itoa(spawnedPid)).Run()
		}
		fmt.Println("Sing-box tray monitor stopped.")
}

func startProxy() {
	if spawnedPid > 0 && runtime.GOOS == "windows" {
		_ = exec.Command("taskkill", "/F", "/PID", strconv.Itoa(spawnedPid)).Run()
		spawnedPid = 0
	}

	_, errExe := os.Stat(exePath)
	_, errConfig := os.Stat(configPath)
	
	if os.IsNotExist(errExe) || os.IsNotExist(errConfig) {
		msg := "Warning: Invalid configuration paths detected!\n"
		if os.IsNotExist(errExe) {
			msg += fmt.Sprintf("- Core executable not found: %s\n", exePath)
		}
		if os.IsNotExist(errConfig) {
			msg += fmt.Sprintf("- Configuration file not found: %s\n", configPath)
		}
		msg += "\nPlease check and correct the paths in config.ini, then click Start Proxy again."
		
		go triggerSystemPopup("Singbox Tray Monitor - Configuration Error", msg)

		mToggle.SetTitle("Start Proxy")
		mToggle.Enable()
		return
	}

	cmd = exec.Command(exePath, "run", "-c", configPath)
	cmd.Dir = filepath.Dir(exePath)

	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: 0x00000008,
		}
	}

	err := cmd.Start()
	if err == nil && cmd.Process != nil {
		spawnedPid = cmd.Process.Pid
		fmt.Printf("Successfully spawned hidden sing-box with PID: %d inside directory: %s\n", spawnedPid, cmd.Dir)
	}
}

func stopProxy() {
	if spawnedPid > 0 && runtime.GOOS == "windows" {
		_ = exec.Command("taskkill", "/F", "/PID", strconv.Itoa(spawnedPid)).Run()
		spawnedPid = 0
		cmd = nil
	} else {
		if runtime.GOOS == "windows" {
			_ = exec.Command("taskkill", "/F", "/IM", "sing-box.exe").Run()
		}
	}
}

func isSingboxAlive() bool {
	client := http.Client{Timeout: 400 * time.Millisecond}
	resp, err := client.Get("http://" + apiURL + "/ui")
	if err == nil && resp.StatusCode == 200 {
		resp.Body.Close()
		return true
	}
	return false
}