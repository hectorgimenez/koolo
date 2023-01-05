package main

import (
	"fmt"
	zapLogger "github.com/hectorgimenez/koolo/cmd/koolo/log"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/memory"
	"log"
	"time"
)

func main() {
	err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}
	logger, err := zapLogger.NewLogger(config.Config.Debug.Log, config.Config.LogFilePath)
	process, err := memory.NewProcess(logger)
	if err != nil {
		panic(err)
	}

	gd := memory.NewGameReader(process)

	start := time.Now()
	for true {
		d := gd.GetData(true)
		fmt.Println(d.MercHPPercent())
		time.Sleep(time.Second)
	}

	fmt.Printf("Read time: %dms\n", time.Since(start).Milliseconds())
}
