package main

import (
	"sync/atomic"

	"github.com/aquiladev/monday/database"
	"github.com/aquiladev/monday/database/policy"
	"github.com/aquiladev/monday/keygen"
	"github.com/aquiladev/monday/pool"
	"github.com/aquiladev/monday/writer"
)

type worker struct {
	started  int32
	shutdown int32
	quit     chan struct{}

	keyGenActor *keygen.Actor
	writeActor  *writer.Actor
}

func (w *worker) Start() {
	// Already started?
	if atomic.AddInt32(&w.started, 1) != 1 {
		return
	}

	workerLog.Trace("Starting worker")

	if cfg.Generating {
		w.keyGenActor.Start()
	}

	if cfg.KeepLocal {
		w.writeActor.Start()
	}
}

func (w *worker) Stop() error {
	// Make sure this only happens once.
	if atomic.AddInt32(&w.shutdown, 1) != 1 {
		workerLog.Info("Worker is already in the process of shutting down")
		return nil
	}

	workerLog.Warn("Worker shutting down")

	if cfg.Generating {
		go w.keyGenActor.Stop()
	}

	if cfg.KeepLocal {
		go w.writeActor.Stop()
	}

	// Signal the remaining goroutines to quit.
	close(w.quit)
	return nil
}

func (w *worker) WaitForShutdown() {
	if cfg.Generating {
		w.keyGenActor.WaitForShutdown()
	}

	if cfg.KeepLocal {
		w.writeActor.WaitForShutdown()
	}
}

func newWorker(cfg *config, db database.DB, pool pool.Pool, policies []policy.StoragePolicy) *worker {
	return &worker{
		quit:        make(chan struct{}),
		keyGenActor: keygen.NewActor(cfg.RangeUrl, pool),
		writeActor:  writer.NewActor(pool, db, policies),
	}
}
