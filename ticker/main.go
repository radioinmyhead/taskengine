package main

import (
	"fmt"
	"time"
)

func main() {
	// worker
	var ticker *time.Ticker
	ticker = time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			fmt.Println("update time", time.Now())
		}
	}()

	time.Sleep(time.Second * 5)
}
