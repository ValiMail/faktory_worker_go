package shared

import (
	"time"

	worker "github.com/ValiMail/faktory_worker_go"
)

func SomeFunc(ctx worker.Context, args ...interface{}) error {
	time.Sleep(2000 * time.Millisecond)
	return nil
}

func NewConfiguredManager() *worker.Manager {
	mgr := worker.NewManager()

	// mgr.Use(func(perform worker.Handler) worker.Handler {
	// 	return func(ctx worker.Context, job *faktory.Job) error {
	// 		// log.Printf("Starting work on job %s of type %s with custom %v\n", ctx.Jid(), ctx.JobType(), job.Custom)
	// 		err := perform(ctx, job)
	// 		// log.Printf("Finished work on job %s with error %v\n", ctx.Jid(), err)
	// 		return err
	// 	}
	// })

	// register job types and the function to execute them
	mgr.Register("SomeJob", SomeFunc)
	mgr.Register("SomeWorker", SomeFunc)

	// pull jobs from these queues, in this order of precedence
	mgr.ProcessStrictPriorityQueues("critical", "default", "bulk")

	return mgr
}
