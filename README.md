# Goal-Container

[![Go Reference](https://pkg.go.dev/badge/github.com/goal-web/container.svg)](https://pkg.go.dev/github.com/goal-web/container)
[![Go Report Card](https://goreportcard.com/badge/github.com/goal-web/container)](https://goreportcard.com/report/github.com/goal-web/container)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](../goal/LICENSE)
![GitHub Stars](https://img.shields.io/github/stars/goal-web/container?style=social)
![Release](https://img.shields.io/github/v/release/goal-web/container?include_prereleases)
![Go Version](https://img.shields.io/badge/go-%3E=%201.25.0-00ADD8?logo=go)
![CI](https://img.shields.io/github/actions/workflow/status/goal-web/container/ci.yml?branch=master&label=CI)
![Lint](https://img.shields.io/github/actions/workflow/status/goal-web/container/lint.yml?branch=master&label=Lint)

[Docs](https://pkg.go.dev/github.com/goal-web/container) · [Issues](https://github.com/goal-web/container/issues) · [Releases](https://github.com/goal-web/container/releases) · [中文文档](./README.cn.md)

IoC container for Goal: callable resolution, argument injection and service registration.

## Highlights

- Resolve callables and constructors
- Argument injection helpers
- Register singletons and factories

## Compatibility

- Go `>= 1.25.0`
- Module path: `github.com/goal-web/container`

## Star History

<a href="https://star-history.com/#goal-web/container&Date"><img src="https://api.star-history.com/svg?repos=goal-web/container&type=Date" alt="Star History Chart"/></a>

![Stargazers over time](https://starchart.cc/goal-web/container.svg)

## Installation

```shell
go get github.com/goal-web/container
```

## Quick Start

```go
package main

import (
    "github.com/goal-web/container"
    "github.com/goal-web/contracts"
)

type Svc struct{ V string }

func NewSvc() *Svc { return &Svc{V: "ok"} }

func main() {
    c := container.New()
    c.Singleton("svc", func() *Svc { return NewSvc() })
    c.Bind("name", func() string { return "goal" })
    v := c.Get("svc").(*Svc)
    _ = v
}
```

## Callable Resolution

```go
package main

import (
    "github.com/goal-web/container"
)

type Repo struct{ }
type UseCase struct{ R *Repo }

func NewRepo() *Repo { return &Repo{} }
func NewUseCase(r *Repo, name string) *UseCase { return &UseCase{R: r} }

func main() {
    c := container.New()
    c.Singleton("repo", func() *Repo { return NewRepo() })
    c.Bind("name", func() string { return "goal" })
    r := c.Call(NewUseCase)[0].(*UseCase)
    _ = r
}
```

## Dependency Injection

```go
package main

import (
    "github.com/goal-web/container"
)

type Repo struct{ }
type Handler struct{ R *Repo }

func NewRepo() *Repo { return &Repo{} }

func main() {
    c := container.New()
    c.Singleton("repo", func() *Repo { return NewRepo() })
    h := &Handler{}
    c.DI(h)
    _ = h
}
```

## Aliases and Instances

```go
package main

import (
    "github.com/goal-web/container"
)

type Repo struct{}

func NewRepo() *Repo { return &Repo{} }

func main() {
    c := container.New()
    c.Singleton("repo", func() *Repo { return NewRepo() })
    c.Alias("repo", "Repository")
    _ = c.Get("Repository")
    h := &struct{ Name string }{Name: "goal"}
    c.Instance("handler", h)
    _ = c.Get("handler")
}
```

## Variadic Arguments

```go
package main

import (
    "github.com/goal-web/container"
)

func sum(nums ...int) int {
    s := 0
    for _, n := range nums { s += n }
    return s
}

func main() {
    c := container.New()
    mf := container.NewMagicalFunc(sum)
    _ = c.StaticCall(mf, 1, 2, 3)
}
```

## DI Tag Overrides

```go
package main

import (
    "github.com/goal-web/container"
)

type Repo struct{}
type Handler struct{ R *Repo `di:"repo"` }

func NewRepo() *Repo { return &Repo{} }

func main() {
    c := container.New()
    c.Singleton("repo", func() *Repo { return NewRepo() })
    h := &Handler{}
    c.DI(h)
    _ = h
}
```

## Component Construct

```go
package main

import (
    "github.com/goal-web/container"
    "github.com/goal-web/contracts"
)

type Cfg struct{ V string }
type Comp struct{ C Cfg }

func (c *Comp) Construct(app contracts.Container) {
    c.C = app.Get("cfg").(Cfg)
}

func main() {
    app := container.New()
    app.Instance("cfg", Cfg{V: "x"})
    comp := &Comp{}
    app.DI(comp)
    _ = comp
}
```

## Error Handling and Flush

```go
package main

import (
    "github.com/goal-web/container"
)

func main() {
    c := container.New()
    c.Flush()
}
```