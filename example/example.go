package main

import (
	"context"
	"fmt"
	"merge-and-conquer/mnc"
	"sync"
)

func main() {
	handler := func(ctx context.Context, inputs []interface{}) (outputs []interface{}) {
		nums := make([]int, len(inputs))

		for idx, input := range inputs {
			num, ok := input.(int)
			if ok {
				nums[idx] = num
			}
		}

		outputs = make([]interface{}, len(inputs))

		for idx, num := range nums {
			outputs[idx] = num * 2
		}
		return
	}

	m, _ := mnc.Init(handler, mnc.WithBatch(3))
	go m.Run()

	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		fmt.Println("try on", i)
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			res := m.Do(i)
			fmt.Println(i, res)
		}(i)
	}

	wg.Wait()
}
