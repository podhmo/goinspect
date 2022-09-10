# goinspect
individual inspection command

## install

```console
$ go install github.com/podhmo/goinspect/cmd/goinspect@latest
```

## how to use

```console
$ goinspect --pkg ./internal/x/... --only F --include-unexported
package github.com/podhmo/goinspect/internal/x

  func x.F(s x.S)
    func x.log() func()
    func x.F0()
      func x.log() func()
      func x.F1()
        func x.H()
    func x.H()
```

[./internal/x/func.go](./internal/x/func.go)

## inspired by

- https://github.com/podhmo/pyinspect
