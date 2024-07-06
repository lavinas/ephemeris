package main

import (
	"context"
	"fmt"
	"time"
)

func numeros(v chan <- int) {
	for i := 0; i < 10; i++ {
		v <- i
		fmt.Println("Enviando: ", i)
	}
	close(v)
}

/*
func cancel(cf context.CancelFunc) {
	time.Sleep(5 * time.Second)
	cf()
}
*/

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	c := make(chan int)
	go numeros(c)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Terminando")
			return
		case v, ok := <-c:
			if ok {
				fmt.Println("Recibiendo: ", v)
				time.Sleep(2 * time.Second)
			}
		}
	}
}
