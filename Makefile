go:
	go build
	go test .

install: go	
	go install -ldflags=-s

vet:
	go vet

bench:
	go test -bench=$(sel) -benchmem -count=$(cnt)
sel=.
cnt=5

.PHONY: go install vet bench
