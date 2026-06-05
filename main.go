package main

import (	
	"github.com/energye/systray"
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