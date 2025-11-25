package tests

import (
    "testing"
    "github.com/goal-web/container"
)

func TestNewMagicalFuncBasics(t *testing.T) {
    fn := func(a int) string { return "x" }
    mf := container.NewMagicalFunc(fn)
    if mf.NumIn() != 1 || mf.NumOut() != 1 { t.Fatalf("NumIn/NumOut mismatch") }
    if len(mf.Arguments()) != 1 || len(mf.Returns()) != 1 { t.Fatalf("args/returns length mismatch") }
    if mf.IsVariadic() { t.Fatalf("should not be variadic") }
    if mf.Signature() == "" { t.Fatalf("signature empty") }
}

func TestNewMagicalFuncPanicOnNonFunc(t *testing.T) {
    defer func() { if recover() == nil { t.Fatalf("expected panic on non-func") } }()
    _ = container.NewMagicalFunc(123)
}

type Repo struct{}
type Handler struct{ R *Repo }

func newRepo() *Repo { return &Repo{} }
func NewHandler(r *Repo) *Handler { return &Handler{R: r} }

func TestContainerBindSingletonInstanceGet(t *testing.T) {
    c := container.New()
    c.Singleton("repo", func() *Repo { return newRepo() })
    r1 := c.Get("repo").(*Repo)
    r2 := c.Get("repo").(*Repo)
    if r1 != r2 { t.Fatalf("singleton not cached") }
    h := &Handler{R: r1}
    c.Instance("handler", h)
    if c.Get("handler").(*Handler) != h { t.Fatalf("instance mismatch") }
    c.Bind("name", func() string { return "goal" })
    if c.Get("name").(string) != "goal" { t.Fatalf("bind get mismatch") }
}

func TestAliasAndHasBound(t *testing.T) {
    c := container.New()
    c.Singleton("repo", func() *Repo { return newRepo() })
    c.Alias("repo", "Repository")
    if !c.HasBound("Repository") { t.Fatalf("HasBound alias failed") }
    if c.Get("Repository") != c.Get("repo") { t.Fatalf("alias resolution failed") }
}

func TestCallInjectsDependencies(t *testing.T) {
    c := container.New()
    c.Singleton("repo", func() *Repo { return newRepo() })
    out := c.Call(NewHandler)
    h := out[0].(*Handler)
    if h.R == nil { t.Fatalf("Call did not inject Repo") }
}

func TestDIPanicOnNonStructPtr(t *testing.T) {
    defer func() { if recover() == nil { t.Fatalf("expected panic on non-struct ptr") } }()
    c := container.New()
    var notStruct = &[]int{}
    c.DI(notStruct)
}

func sum(nums ...int) int {
    s := 0
    for _, n := range nums { s += n }
    return s
}

func TestStaticCallVariadic(t *testing.T) {
    c := container.New()
    mf := container.NewMagicalFunc(sum)
    out := c.StaticCall(mf, 1, 2, 3)
    if out[0].(int) != 6 { t.Fatalf("variadic sum mismatch: %v", out[0]) }
}

func TestBindPanicOnWrongProviderSignature(t *testing.T) {
    defer func() { if recover() == nil { t.Fatalf("expected panic on wrong provider") } }()
    c := container.New()
    c.Bind("bad", func() {})
}