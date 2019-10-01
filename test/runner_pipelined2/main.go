package main

import (
	"fmt"

	"github.com/ValiMail/faktory_worker_go/test/shared"
)

func main() {
	fmt.Println("Starting pipelined runner...")

	mgr := shared.NewConfiguredManager()

	// use up to N goroutines to fetch jobs
	mgr.Dispatchers = 32
	// use up to N goroutines to execute jobs
	mgr.Concurrency = 2048

	// Start processing jobs, this method does not return
	mgr.RunPipelined2()
}
