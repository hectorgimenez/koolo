package main

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/memory"
	"github.com/hectorgimenez/koolo/internal/town"
	"go.uber.org/zap"
	"log"
)

func main() {
	err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	logger := zap.NewNop()

	process, err := memory.NewProcess(logger)
	if err != nil {
		panic(err)
	}

	gr := memory.NewGameReader(process)

	bm := health.NewBeltManager(logger)
	//hm := health.NewHealthManager(logger, bm)
	sm := town.NewShopManager(logger, bm)
	char, err := character.BuildCharacter(logger)
	if err != nil {
		panic(err)
	}
	b := action.NewBuilder(logger, sm, bm, gr, char)
	a := b.ItemPickup(true, -1)

	gr.GetData(true)

	for err == nil {
		helper.Sleep(1000)
		data := gr.GetData(false)
		err = a.NextStep(logger, data)
	}
}
