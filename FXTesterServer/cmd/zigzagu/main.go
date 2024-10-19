package main

import (
	"fmt"
	"sync"
)

func main() {

	a := 0

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			fmt.Println("call done")
			wg.Done()
		}()

		if a == 0 {
			panic("test!!")
		}
	}()

	wg.Wait()

	fmt.Println("finish")

}
