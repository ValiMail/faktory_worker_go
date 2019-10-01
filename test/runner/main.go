package main

import (
	"fmt"
	"time"

	worker "github.com/ValiMail/faktory_worker_go"
)

func someFunc(ctx worker.Context, args ...interface{}) error {
	time.Sleep(100 * time.Millisecond)
	return nil
}

func main() {
	fmt.Println("Starting basic runner...")

	mgr := worker.NewManager()

	// use up to N goroutines to execute jobs
	mgr.Concurrency = 768

	// mgr.Use(func(perform worker.Handler) worker.Handler {
	// 	return func(ctx worker.Context, job *faktory.Job) error {
	// 		// log.Printf("Starting work on job %s of type %s with custom %v\n", ctx.Jid(), ctx.JobType(), job.Custom)
	// 		err := perform(ctx, job)
	// 		// log.Printf("Finished work on job %s with error %v\n", ctx.Jid(), err)
	// 		return err
	// 	}
	// })

	// register job types and the function to execute them
	mgr.Register("SomeJob", someFunc)
	mgr.Register("SomeWorker", someFunc)

	// pull jobs from these queues, in this order of precedence
	mgr.ProcessStrictPriorityQueues("critical", "default", "bulk")

	// Start processing jobs, this method does not return
	mgr.Run()
}
