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
		resetSystemProxy()
		fmt.Println("Sing-box tray monitor stopped.")
}

func startProxy() {
	if spawnedPid > 0 && runtime.GOOS == "windows" {
		_ = exec.Command("taskkill", "/F", "/PID", strconv.Itoa(spawnedPid)).Run()
		spawnedPid = 0
	}
	
	// Use absolue path to avoid issue on startup with windows
	var absExePath, absConfigPath string

	if filepath.IsAbs(exePath) {
		absExePath = exePath
	} else {
		absExePath = filepath.Join(baseDir, exePath)
	}

	if filepath.IsAbs(configPath) {
		absConfigPath = configPath
	} else {
		absConfigPath = filepath.Join(baseDir, configPath)
	}

	_, errExe := os.Stat(absExePath)
	_, errConfig := os.Stat(absConfigPath)
	
	if os.IsNotExist(errExe) || os.IsNotExist(errConfig) {
		msg := "Warning: Invalid configuration paths detected!\n"
		if os.IsNotExist(errExe) {
			msg += fmt.Sprintf("- Core executable not found: %s\n", absExePath)
		}
		if os.IsNotExist(errConfig) {
			msg += fmt.Sprintf("- Configuration file not found: %s\n", absConfigPath)
		}
		msg += "\nPlease check and correct the paths in config.ini, then click Start Proxy again."
		
		go triggerSystemPopup("Singbox Tray Monitor - Configuration Error", msg)

		mToggle.SetTitle("Start Proxy")
		mToggle.Enable()
		return
	}
	
	cmd = exec.Command(absExePath, "run", "-c", absConfigPath)
	cmd.Dir = filepath.Dir(absExePath)

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

	resetSystemProxy()
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

func resetSystemProxy() {
	if runtime.GOOS != "windows" {
		return
	}
	internetSettingsKey := `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`
	_ = exec.Command("reg", "add", internetSettingsKey, "/v", "ProxyEnable", "/t", "REG_DWORD", "/d", "0", "/f").Run()
	_ = exec.Command("reg", "add", internetSettingsKey, "/v", "ProxyServer", "/t", "REG_SZ", "/d", "", "/f").Run()
}
