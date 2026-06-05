package main

import (
	"embed"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/energye/systray"
)

//go:embed assets/*
var assetFiles embed.FS

var (
	fastCheckCh chan bool

	mToggle    *systray.MenuItem
	mDashboard *systray.MenuItem
	isRunning  bool
	iconGreen  []byte
	iconRed    []byte

	// Submenu items under Settings
	mAutoConnect  *systray.MenuItem
	mKeepAlive    *systray.MenuItem
	mStartWithWin *systray.MenuItem
)

func setupUI() {
	fmt.Println("systray.onReady started")

	fastCheckCh = make(chan bool, 1)

	loadAndSyncConfig()

	iconGreen, _ = assetFiles.ReadFile("assets/online.ico")
	iconRed, _ = assetFiles.ReadFile("assets/offline.ico")

	if len(iconGreen) == 0 {
		fmt.Println("CRITICAL ERROR: Embedded assets/online.ico is empty!")
	}

	if len(iconRed) == 0 {
		fmt.Println("CRITICAL ERROR: Embedded assets/offline.ico is empty!")
	} else {
		systray.SetIcon(iconRed)
	}

	systray.SetTitle("Proxy")
	systray.SetTooltip("Sing-box Tray Monitor")

	systray.SetOnClick(func(menu systray.IMenu) {
		if menu != nil {
			menu.ShowMenu()
		}
	})
	systray.SetOnRClick(func(menu systray.IMenu) {
		if menu != nil {
			menu.ShowMenu()
		}
	})

	systray.CreateMenu()

	mToggle = systray.AddMenuItem("Checking Status...", "Verifying sing-box status")
	mToggle.Disable()

	mDashboard = systray.AddMenuItem("Open Dashboard", "Open MetaCubeXD web dashboard")
	if checkDashboardEnabled() {
		mDashboard.Enable()
	} else {
		mDashboard.Disable()
		mDashboard.SetTooltip("Dashboard is disabled in your config.json")
	}

	mSettings := systray.AddMenuItem("Settings", "Configure tray options")
	mAutoConnect = mSettings.AddSubMenuItemCheckbox("Auto Connect", "Automatically connect sing-box on start starts", autoConnect)
	mKeepAlive = mSettings.AddSubMenuItemCheckbox("Keep Alive", "Restart sing-box automatically if it crashes", keepAlive)
	mStartWithWin = mSettings.AddSubMenuItemCheckbox("Start with Windows", "Launch automatically on Windows startup", startWithWindows)

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Exit", "Stop core and exit application")

	mToggle.Click(func() {
		mToggle.Disable()
		mToggle.SetTitle("Processing...")

		loadAndSyncConfig()

		if !isRunning {
			startProxy()
		} else {
			stopProxy()
		}

		select {
		case fastCheckCh <- true:
		default:
		}
	})

	mDashboard.Click(func() {
		loadAndSyncConfig()
		openDashboardURL("http://" + apiURL + "/ui")
	})

	mAutoConnect.Click(func() {
		autoConnect = !mAutoConnect.Checked()
		if autoConnect {
			mAutoConnect.Check()
		} else {
			mAutoConnect.Uncheck()
		}
		saveConfig()
	})

	mKeepAlive.Click(func() {
		keepAlive = !mKeepAlive.Checked()
		if keepAlive {
			mKeepAlive.Check()
		} else {
			mKeepAlive.Uncheck()
		}
		saveConfig()
	})

	mStartWithWin.Click(func() {
		startWithWindows = !mStartWithWin.Checked()
		if startWithWindows {
			mStartWithWin.Check()
			toggleWindowsStartup(true)
		} else {
			mStartWithWin.Uncheck()
			toggleWindowsStartup(false)
		}
		saveConfig()
	})

	mQuit.Click(func() {
		systray.Quit()
	})

	checkStatusAndUpdateUI()

	if !isSingboxAlive() && autoConnect {
		fmt.Println("Auto-connect is active and core is offline, launching...")
		startProxy()
	}

	go func() {
		normalTicker := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-normalTicker.C:
				checkStatusAndUpdateUI()
				
				if keepAlive && spawnedPid > 0 && !isSingboxAlive() {
					fmt.Println("Keep Alive triggered: Sing-box went offline unexpectedly. Restarting...")
					startProxy()
				}
			case <-fastCheckCh:
				for i := 0; i < 4; i++ {
					time.Sleep(500 * time.Millisecond)
					checkStatusAndUpdateUI()
				}
			}
		}
	}()
}

func checkStatusAndUpdateUI() {
	if isSingboxAlive() {
		isRunning = true
		if len(iconGreen) > 0 {
			systray.SetIcon(iconGreen)
		}
		mToggle.SetTitle("Stop Proxy")
		mToggle.Enable() 
	} else {
		isRunning = false
		if len(iconRed) > 0 {
			systray.SetIcon(iconRed)
		}
		mToggle.SetTitle("Start Proxy")
		mToggle.Enable()
	}
}

func openDashboardURL(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}
	_ = err
}