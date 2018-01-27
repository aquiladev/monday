package main

import (
	"sync/atomic"

	"github.com/aquiladev/monday/database"
	"github.com/aquiladev/monday/database/policy"
	"github.com/aquiladev/monday/keygen"
	"github.com/aquiladev/monday/pool"
)

type worker struct {
	started  int32
	shutdown int32
	quit     chan struct{}

	keyGenActor *keygen.Actor
}

func (w *worker) Start() {
	// Already started?
	if atomic.AddInt32(&w.started, 1) != 1 {
		return
	}

	workerLog.Trace("Starting worker")

	w.keyGenActor.Start()
}

func (w *worker) Stop() error {
	// Make sure this only happens once.
	if atomic.AddInt32(&w.shutdown, 1) != 1 {
		workerLog.Info("Worker is already in the process of shutting down")
		return nil
	}

	workerLog.Warn("Worker shutting down")

	go w.keyGenActor.Stop()

	// Signal the remaining goroutines to quit.
	close(w.quit)
	return nil
}

func (w *worker) WaitForShutdown() {
	w.keyGenActor.WaitForShutdown()
}

func newWorker(cfg *config, db database.DB, pool pool.Pool, policies []policy.StoragePolicy) *worker {
	return &worker{
		quit:        make(chan struct{}),
		keyGenActor: keygen.NewActor(cfg.RangeUrl, pool, cfg.KeepLocal, db, policies),
	}
}
