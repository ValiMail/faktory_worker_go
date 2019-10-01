package faktory_worker

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	faktory "github.com/contribsys/faktory/client"
)

// jobResult represents the results of an attempt to execute a job
type jobResult struct {
	jid       string
	err       error
	backtrace []byte
}

// RunPipelined starts processing jobs, using one pool of goroutines to fetch work, and another to execute it.
// This method does not return.
func (mgr *Manager) RunPipelined() {
	// This will signal to Faktory that all connections from this process
	// are worker connections.
	rand.Seed(time.Now().UnixNano())
	faktory.RandomProcessWid = strconv.FormatInt(rand.Int63(), 32)

	if mgr.Pool == nil {
		// Configure max connection pool size to number of fetchers + number of result-reporters + 1 for
		// the heartbeat process
		pool, err := NewChannelPool(mgr.Dispatchers*2+1, mgr.Dispatchers*2+1, func() (Closeable, error) { return faktory.Open() })
		if err != nil {
			panic(err)
		}
		mgr.Pool = pool
	}

	// a channel to distribute jobs to worker goroutines.  unbuffered to provide backpressure.
	in := make(chan *faktory.Job)
	// a channel to gather job results from worker goroutines so they can be returned to the server.
	out := make(chan *jobResult)

	mgr.fireEvent(Startup)

	go heartbeat(mgr)

	// start a set of goroutines to report results to the server, and to fetch jobs
	for i := 0; i < mgr.Dispatchers; i++ {
		go reportResultsPipelined(mgr, out)
		go fetchJobsPipelined(mgr, in)
	}

	for i := 0; i < mgr.Concurrency; i++ {
		go processPipelined(mgr, in, out, i)
	}

	go handleSignalsPipelined(mgr)

	mgr.shutdownWaiter.Add(1)
	defer mgr.shutdownWaiter.Done()

	<-mgr.done
	fmt.Println("Draining...")
	close(in)
	// TODO: Drain result queue / wait for workers to finish...
}

func handleSignalsPipelined(mgr *Manager) {
	sigchan := hookSignals()

	for {
		sig := <-sigchan
		handleEvent(signalMap[sig], mgr)
	}
}

func fetchJobsPipelined(mgr *Manager, in chan *faktory.Job) {
	mgr.preShutdownWaiter.Add(1)
	defer func() {
		mgr.preShutdownWaiter.Done()
	}()

	// delay initial fetch randomly to prevent thundering herd.
	time.Sleep(time.Duration(rand.Int31()))

	for {
		select {
		case <-mgr.preDone:
			return
		default:
		}

		if mgr.quiet {
			return
		}

		var job *faktory.Job
		var err error
		err = mgr.with(func(c *faktory.Client) (err error) {
			job, err = c.Fetch(mgr.queueList()...)
			return
		})

		if err != nil {
			mgr.Logger.Error(err)
			time.Sleep(1 * time.Second)
			continue
		}

		// job can be nil if nothing was received from the server!
		if job == nil {
			continue
		}

		in <- job
	}
}

func reportResultsPipelined(mgr *Manager, out chan *jobResult) {
	mgr.shutdownWaiter.Add(1)
	defer mgr.shutdownWaiter.Done()
	for result := range out {
		err := mgr.with(func(c *faktory.Client) (err error) {
			if result.err != nil {
				err = c.Fail(result.jid, result.err, result.backtrace)
			} else {
				err = c.Ack(result.jid)
			}
			return
		})
		if err != nil {
			mgr.Logger.Error(err)
		}
	}
}

func processPipelined(mgr *Manager, in chan *faktory.Job, out chan *jobResult, idx int) {
	mgr.shutdownWaiter.Add(1)
	defer mgr.shutdownWaiter.Done()

	for job := range in {
		perform := mgr.jobHandlers[job.Type]
		if perform == nil {
			out <- &jobResult{job.Jid, fmt.Errorf("No handler for %s", job.Type), nil}
			continue
		}

		h := perform
		for i := len(mgr.middleware) - 1; i >= 0; i-- {
			h = mgr.middleware[i](h)
		}

		err := h(ctxFor(job), job)
		out <- &jobResult{job.Jid, err, nil}
	}
}
