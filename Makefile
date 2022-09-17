test:
	go test ./...

dump-examples:
	mkdir -p internal/testdata
	rm -f /tmp/goinspect
	go build  -o /tmp/goinspect ./cmd/goinspect/
	/tmp/goinspect ./cmd/goinspect/main.go --pkg ./internal/x/...  > internal/testdata/x.default.output
	/tmp/goinspect ./cmd/goinspect/main.go --pkg ./internal/x/...  --expand-all > internal/testdata/x.expand.output
	/tmp/goinspect ./cmd/goinspect/main.go --pkg ./internal/x/...  --include-unexported --expand-all > internal/testdata/x.expand-with-unexported.output
	/tmp/goinspect ./cmd/goinspect/main.go --pkg ./internal/x/...  --expand-all --only F0 > internal/testdata/x.F0.expand.output
	/tmp/goinspect ./cmd/goinspect/main.go --pkg ./internal/x/...  --expand-all --only W0 > internal/testdata/x.W0.expand.output
	/tmp/goinspect ./cmd/goinspect/main.go --pkg ./internal/x/...  --expand-all --only W0.M0 > internal/testdata/x.W0.M0.expand.output
	/tmp/goinspect ./cmd/goinspect/main.go --pkg ./internal/x/...  --expand-all --only R > internal/testdata/x.R.expand.output
	/tmp/goinspect ./cmd/goinspect/main.go --pkg ./internal/x/...  --expand-all --only Odd > internal/testdata/x.Odd.expand.output	
