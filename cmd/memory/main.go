package main

import (
	"fmt"
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
	win.SetForegroundWindow(window)
	memory2.HWND = window

	pos := win.WINDOWPLACEMENT{}
	point := win.POINT{}
	win.ClientToScreen(window, &point)
	win.GetWindowPlacement(window, &pos)

	hid.WindowLeftX = int(point.X)
	hid.WindowTopY = int(point.Y)
	hid.GameAreaSizeX = int(pos.RcNormalPosition.Right) - hid.WindowLeftX - 9
	hid.GameAreaSizeY = int(pos.RcNormalPosition.Bottom) - hid.WindowTopY - 9

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

		//screenshot := helper.Screenshot()
		//matScreen, _ := gocv.ImageToMatRGB(screenshot)
		//
		//skillMat := gocv.IMRead("assets/skills/blizzard.png", gocv.IMReadUnchanged)
		//
		//rgb := skillMat.Clone()
		//gocv.CvtColor(skillMat, &rgb, gocv.ColorBGRAToBGR)
		//
		//res := gocv.NewMat()
		//gocv.MatchTemplate(matScreen, rgb, &res, gocv.TmCcoeffNormed, gocv.NewMat())
		//
		//_, maxVal, _, maxPos := gocv.MinMaxLoc(res)

		//rs := gcv.FindAllImg(img1, skill, 0.4)

		//fmt.Println(res, maxVal, maxPos)
		//f, _ := os.Create("data.bin")
		//enc := gob.NewEncoder(f)
		//err := enc.Encode(&d)
		//fmt.Println(err)
		//f.Close()
		d.Roster.FindByName("Ayuso")
		fmt.Println(d.PlayerUnit.HPPercent())
		time.Sleep(time.Millisecond * 500)
	}

	fmt.Printf("Read time: %dms\n", time.Since(start).Milliseconds())
}
