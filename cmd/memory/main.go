package main

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/memory"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/reader"
	"log"
	"time"
)

func main() {
	err := config.Load()
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
	for true {
		d := gr.GetData(false)
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
