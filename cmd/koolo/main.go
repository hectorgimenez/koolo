package main

import (
	"context"
	"fmt"
	zapLogger "github.com/hectorgimenez/koolo/cmd/koolo/log"
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/kbinani/screenshot"
	"gocv.io/x/gocv"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	logger, err := zapLogger.NewLogger(cfg.Debug, cfg.LogFilePath)
	if err != nil {
		log.Fatalf("Error starting logger: %s", err.Error())
	}
	defer logger.Sync()

	supervisor := koolo.NewSupervisor(cfg)

	ctx := context.Background()
	err = supervisor.Start(ctx)
	tf, err := koolo.NewTemplateFinder(logger, "assets/templates")
	img, err := screenshot.CaptureDisplay(0)
	mat, _ := gocv.ImageToMatRGB(img)
	tl := tf.Find("IDEA", mat)
	fmt.Println(tl)
	if err != nil {
		log.Fatalf("Error running Koolo: %s", err.Error())
	}
}
