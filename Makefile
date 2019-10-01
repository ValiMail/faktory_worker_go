test:
	go test -race ./...

faktory_clean:
	rm -rf ~/.faktory/db/
	faktory

stats:
	while [ 1 ]; do \
		curl http://localhost:7420/stats 2>/dev/null | \
			jq -c '[.faktory.total_enqueued, .faktory.total_processed, .faktory.total_failures]' 2>/dev/null; \
			sleep 1; \
	done

producer:
	go run -race test/producer/main.go
	killall faktory
	say Done

runner:
	go run -race test/runner/main.go

runner_pipelined:
	go run -race test/runner_pipelined/main.go

runner_pipelined2:
	go run -race test/runner_pipelined2/main.go

.PHONY: work test
