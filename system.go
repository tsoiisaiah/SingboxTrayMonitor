package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"unsafe"
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
		h, _ := syscall.LoadDLL("user32.dll")
		p, _ := h.FindProc("MessageBoxW")
		uTitle := syscall.StringToUTF16Ptr(title)
		uMsg := syscall.StringToUTF16Ptr(message)
		
		// MessageBoxW (HWND, Text, Title, Flags)
		ret, _, err := p.Call(0, uintptr(unsafe.Pointer(uMsg)), uintptr(unsafe.Pointer(uTitle)), 0x40)
		if ret == 0 {
			fmt.Printf("[SysPopup] Popup Error (ret=0): %v\n", err)
		} else {
			fmt.Println("[SysPopup] Popup displayed successfully.")
		}

	case "darwin":
		script := fmt.Sprintf("display dialog %q with title %q buttons {\"OK\"} default button \"OK\" with icon caution", message, title)
		_ = exec.Command("osascript", "-e", script).Start()
	default:
		_ = exec.Command("zenity", "--warning", "--title="+title, "--text="+message, "--width=400").Start()
	}
}