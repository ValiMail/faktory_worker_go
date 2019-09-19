package main

import (
	"time"

	worker "github.com/ValiMail/faktory_worker_go"
	faktory "github.com/contribsys/faktory/client"
)

func someFunc(ctx worker.Context, args ...interface{}) error {
	// log.Printf("Working on job %s\n", ctx.Jid())
	// log.Printf("Context %v\n", ctx)
	// log.Printf("Args %v\n", args)
	time.Sleep(100 * time.Millisecond)
	return nil
}

func main() {
	mgr := worker.NewManager()
	mgr.Use(func(perform worker.Handler) worker.Handler {
		return func(ctx worker.Context, job *faktory.Job) error {
			// log.Printf("Starting work on job %s of type %s with custom %v\n", ctx.Jid(), ctx.JobType(), job.Custom)
			err := perform(ctx, job)
			// log.Printf("Finished work on job %s with error %v\n", ctx.Jid(), err)
			return err
		}
	})

	// register job types and the function to execute them
	mgr.Register("SomeJob", someFunc)
	mgr.Register("SomeWorker", someFunc)
	//mgr.Register("AnotherJob", anotherFunc)

	// use up to N goroutines to fetch jobs, and N goroutines to report results
	// to the server
	mgr.Dispatchers = 20
	// use up to N goroutines to execute jobs
	mgr.Concurrency = 1000

	// pull jobs from these queues, in this order of precedence
	mgr.ProcessStrictPriorityQueues("critical", "default", "bulk")

	var quit bool
	mgr.On(worker.Shutdown, func() {
		quit = true
	})
	for i := 0; i < mgr.Dispatchers; i++ {
		go func() {
			client, err := faktory.Open()
			if err != nil {
				panic(err)
			}

			for {
				if quit {
					return
				}
				produce(client)
			}
		}()
	}

	// Start processing jobs, this method does not return
	mgr.Run()
}

// Push something for us to work on.
func produce(client *faktory.Client) {
	job := faktory.NewJob("SomeJob", 1, 2, "hello")
	job.Custom = map[string]interface{}{
		"hello": "world",
	}
	err := client.Push(job)
	if err != nil {
		panic(err)
	}
	// fmt.Println(cl.Info())
}
