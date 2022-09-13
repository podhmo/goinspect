test:
	go test ./...

dump-examples:
	mkdir -p internal/testdata
	go run ./cmd/goinspect/main.go --pkg ./internal/x/...  > internal/testdata/x.default.output
	go run ./cmd/goinspect/main.go --pkg ./internal/x/...  --expand-all > internal/testdata/x.expand.output