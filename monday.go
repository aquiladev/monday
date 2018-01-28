package main

import (
	"github.com/aquiladev/monday/database"
	"github.com/aquiladev/monday/database/policy"
	"github.com/aquiladev/monday/storage"
	"github.com/aquiladev/monday/pool"
)

var (
	cfg *config
)

func mondayMain(workerChan chan<- *worker) error {
	mcfg, _, err := loadConfig()
	if err != nil {
		log.Error(err)
		return err
	}
	log.Infof("Config %+v", mcfg)

	cfg = mcfg
	defer func() {
		if logRotator != nil {
			logRotator.Close()
		}
	}()

	// Get a channel that will be closed when a shutdown signal has been
	// triggered either from an OS signal such as SIGINT (Ctrl+C) or from
	// another subsystem such as the RPC server.
	interrupt := interruptListener()
	defer log.Info("Shutdown complete")

	// Show version at startup.
	log.Infof("Version %s", version())

	// Return now if an interrupt signal was triggered.
	if interruptRequested(interrupt) {
		return nil
	}

	var db database.DB = nil
	policies := make([]policy.StoragePolicy, 0)
	if cfg.KeepLocal {
		// Create local database
		var err error
		db, err = database.NewDB(cfg.DataDir)
		if err != nil {
			log.Errorf("%v", err)
			return err
		}
		defer func() {
			// Ensure the database is sync'd and closed on shutdown.
			log.Infof("Gracefully shutting down the database...")
			db.Close()
		}()

		// Create storage policies
		policies = append(policies, policy.NewMaxDbSize(cfg.DataDir, cfg.MaxDbSize))
		policies = append(policies, policy.NewMaxDiskUsage(cfg.DataDir, cfg.MaxDiskUsage))
	}

	// Create pool
	blobRepo, err := storage.NewAzureBlobRepository(
		cfg.StorageAccountName,
		cfg.StorageAccountKey,
		cfg.StorageBlobName)
	if err != nil {
		return err
	}

	memQueue := pool.NewMemQueue(cfg.MemPoolCapacity)
	keyPool := pool.NewMultiLevelPool(memQueue, blobRepo, cfg.NumOfMessages, cfg.KeepLocal)

	// Create worker and start it.
	worker := newWorker(cfg, db, keyPool, policies)
	defer func() {
		log.Infof("Gracefully shutting down the worker...")
		worker.Stop()
		worker.WaitForShutdown()
		log.Infof("Worker shutdown complete")
	}()
	worker.Start()
	if workerChan != nil {
		workerChan <- worker
	}

	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interrupt
	return nil
}
