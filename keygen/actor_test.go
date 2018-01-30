package keygen

import (
	"testing"
	"github.com/aquiladev/monday/pool"
	"github.com/aquiladev/monday/storage"
)

func BenchmarkActorGenerate(b *testing.B) {
	config := mockConfig()
	a := NewActor("", &mockPool{})

	for i := 0; i < b.N; i++ {
		a.generate(config)
	}
}

func mockConfig() *RangeConfig {
	return &RangeConfig{
		Range: Range{
			From: "100152338825365595862742132647329357860924845607696427128849574371092698716",
			To:   "100152338825365595862742132647329357860924845607696427128849574371092699715",
		},
	}
}

type mockPool struct{}

var _ pool.Pool = (*mockPool)(nil)

func (q *mockPool) Put(msg *storage.Message) error {
	return nil
}

func (q *mockPool) Pop() ([]*storage.Message, error) {
	return nil, nil
}
