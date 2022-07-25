package mnc

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type io struct {
	input  interface{}
	output chan interface{}
}

type mnc struct {
	batch   int
	handler Handler
	io      chan io
}

type Handler func(ctx context.Context, inputs []interface{}) (outputs []interface{})

type Options func(m *mnc)

// Init initiate mnc
func Init(h Handler, opts ...Options) (m mnc, err error) {
	if h == nil {
		err = errors.New("nil handler")
		return
	}

	m = mnc{
		batch:   1,
		handler: h,
		io:      make(chan io),
	}

	for _, apply := range opts {
		apply(&m)
	}

	return
}

// WithBatch define batch size
func WithBatch(batch int) (o Options) {
	return func(m *mnc) {
		m.batch = batch
	}
}

// Do wrap your input into an io. It ggenerate output channel,
// wrap your input and the output channel into a struct, then
// pass the io to be handled by handler
func (m *mnc) Do(input interface{}) (res interface{}) {
	chOut := make(chan interface{})

	io := io{
		input:  input,
		output: chOut,
	}

	m.io <- io
	res = <-chOut

	return
}

// Run a background process that receive input from io channel and pass
// received io into handler
func (m *mnc) Run() {
	var inputs []interface{}
	var outputs []chan interface{}
	var ticker *time.Ticker = &time.Ticker{}
	var do bool

	for {
		fmt.Println("ready!")

		// waiting for input
		select {

		// got io input
		case io := <-m.io:
			fmt.Println("got io")
			inputs = append(inputs, io.input)
			outputs = append(outputs, io.output)

			if len(inputs) == 1 {
				// wait for 2 second for the batch size to be filled.
				// if 2 second passed and input batch size is not
				// filled, ticker will be used to trigger handler to
				// handle existing inputs
				ticker = time.NewTicker(2 * time.Second)
			}

		// got ticker
		case t := <-ticker.C:
			fmt.Println("got t at", t)
			do = true
		}

		// if len of inputs equal batch size,
		// or if the process already wait long enough,
		// handle the request
		if len(inputs) == m.batch || do {
			go m.handle(inputs, outputs)
			inputs = []interface{}{}
			outputs = []chan interface{}{}

			do = false
			ticker.Stop()
		}
	}
}

func (m *mnc) handle(inputs []interface{}, outputs []chan interface{}) {
	fmt.Println("handle!")

	if len(inputs) != len(outputs) {
		return
	}

	ctx := context.Background()
	// execute handler
	res := m.handler(ctx, inputs)

	if len(res) != len(outputs) {
		return
	}

	for i, r := range res {
		outputs[i] <- r
	}
}
