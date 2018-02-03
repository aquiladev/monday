package writer

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/aquiladev/monday/database"
	"github.com/aquiladev/monday/database/policy"
	"github.com/aquiladev/monday/pool"
	"github.com/aquiladev/monday/storage"
)

type Actor struct {
	started  int32
	shutdown int32
	quit     chan struct{}
	wg       sync.WaitGroup

	handledLogMsg int64
	lastLogTime   time.Time

	pool     pool.Pool
	db       database.DB
	policies []policy.StoragePolicy
}

func (a *Actor) start() {
out:
	for {
		select {
		case <-a.quit:
			a.logProgress(true)
			break out
		default:
		}

		if err := a.handleMessages(); err != nil {
			log.Error(err)
		}
		a.logProgress(false)
	}

	a.wg.Done()
}

func (a *Actor) handleMessages() error {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()

	if err := a.checkPolicies(); err != nil {
		return err
	}

	messages, err := a.pool.Pop()
	if err != nil || len(messages) == 0 {
		return err
	}
	log.Tracef("Retrieved %d messages", len(messages))

	ch := make(chan error)
	defer close(ch)

	for _, m := range messages {
		go func(msg *storage.Message, errCh chan error) {
			log.Tracef("Retrieved %d keys", len(msg.Keys))

			list := make([]*database.KeyValue, len(msg.Keys))
			for i, k := range msg.Keys {
				list[i] = &database.KeyValue{
					K: []byte(k.PublicKey),
					V: []byte(k.PrivateKey),
				}
			}

			errCh <- a.db.PutBatch(list)
		}(m, ch)
	}

	for i := 0; i < len(messages); i++ {
		if resErr := <-ch; resErr != nil {
			err = resErr
			continue
		}
		a.handledLogMsg++
	}

	return err
}

func (a *Actor) checkPolicies() error {
	for _, p := range a.policies {
		_, err := p.IsAccept()
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Actor) logProgress(force bool) {
	now := time.Now()
	duration := now.Sub(a.lastLogTime)
	if duration < time.Second*time.Duration(20) && !force {
		return
	}

	// Truncate the duration to 10s of milliseconds.
	durationMillis := int64(duration / time.Millisecond)
	tDuration := 10 * time.Millisecond * time.Duration(durationMillis/10)

	// Log information about messages.
	messageStr := "messages"
	if a.handledLogMsg == 1 {
		messageStr = "message"
	}

	log.Infof("Handled %d %s in the last %s", a.handledLogMsg, messageStr, tDuration)

	a.handledLogMsg = 0
	a.lastLogTime = now
}

func (a *Actor) Start() {
	// Already started?
	if atomic.AddInt32(&a.started, 1) != 1 {
		return
	}

	log.Trace("Starting writer")
	a.wg.Add(1)
	go a.start()
}

func (a *Actor) Stop() {
	if atomic.AddInt32(&a.shutdown, 1) != 1 {
		log.Warnf("Writer is already in the process of shutting down")
	}

	log.Infof("Writer shutting down")
	close(a.quit)
	a.wg.Wait()
}

func (a *Actor) WaitForShutdown() {
	a.wg.Wait()
}

func NewActor(pool pool.Pool, db database.DB, policies []policy.StoragePolicy) *Actor {
	return &Actor{
		pool:        pool,
		db:          db,
		policies:    policies,
		quit:        make(chan struct{}),
		lastLogTime: time.Now(),
	}
}
