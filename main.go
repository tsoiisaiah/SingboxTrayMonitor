package main

import (
	"os"
	"github.com/energye/systray"
)

var (
	baseDir string
)

func init() {
	baseDir, _ = os.Getwd()
}

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