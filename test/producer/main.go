package main

import (
	"fmt"
	"sync"
	"time"

	faktory "github.com/contribsys/faktory/client"
)

func main() {
	numJobs := 1000000
	numProducers := 10

	fmt.Printf("Producing %d jobs, spread across %d producers\n", numJobs, numProducers)

	start := time.Now().UTC()
	var wg sync.WaitGroup
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()

			client, err := faktory.Open()
			if err != nil {
				panic(err)
			}

			fmt.Printf("Beginning production from worker %d.\n", i)
			for i := 0; i < (numJobs / numProducers); i++ {
				produce(client)
			}
		}()
	}
	fmt.Println("Workers spawning loop completed.")
	wg.Wait()

	end := time.Now().UTC()
	elapsed := end.Sub(start).Seconds()
	fmt.Printf("Queued %d jobs in %f seconds\n", numJobs, elapsed)
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
}
