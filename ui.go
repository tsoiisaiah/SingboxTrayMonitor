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
	mDashboard   *systray.MenuItem
	dashboardEnabled bool
	isRunning  bool
	iconGreen  []byte
	iconRed   []byte
	
	isProcessing bool 

	// Submenu items under Settings
	mAutoConnect  *systray.MenuItem
	mKeepAlive        *systray.MenuItem
	keepAliveRetryCount int
	mStartWithWin *systray.MenuItem
)

func setupUI() {
	fmt.Println("systray.onReady started")

	fastCheckCh = make(chan bool, 1)

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
	dashboardEnabled = checkDashboardEnabled()
	if !dashboardEnabled {
		mDashboard.Disable()
		mDashboard.SetTooltip("Dashboard is disabled in your config.json")
	}

	mSettings := systray.AddMenuItem("Settings", "Configure tray options")
	mAutoConnect = mSettings.AddSubMenuItemCheckbox("Auto Connect", "Automatically connect sing-box on start starts", autoConnect)
	mKeepAlive = mSettings.AddSubMenuItemCheckbox("Keep Alive", "Restart sing-box automatically if it crashes", keepAlive)
	mStartWithWin = mSettings.AddSubMenuItemCheckbox("Start with Windows", "Launch automatically on Windows startup", startWithWindows)

	mTool := systray.AddMenuItem("Tool", "Utility debugging tools")
	mResetSysProxy := mTool.AddSubMenuItem("Reset System Proxy", "Force turn off windows system proxy settings")

	systray.AddSeparator()

	mVersion := systray.AddMenuItem("Version: "+appVersion, "Click to visit release page")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Exit", "Stop core and exit application")

	mToggle.Click(func() {
		isProcessing = true
		mToggle.Disable()

		if !isRunning {
			mToggle.SetTitle("Starting...")
			keepAliveRetryCount = 0
			startProxy()
		} else {
			mToggle.SetTitle("Stopping...")
			stopProxy()
		}

		loadAndSyncConfig()

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
	
	mResetSysProxy.Click(func() {
		if isSingboxAlive() {
			stopProxy()
		}
		resetSystemProxy()
	})
	
	mVersion.Click(func() {
		openDashboardURL(homepageURL)
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
					if keepAliveRetryCount < 0 {
						// Do nothing, stop retrying
					} else if keepAliveRetryCount < 5 {
						// 2. Retry logic (0 to 4)
						keepAliveRetryCount++
						fmt.Printf("Keep Alive triggered: Attempt %d/5. Restarting...\n", keepAliveRetryCount)
						startProxy()
					} else {
						// 3. Abort state (counter == 5)
						go triggerSystemPopup(
							"Keep Alive Error",
							fmt.Sprintf("Failed to restart Sing-box after %d consecutive attempts. Please check your configuration or try starting manually.", keepAliveRetryCount),
						)
						keepAliveRetryCount = -1 // Mark as aborted
					}
				}
			case <-fastCheckCh:
				for i := 0; i < 4; i++ {
					time.Sleep(500 * time.Millisecond)
					
					// If the background target state has already settled early, break out to release UI instantly
					alive := isSingboxAlive()
					if (isProcessing && !isRunning && alive) || (isProcessing && isRunning && !alive) {
						syncUIState(alive)
						break
					}
					
					syncUIState(alive)
				}
			}
		}
	}()
}

func checkStatusAndUpdateUI() {
	if isProcessing {
		return
	}

	// Launch an isolated, lightweight concurrent Goroutine task handle to drop UI blocking thread
	go func() {
		alive := isSingboxAlive()
		syncUIState(alive)
	}()
}

func syncUIState(alive bool) {
	if alive {
		isRunning = true
		keepAliveRetryCount = 0 // Reset counter when proxy successfully comes back online
		if len(iconGreen) > 0 {
			systray.SetIcon(iconGreen)
		}
		mToggle.SetTitle("Stop Proxy")
		isProcessing = false
		mToggle.Enable()
		if dashboardEnabled {
			mDashboard.Enable()
		}
	} else {
		isRunning = false
		if len(iconRed) > 0 {
			systray.SetIcon(iconRed)
		}
		mToggle.SetTitle("Start Proxy")
		isProcessing = false
		mToggle.Enable()
		if dashboardEnabled {
			mDashboard.Disable()
		}
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