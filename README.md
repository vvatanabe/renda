# Renda [![Go](https://github.com/vvatanabe/renda/actions/workflows/go.yml/badge.svg)](https://github.com/vvatanabe/renda/actions/workflows/go.yml)

Renda is a go library that repeatedly executes any processes.

The name "Renda" comes from the Japanese word "連打".

## Requires

- Go 1.17+

## Installation

This package can be installed with the go get command:

```
$ go get -u github.com/vvatanabe/renda
```

## Usage

### Basically

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/vvatanabe/renda"
)

func main() {
	workers := renda.Workers(10)
	maxWorkers := renda.MaxWorkers(20)

	hello := func() (interface{}, error) {
		return "Hello 連打", nil
	}
	rate := &renda.Rate{Freq: 3, Per: time.Second}
	duration := 5 * time.Second

	r := renda.NewRenda(workers, maxWorkers)
	res := r.Start(hello, rate, duration)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	for {
		select {
		case <-sig:
			r.Stop()
			return
		case r, ok := <-res:
			if !ok {
				return
			}
			v, _ := json.Marshal(r)
			fmt.Println(string(v))
		}
	}
}
```

## Acknowledgments

[github.com/tsenart/vegeta](https://github.com/tsenart/vegeta)

## Bugs and Feedback

For bugs, questions and discussions please use the GitHub Issues.
