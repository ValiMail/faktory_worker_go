package faktory_worker

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	faktory "github.com/contribsys/faktory/client"
)

// RunPipelined2 starts processing jobs, using one pool of goroutines to fetch work, and another to execute it.
// This method does not return.
func (mgr *Manager) RunPipelined2() {
	// This will signal to Faktory that all connections from this process
	// are worker connections.
	rand.Seed(time.Now().UnixNano())
	faktory.RandomProcessWid = strconv.FormatInt(rand.Int63(), 32)

	if mgr.Pool == nil {
		// Configure max connection pool size to number of fetchers + number of result-reporters + 1 for
		// the heartbeat process
		pool, err := NewChannelPool(mgr.Dispatchers+mgr.Concurrency+1, mgr.Dispatchers+mgr.Concurrency+1, func() (Closeable, error) { return faktory.Open() })
		if err != nil {
			panic(err)
		}
		mgr.Pool = pool
	}

	// a channel to distribute jobs to worker goroutines.  unbuffered to provide backpressure.
	in := make(chan *faktory.Job)

	mgr.fireEvent(Startup)

	go heartbeat(mgr)

	// start a set of goroutines to report results to the server, and to fetch jobs
	for i := 0; i < mgr.Dispatchers; i++ {
		go fetchJobsPipelined(mgr, in)
	}

	for i := 0; i < mgr.Concurrency; i++ {
		go processPipelined2(mgr, in, i)
	}

	go handleSignalsPipelined(mgr)

	mgr.shutdownWaiter.Add(1)
	defer mgr.shutdownWaiter.Done()

	<-mgr.done
	fmt.Println("Draining...")
	close(in)
	// TODO: Drain result queue / wait for workers to finish...
}

// func reportResultsPipelined2(mgr *Manager) {
// 	mgr.shutdownWaiter.Add(1)
// 	defer mgr.shutdownWaiter.Done()
// 	for result := range out {
// 		err := mgr.with(func(c *faktory.Client) (err error) {
// 			if result.err != nil {
// 				err = c.Fail(result.jid, result.err, result.backtrace)
// 			} else {
// 				err = c.Ack(result.jid)
// 			}
// 			return
// 		})
// 		if err != nil {
// 			mgr.Logger.Error(err)
// 		}
// 	}
// }

func reportResultPipelined2(mgr *Manager, job *faktory.Job, err error) error {
	return mgr.with(func(c *faktory.Client) error {
		if err != nil {
			return c.Fail(job.Jid, err, nil)
		}
		return c.Ack(job.Jid)
	})
}

func processPipelined2(mgr *Manager, in chan *faktory.Job, idx int) {
	mgr.shutdownWaiter.Add(1)
	defer mgr.shutdownWaiter.Done()

	for job := range in {
		perform := mgr.jobHandlers[job.Type]
		if perform == nil {
			reportErr := reportResultPipelined2(mgr, job, fmt.Errorf("No handler for %s", job.Type))
			if reportErr != nil {
				mgr.Logger.Error(reportErr)
			}

			continue
		}

		h := perform
		for i := len(mgr.middleware) - 1; i >= 0; i-- {
			h = mgr.middleware[i](h)
		}

		err := h(ctxFor(job), job)
		reportErr := reportResultPipelined2(mgr, job, err)
		if reportErr != nil {
			mgr.Logger.Error(reportErr)
		}
	}
}
