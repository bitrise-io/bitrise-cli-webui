package main

import (
	"fmt"
	"time"
)

func main() {
	d := make(chan int, 1)

	fmt.Println("Starting long process\n")
	i := 0
	ticker := time.NewTicker(time.Millisecond * 500)
	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Printf(".")
				i++
				if i > 30 {
					ticker.Stop()
					d <- 1
				}

			}
		}
	}()
	<-d
	fmt.Println("\n", "End")
}
