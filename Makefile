FAKTORY_HOST=sinclair.local
FAKTORY_URL=tcp://${FAKTORY_HOST}:7419

test:
	go test -race ./...

faktory_clean:
	rm -rf ~/.faktory/db/
	faktory

stats:
	while [ 1 ]; do \
		curl http://${FAKTORY_HOST}:7420/stats 2>/dev/null | \
			jq -c '[.faktory.total_enqueued, .faktory.total_processed, .faktory.total_failures]' 2>/dev/null; \
			sleep 1; \
	done

producer:
	FAKTORY_URL=${FAKTORY_URL} go run -race test/producer/main.go
	@curl http://${FAKTORY_HOST}:7420/stats 2>/dev/null | \
		jq -c '[.faktory.total_enqueued, .faktory.total_processed, .faktory.total_failures]' 2>/dev/null
	say Done

runner:
	FAKTORY_URL=${FAKTORY_URL} go run -race test/runner/main.go

runner_pipelined:
	FAKTORY_URL=${FAKTORY_URL} go run -race test/runner_pipelined/main.go

runner_pipelined2:
	FAKTORY_URL=${FAKTORY_URL} go run -race test/runner_pipelined2/main.go

.PHONY: work test
