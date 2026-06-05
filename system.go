package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func toggleWindowsStartup(register bool) {
	if runtime.GOOS != "windows" {
		return
	}
	exe, _ := os.Executable()
	keyPath := `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	appName := "SingboxTrayMonitor"

	if register {
		_ = exec.Command("reg", "add", keyPath, "/v", appName, "/t", "REG_SZ", "/d", exe, "/f").Start()
		fmt.Println("Registered to Windows Startup Registry successfully.")
	} else {
		_ = exec.Command("reg", "delete", keyPath, "/v", appName, "/f").Start()
		fmt.Println("Removed from Windows Startup Registry successfully.")
	}
}

func triggerSystemPopup(title, message string) {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("msg", "*", "/TIME:10", message).Start()
	case "darwin":
		script := fmt.Sprintf("display dialog %q with title %q buttons {\"確定\"} default button \"確定\" with icon caution", message, title)
		_ = exec.Command("osascript", "-e", script).Start()
	default:
		_ = exec.Command("zenity", "--warning", "--title="+title, "--text="+message, "--width=400").Start()
	}
}