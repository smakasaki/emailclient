package main

import (
	"emailclient/pkg/gui"
	"os"
)

func main() {
	os.Setenv("FYNE_THEME", "light")
	gui.RunAppGUI()

}
