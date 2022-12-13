go:
	go vet
	go test .

install: go	
	go install

bench:
	go test -bench=$(sel) -benchmem -count=$(cnt)
sel=.
cnt=5

.PHONY: go install bench
