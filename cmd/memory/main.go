package main

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/memory"
	"time"
)

func main() {
	process, err := memory.NewProcess()
	if err != nil {
		panic(err)
	}

	gd := memory.NewGameReader(process)

	start := time.Now()
	d := gd.GetData(true)
	fmt.Println(d)
	fmt.Printf("Read time: %dms\n", time.Since(start).Milliseconds())
}
