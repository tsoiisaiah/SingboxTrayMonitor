package main

import (	
	"github.com/energye/systray"
)

var (
	appVersion = "dev"
	homepageURL  = "https://github.com/tsoiisaiah/SingboxTrayMonitor" 
)

func main() {
	MainRun()
}

func MainRun() {
	systray.Run(onReady, onExit)
}

func onReady() {
	setupUI();
}

func onExit() {
	processCleanUp();
}