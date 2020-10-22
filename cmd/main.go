package main

import (
	"log"
	"strconv"
)

func main() {

	ch := tick()
	for i := range ch {
		log.Print(i)
	}
}
func tick() <-chan string {
	ch := make(chan string)
	go func() {
		for i := 0; i < 10; i++ {
			ch <- strconv.FormatInt(int64(i), 10)
		}
		close(ch)
	}()
	return ch
}
