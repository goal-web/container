package tests

import (
	"fmt"
	"github.com/goal-web/container"
	"github.com/goal-web/contracts"
	"github.com/goal-web/supports/utils"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type DemoParam struct {
	Id string
}

type DemoComponent struct {
	Param DemoParam
}

func TestArgumentsTypeMap(t *testing.T) {
	args := container.NewArgumentsTypeMap([]interface{}{"啦啦啦", DemoParam{Id: "111"}})
	str := args.Pull("string")
	fmt.Println(str)
	assert.True(t, str == "啦啦啦")

	args = container.NewArgumentsTypeMap([]interface{}{})
	assert.True(t, args.Pull("string") == nil)
}

func TestBaseContainer(t *testing.T) {
	app := container.New()

	app.Instance("a", "a")
	assert.True(t, app.HasBound("a"))
	assert.True(t, app.Get("a") == "a")

	app.Alias("a", "A")

	assert.True(t, app.Get("A") == "a")
	assert.True(t, app.HasBound("A"))

	app.Bind("DemoParam", func() DemoParam {
		return DemoParam{Id: "测试一下"}
	})

	assert.True(t, app.Get(utils.GetTypeKey(reflect.TypeOf(DemoParam{}))).(DemoParam).Id == "测试一下")

	app.Call(container.NewMagicalFunc(func(param DemoParam) {
		assert.True(t, param.Id == "测试一下")
	}))

}

func TestContainer(t *testing.T) {
	app := container.New()

	app.Bind("DemoParam", func() DemoParam {
		return DemoParam{Id: "没有外部参数的话，从容器中获取"}
	})

	fn := container.NewMagicalFunc(func(param DemoParam) string {
		return param.Id
	})

	// 自己传参
	assert.True(t, app.Call(fn, DemoParam{Id: "优先使用外部参数"})[0] == "优先使用外部参数")

	// 不传参，使用容器中的实例
	assert.True(t, app.Call(fn)[0] == "没有外部参数的话，从容器中获取")

}

type DemoStruct struct {
	Param  DemoParam `di:""`       // 注入对应类型的实例
	Config string    `di:"config"` // 注入指定 key 的实例
}

func TestContainerMake(t *testing.T) {
	app := container.New()

	app.Instance("config", "通过容器设置的配置")

	app.Bind("DemoParam", func() DemoParam {
		return DemoParam{Id: "没有外部参数的话，从容器中获取"}
	})

	demo := &DemoStruct{}

	app.DI(demo)

	fmt.Println(demo)
}

func TestAliasType(t *testing.T) {
	app := container.New()

	app.Singleton("param", func() DemoParam {
		return DemoParam{
			Id: "a",
		}
	})

	type AliasParam DemoParam

	app.Call(container.NewMagicalFunc(func(param AliasParam) {
		fmt.Println(param)
	}), app.Get("param"))
}

type DemoStruct2 struct {
	DemoStruct
}

func (d *DemoStruct2) Construct(container2 contracts.Container) {
	d.DemoStruct = container2.Get("struct").(DemoStruct)
}

// 调用方法支持注入自定义类
func TestAutoContainer(t *testing.T) {
	app := container.New()

	app.Singleton("struct", func() DemoStruct {
		return DemoStruct{
			Param:  DemoParam{Id: "id"},
			Config: "config",
		}
	})

	//struct2Type := reflect.TypeOf(DemoStruct2{})
	//struct2Value := reflect.New(struct2Type).Interface()
	struct2Value := &DemoStruct2{}

	app.DI(struct2Value)

	app.Call(container.NewMagicalFunc(func(struct2 DemoStruct2) {
		assert.True(t, struct2.Config == "config" && struct2.Param.Id == "id")
	}))

	app.Call(container.NewMagicalFunc(func(struct2 DemoStruct2, struct1 DemoStruct) { // 因为 DemoStruct2 实现了 contracts.Component 所以不会使用自定义参数
		assert.True(t, struct2.Config == "config" && struct2.Param.Id == "id")
		assert.True(t, struct1.Config == "config22" && struct1.Param.Id == "custom")
	}), DemoStruct{
		Param:  DemoParam{Id: "custom"},
		Config: "config22",
	})
}

// 测试控制器执行
type DemoDependent struct {
	Id string
}

type DemoController struct {
	Dep DemoDependent `di:""` // 表示需要注入
}

func (this *DemoController) PrintDep() {
	fmt.Println(this.Dep)
}

func TestControllerCall(t *testing.T) {
	app := container.New()
	app.Singleton("DemoDependent", func() DemoDependent {
		return DemoDependent{
			Id: "id ddd",
		}
	})

	controller := &DemoController{}

	app.DI(controller)

	app.Call(container.NewMagicalFunc(controller.PrintDep))
}

func TestCallAndDIContainer(t *testing.T) {
	app := container.New()

	app.Call(container.NewMagicalFunc(func(container2 contracts.Container) {
		fmt.Println(container2)
	}))
}

func TestDIDontDefineValue(t *testing.T) {
	app := container.New()

	app.Call(func(name string, model contracts.Model) {
		fmt.Println(name, model)
	})
}

/**
goos: darwin
goarch: amd64
pkg: github.com/goal-web/container/tests
cpu: Intel(R) Core(TM) i7-7660U CPU @ 2.50GHz
BenchmarkCall
BenchmarkCall-4   	  854979	      1175 ns/op
*/
func BenchmarkCall(b *testing.B) {
	app := container.New()
	app.Singleton("DemoDependent", func() DemoDependent {
		return DemoDependent{
			Id: "id ddd",
		}
	})

	for i := 0; i < b.N; i++ {
		app.Call(func(dependent DemoDependent) {

		})
	}
}

/**
goos: darwin
goarch: amd64
pkg: github.com/goal-web/container/tests
cpu: Intel(R) Core(TM) i7-7660U CPU @ 2.50GHz
BenchmarkStaticCall
BenchmarkStaticCall-4   	 1048639	      1074 ns/op
*/
func BenchmarkStaticCall(b *testing.B) {
	app := container.New()
	app.Singleton("DemoDependent", func() DemoDependent {
		return DemoDependent{
			Id: "id ddd",
		}
	})
	staticFunc := container.NewMagicalFunc(func(dependent DemoDependent) {
	})

	for i := 0; i < b.N; i++ {
		app.Call(staticFunc)
	}
}
