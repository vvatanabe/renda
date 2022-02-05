package renda

import (
	"math"
	"sync"
	"time"
)

const (
	DefaultWorkers    = 10
	DefaultMaxWorkers = math.MaxUint64
)

func NewRenda(opts ...func(*Renda)) *Renda {
	r := &Renda{
		stopch:     make(chan struct{}),
		workers:    DefaultWorkers,
		maxWorkers: DefaultMaxWorkers,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func Workers(n uint64) func(*Renda) {
	return func(a *Renda) { a.workers = n }
}

func MaxWorkers(n uint64) func(*Renda) {
	return func(a *Renda) { a.maxWorkers = n }
}

type Renda struct {
	stopch     chan struct{}
	workers    uint64
	maxWorkers uint64
	seqmu      sync.Mutex
	seq        uint64
	began      time.Time
}

func (r *Renda) Start(f func() (interface{}, error), rate *Rate, du time.Duration) <-chan *Result {

	r.began = time.Now()

	var wg sync.WaitGroup

	workers := r.workers
	if workers > r.maxWorkers {
		workers = r.maxWorkers
	}

	results := make(chan *Result)
	ticks := make(chan struct{})
	for i := uint64(0); i < workers; i++ {
		wg.Add(1)
		go r.attack(f, &wg, ticks, results)
	}

	go func() {
		defer close(results)
		defer wg.Wait()
		defer close(ticks)

		began, count := time.Now(), uint64(0)
		for {
			elapsed := time.Since(began)
			if du > 0 && elapsed > du {
				return
			}

			wait, stop := rate.Pace(elapsed, count)
			if stop {
				return
			}

			time.Sleep(wait)

			if workers < r.maxWorkers {
				select {
				case ticks <- struct{}{}:
					count++
					continue
				case <-r.stopch:
					return
				default:
					workers++
					wg.Add(1)
					go r.attack(f, &wg, ticks, results)
				}
			}

			select {
			case ticks <- struct{}{}:
				count++
			case <-r.stopch:
				return
			}

		}

	}()

	return results
}

func (r *Renda) Stop() {
	select {
	case <-r.stopch:
		return
	default:
		close(r.stopch)
	}
}

func (r *Renda) attack(f func() (interface{}, error), workers *sync.WaitGroup, ticks <-chan struct{}, results chan<- *Result) {
	defer workers.Done()
	for range ticks {
		results <- r.hit(f)
	}
}

func (r *Renda) hit(f func() (interface{}, error)) *Result {
	var (
		res Result
		val interface{}
		err error
	)

	r.seqmu.Lock()
	res.Timestamp = r.began.Add(time.Since(r.began))
	res.Seq = r.seq
	r.seq++
	r.seqmu.Unlock()

	defer func() {
		res.Latency = time.Since(res.Timestamp)
		if val != nil {
			res.Value = val
		}
		if err != nil {
			res.Error = err
		}
	}()

	val, err = f()
	if err != nil {
		return &res
	}
	return &res
}

type Result struct {
	Seq       uint64        `json:"seq"`
	Timestamp time.Time     `json:"timestamp"`
	Latency   time.Duration `json:"latency"`
	Value     interface{}   `json:"value"`
	Error     error         `json:"error"`
}
