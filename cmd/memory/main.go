package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/hectorgimenez/d2go/pkg/memory"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/reader"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/lxn/win"
	"go.uber.org/zap"
)

func main() {
	window := robotgo.FindWindow("Diablo II: Resurrected")
	if window == win.HWND_TOP {
		panic("Diablo II: Resurrected window not found")
	}
	win.SetForegroundWindow(window)

	// Exclude border offsets
	// TODO: Improve this, maybe getting window content coordinates?
	pos := win.WINDOWPLACEMENT{}
	win.GetWindowPlacement(window, &pos)
	hid.WindowLeftX = int(pos.RcNormalPosition.Left) + 8
	hid.WindowTopY = int(pos.RcNormalPosition.Top) + 31
	hid.GameAreaSizeX = int(pos.RcNormalPosition.Right) - hid.WindowLeftX - 10
	hid.GameAreaSizeY = int(pos.RcNormalPosition.Bottom) - hid.WindowTopY - 10

	err := config.Load()
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
	logger, _ := zap.NewDevelopment()
	tf, err := ui.NewTemplateFinder(logger, "assets")
	if err != nil {
		panic(err)
	}

	for true {
		d := gr.GetData(false)

		tf.Find("59", helper.Screenshot())
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
