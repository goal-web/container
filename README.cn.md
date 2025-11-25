# Goal-Container

[![Go Reference](https://pkg.go.dev/badge/github.com/goal-web/container.svg)](https://pkg.go.dev/github.com/goal-web/container)
[![Go Report Card](https://goreportcard.com/badge/github.com/goal-web/container)](https://goreportcard.com/report/github.com/goal-web/container)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](../goal/LICENSE)
![GitHub Stars](https://img.shields.io/github/stars/goal-web/container?style=social)
![Release](https://img.shields.io/github/v/release/goal-web/container?include_prereleases)
![Go Version](https://img.shields.io/badge/go-%3E=%201.25.0-00ADD8?logo=go)
![CI](https://img.shields.io/github/actions/workflow/status/goal-web/container/ci.yml?branch=master&label=CI)
![Lint](https://img.shields.io/github/actions/workflow/status/goal-web/container/lint.yml?branch=master&label=Lint)

[Docs](https://pkg.go.dev/github.com/goal-web/container) · [Issues](https://github.com/goal-web/container/issues) · [Releases](https://github.com/goal-web/container/releases) · [English](./README.md)

Goal IoC 容器：可调用解析、参数注入与服务注册。

## 亮点

- 解析可调用与构造器
- 参数注入辅助
- 单例与工厂注册

## 兼容性

- Go `>= 1.25.0`
- 模块路径：`github.com/goal-web/container`

## Star History

<a href="https://star-history.com/#goal-web/container&Date"><img src="https://api.star-history.com/svg?repos=goal-web/container&type=Date" alt="Star History Chart"/></a>

![Stargazers over time](https://starchart.cc/goal-web/container.svg)

## 安装

```shell
go get github.com/goal-web/container
```

## 快速开始

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

## 可调用解析

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

## 依赖注入

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
