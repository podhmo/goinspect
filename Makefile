test:
	go test ./...

dump-examples:
	mkdir -p internal/testdata
	go run ./cmd/goinspect/main.go --pkg ./internal/x/...  > internal/testdata/x.default.output
	go run ./cmd/goinspect/main.go --pkg ./internal/x/...  --expand-all > internal/testdata/x.expand.output
	go run ./cmd/goinspect/main.go --pkg ./internal/x/...  --expand-all --only F0 > internal/testdata/x.F0.expand.output
	go run ./cmd/goinspect/main.go --pkg ./internal/x/...  --expand-all --only W0 > internal/testdata/x.W0.expand.output
	go run ./cmd/goinspect/main.go --pkg ./internal/x/...  --expand-all --only W0.M0 > internal/testdata/x.W0.M0.expand.output
	go run ./cmd/goinspect/main.go --pkg ./internal/x/...  --expand-all --only R > internal/testdata/x.R.expand.output
	go run ./cmd/goinspect/main.go --pkg ./internal/x/...  --expand-all --only Odd > internal/testdata/x.Odd.expand.output