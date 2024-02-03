package main

import (
	"log"
	"syscall"

	"github.com/hectorgimenez/d2go/pkg/memory"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/reader"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/lxn/win"
	"go.uber.org/zap"
)

func main() {
	err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	logger, _ := zap.NewDevelopment()

	ptr, err := syscall.UTF16PtrFromString("Diablo II: Resurrected")
	if err != nil {
		log.Fatalf(err.Error())
	}
	window := win.FindWindow(nil, ptr)

	pos := win.WINDOWPLACEMENT{}
	point := win.POINT{}
	win.ClientToScreen(window, &point)
	win.GetWindowPlacement(window, &pos)

	hid.WindowLeftX = int(point.X + 1)
	hid.WindowTopY = int(point.Y)
	hid.GameAreaSizeX = int(pos.RcNormalPosition.Right) - hid.WindowLeftX - 9
	hid.GameAreaSizeY = int(pos.RcNormalPosition.Bottom) - hid.WindowTopY - 9
	helper.Sleep(1000)

	process, err := memory.NewProcess()
	if err != nil {
		panic(err)
	}

	gd := memory.NewGameReader(process)
	gr := &reader.GameReader{GameReader: gd}
	bm := health.NewBeltManager(logger)
	sm := town.NewShopManager(logger, bm)
	char, err := character.BuildCharacter(logger)
	if err != nil {
		panic(err)
	}

	b := action.NewBuilder(logger, sm, bm, gr, char)
	a := b.ItemPickup(true, -1)

	gr.GetData(true)

	for err == nil {
		data := gr.GetData(false)
		err = a.NextStep(logger, data)
	}
}
