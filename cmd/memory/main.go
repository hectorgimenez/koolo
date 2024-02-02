package main

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/helper"
	memory2 "github.com/hectorgimenez/koolo/internal/memory"
	"log"
	"syscall"
	"time"

	"github.com/hectorgimenez/d2go/pkg/memory"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/reader"
	"github.com/lxn/win"
)

func main() {
	ptr, err := syscall.UTF16PtrFromString("Diablo II: Resurrected")
	if err != nil {
		panic(err)
	}
	window := win.FindWindow(nil, ptr)
	if window == win.HWND_TOP {
		panic("Diablo II: Resurrected window not found")
	}
	memory2.HWND = window

	pos := win.WINDOWPLACEMENT{}
	point := win.POINT{}
	win.ClientToScreen(window, &point)
	win.GetWindowPlacement(window, &pos)

	hid.WindowLeftX = int(point.X)
	hid.WindowTopY = int(point.Y)
	hid.GameAreaSizeX = int(pos.RcNormalPosition.Right) - hid.WindowLeftX - 9
	hid.GameAreaSizeY = int(pos.RcNormalPosition.Bottom) - hid.WindowTopY - 9

	helper.Screenshot()

	err = config.Load()
	config.Config.Debug.RenderMap = true
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}
	process, err := memory.NewProcess()
	if err != nil {
		panic(err)
	}

	gd := memory.NewGameReader(process)
	gr := reader.GameReader{
		GameReader: gd,
	}

	start := time.Now()
	gr.GetData(true)
	if err != nil {
		panic(err)
	}

	for true {
		d := gr.GetData(false)

		d.Roster.FindByName("Ayuso")
		fmt.Println(d.PlayerUnit.HPPercent())
		time.Sleep(time.Millisecond * 500)
	}

	fmt.Printf("Read time: %dms\n", time.Since(start).Milliseconds())
}
