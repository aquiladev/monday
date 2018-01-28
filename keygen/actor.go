package keygen

import (
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aquiladev/monday/pool"
	"github.com/aquiladev/monday/storage"
	"github.com/aquiladev/monday/util"
)

type Actor struct {
	started  int32
	shutdown int32
	quit     chan struct{}
	wg       sync.WaitGroup

	handledLogPg int64
	lastLogTime  time.Time

	rangeUrl string
	pool     pool.Pool
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

		config, err := fetchRange(a.rangeUrl)
		if err != nil ||
			len(config.Range.From) == 0 ||
			len(config.Range.To) == 0 {
			log.Errorf("Error while fetching range %+v", err)

			time.Sleep(time.Duration(1000) * time.Millisecond)
			continue
		}

		log.Infof("Fetched range: %+v", config)
		a.generate(config)

		a.handledLogPg++
		a.logProgress(false)
	}

	a.wg.Done()
}

func (a *Actor) generate(config *RangeConfig) {
	from, _ := util.Parse(config.Range.From)
	to, _ := util.Parse(config.Range.To)
	to = to.Add(to, big.NewInt(1))
	pages := int(big.NewInt(0).Sub(to, from).Int64())

	var keys []*storage.KeyPair
	for pages > 0 {
		bucketSize := runtime.NumCPU()
		restPages := pages - bucketSize
		rangeFrom := big.NewInt(0).Sub(to, big.NewInt(int64(pages)))
		size := bucketSize

		if restPages < 0 {
			size = pages
		}

		bucketKeys := a.generateBucket(rangeFrom, size)
		keys = append(keys, bucketKeys...)
		pages = restPages
	}

	if err := a.pool.Put(&storage.Message{Keys: keys}); err != nil {
		log.Error(err)
	}
}

func (a *Actor) generateBucket(from *big.Int, amount int) []*storage.KeyPair {
	ch := make(chan bool)
	defer close(ch)

	var keys []*storage.KeyPair
	for i := 0; i < amount; i++ {
		page := big.NewInt(0).Add(from, big.NewInt(int64(i)))
		go func(done chan bool) {
			pageKeys := a.generatePage(page)
			keys = append(keys, pageKeys...)
			done <- true
		}(ch)
	}

	for i := 0; i < amount; i++ {
		<-ch
	}

	return keys
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

func NewActor(configUrl string, pool pool.Pool) *Actor {
	return &Actor{
		rangeUrl:    configUrl,
		pool:        pool,
		quit:        make(chan struct{}),
		lastLogTime: time.Now(),
	}
}
