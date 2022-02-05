package renda

import (
	"testing"
	"time"
)

func TestRendaRate(t *testing.T) {
	t.Parallel()
	rate := &Rate{Freq: 100, Per: time.Second}
	r := NewRenda()
	var hits uint64
	for _ = range r.Start(func() (interface{}, error) {
		hits++
		return nil, nil
	}, rate, 1*time.Second) {

	}
	if got, want := hits, uint64(rate.Freq); got != want {
		t.Fatalf("got: %v, want: %v", got, want)
	}
}
