package keygen

import (
	"math/big"
	"sync/atomic"
	"sync"
	"time"

	"github.com/aquiladev/monday/database"
	"github.com/aquiladev/monday/database/policy"
	"github.com/aquiladev/monday/pool"
	"github.com/aquiladev/monday/util"
	"github.com/aquiladev/monday/storage"
)

type Actor struct {
	started  int32
	shutdown int32
	quit     chan struct{}
	wg       sync.WaitGroup

	handledLogPg int64
	lastLogTime  time.Time

	configUrl string
	pool      pool.Pool
	keepLocal bool
	db        database.DB
	policies  []policy.StoragePolicy
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

		config, err := fetchRange(a.configUrl)
		if err != nil ||
			len(config.Range.From) == 0 ||
			len(config.Range.To) == 0 {
			log.Errorf("Error while fetching config %+v", err)

			time.Sleep(time.Duration(1000) * time.Millisecond)
			continue
		}

		log.Infof("Fetched range: %+v", config)
		if err := a.generate(config); err != nil {
			log.Error(err)
		}

		a.handledLogPg++
		a.logProgress(false)
	}

	a.wg.Done()
}

func (a *Actor) generate(config *RangeConfig) error {
	from, _ := util.Parse(config.Range.From)
	to, _ := util.Parse(config.Range.To)
	to = to.Add(to, big.NewInt(1))

	pages := int(big.NewInt(0).Sub(to, from).Int64())

	var keys []*storage.KeyPair
	for i := 0; i < pages; i++ {
		page := big.NewInt(0).Add(from, big.NewInt(int64(i)))
		pageKeys := a.generatePage(page)
		keys = append(keys, pageKeys...)
	}

	if a.isAcceptPolicies() {
		list := make([]*database.KeyValue, len(keys))
		for i, k := range keys {
			list[i] = &database.KeyValue{
				K: []byte(k.PublicKey),
				V: []byte(k.PrivateKey),
			}
		}

		err := a.db.PutBatch(list)
		if err == nil {
			return nil
		}
		log.Error(err)
	}

	return a.queueKeys(keys)
}

func (a *Actor) generatePage(page *big.Int) []*storage.KeyPair {
	pageSize := 128
	pageD := big.NewInt(0).Mul(page, big.NewInt(int64(pageSize)))

	var keys []*storage.KeyPair
	for i := 0; i < pageSize; i++ {
		d := big.NewInt(0).Add(pageD, big.NewInt(int64(i)))
		key := generate(d)

		keys = append(keys, &storage.KeyPair{
			PrivateKey: key.PrivKey,
			PublicKey:  key.PubKey,
		})
		keys = append(keys, &storage.KeyPair{
			PrivateKey: key.PrivKey,
			PublicKey:  key.CompressedPubKey,
		})
	}

	return keys
}

func (a *Actor) isAcceptPolicies() bool {
	if a.keepLocal {
		policyAccepted := true
		for _, p := range a.policies {
			accept, err := p.IsAccept()
			if err != nil {
				log.Error(err)
			}

			policyAccepted = policyAccepted && accept
		}

		return policyAccepted
	}

	return false
}

func (a *Actor) queueKeys(keys []*storage.KeyPair) error {
	return a.pool.Put(&storage.Message{
		Keys:             keys,
	})
}

func (a *Actor) logProgress(force bool) {
	now := time.Now()
	duration := now.Sub(a.lastLogTime)
	if duration < time.Second*time.Duration(10) && !force {
		return
	}

	// Truncate the duration to 10s of milliseconds.
	durationMillis := int64(duration / time.Millisecond)
	tDuration := 10 * time.Millisecond * time.Duration(durationMillis/10)

	// Log information about pages.
	pageStr := "batches"
	if a.handledLogPg == 1 {
		pageStr = "batch"
	}

	log.Infof("Handled %d %s in the last %s",
		a.handledLogPg, pageStr, tDuration)

	a.handledLogPg = 0
	a.lastLogTime = now
}

func (a *Actor) Start() {
	// Already started?
	if atomic.AddInt32(&a.started, 1) != 1 {
		return
	}

	log.Trace("Starting generator")
	a.wg.Add(1)
	go a.start()
}

func (a *Actor) Stop() {
	if atomic.AddInt32(&a.shutdown, 1) != 1 {
		log.Warnf("Generator is already in the process of shutting down")
		return
	}

	log.Infof("Generator shutting down")
	close(a.quit)
	a.wg.Wait()
}

func (a *Actor) WaitForShutdown() {
	a.wg.Wait()
}

func NewActor(
	configUrl string,
	pool pool.Pool,
	keepLocal bool,
	db database.DB,
	policies []policy.StoragePolicy) *Actor {

	return &Actor{
		configUrl:   configUrl,
		pool:        pool,
		keepLocal:   keepLocal,
		db:          db,
		policies:    policies,
		quit:        make(chan struct{}),
		lastLogTime: time.Now(),
	}
}
