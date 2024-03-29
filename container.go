package container

import (
	"errors"
	"fmt"
	"github.com/goal-web/contracts"
	"github.com/goal-web/supports/exceptions"
	"github.com/goal-web/supports/utils"
	"reflect"
	"sync"
)

var (
	CallerTypeError = errors.New("参数类型必须是有一个返回值的函数")
)

type Container struct {
	binds        map[string]contracts.MagicalFunc
	singletons   map[string]contracts.MagicalFunc
	instances    sync.Map
	aliases      sync.Map
	argProviders []func(key string, p reflect.Type, arguments ArgumentsTypeMap) interface{}
}

func newInstanceProvider(provider interface{}) contracts.MagicalFunc {
	magicalFn := NewMagicalFunc(provider)
	if magicalFn.NumOut() != 1 {
		exceptions.Throw(CallerTypeError)
	}
	return magicalFn
}

func New() contracts.Container {
	container := &Container{}
	container.argProviders = []func(key string, p reflect.Type, arguments ArgumentsTypeMap) interface{}{
		func(key string, _ reflect.Type, arguments ArgumentsTypeMap) interface{} {
			return arguments.Pull(key) // 外部参数里面类型完全相等的参数
		},
		func(key string, argType reflect.Type, arguments ArgumentsTypeMap) interface{} {
			return arguments.FindConvertibleArg(key, argType) // 外部参数可转换的参数
		},
		func(key string, argType reflect.Type, arguments ArgumentsTypeMap) interface{} {
			return container.GetByArguments(key, arguments) // 从容器中获取参数
		},
		func(key string, argType reflect.Type, arguments ArgumentsTypeMap) interface{} {
			// 尝试 new 一个然后通过容器注入
			var (
				tempInstance interface{}
				isPtr        = argType.Kind() == reflect.Ptr
			)
			if isPtr {
				tempInstance = reflect.New(argType.Elem()).Interface()
			} else {
				tempInstance = reflect.New(argType).Interface()
			}
			container.DIByArguments(tempInstance, arguments)
			if isPtr {
				return tempInstance
			}
			return reflect.ValueOf(tempInstance).Elem().Interface()
		},
	}
	container.Flush()
	return container
}

func (container *Container) Bind(key string, provider interface{}) {
	magicalFn := newInstanceProvider(provider)
	container.binds[container.GetKey(key)] = magicalFn
	container.Alias(key, utils.GetTypeKey(magicalFn.Returns()[0]))
}

func (container *Container) Instance(key string, instance interface{}) {
	container.instances.Store(container.GetKey(key), instance)
}

func (container *Container) Singleton(key string, provider interface{}) {
	magicalFn := newInstanceProvider(provider)
	container.singletons[container.GetKey(key)] = magicalFn
	container.Alias(key, utils.GetTypeKey(magicalFn.Returns()[0]))
}

func (container *Container) HasBound(key string) bool {
	key = container.GetKey(key)
	if _, existsBind := container.binds[key]; existsBind {
		return true
	}
	if _, existsSingleton := container.singletons[key]; existsSingleton {
		return true
	}
	if _, existsInstance := container.instances.Load(key); existsInstance {
		return true
	}
	return false
}

func (container *Container) Alias(key string, alias string) {
	container.aliases.Store(alias, key)
}

func (container *Container) GetKey(alias string) string {
	if value, existsAlias := container.aliases.Load(alias); existsAlias {
		return value.(string)
	}
	return alias
}

func (container *Container) Flush() {
	container.instances = sync.Map{}
	container.singletons = make(map[string]contracts.MagicalFunc, 0)
	container.binds = make(map[string]contracts.MagicalFunc, 0)
	container.aliases = sync.Map{}
}

func (container *Container) Get(key string, args ...interface{}) interface{} {
	key = container.GetKey(key)
	if tempInstance, existsInstance := container.instances.Load(key); existsInstance {
		return tempInstance
	}
	if singletonProvider, existsProvider := container.singletons[key]; existsProvider {
		value := container.Call(singletonProvider, args...)[0]
		container.instances.Store(key, value)
		return value
	}
	if instanceProvider, existsProvider := container.binds[key]; existsProvider {
		return container.Call(instanceProvider, args...)[0]
	}
	return nil
}

func (container *Container) GetByArguments(key string, arguments ArgumentsTypeMap) interface{} {
	key = container.GetKey(key)
	if tempInstance, existsInstance := container.instances.Load(key); existsInstance {
		return tempInstance
	}
	if singletonProvider, existsProvider := container.singletons[key]; existsProvider {
		value := container.StaticCallByArguments(singletonProvider, arguments)[0]
		container.instances.Store(key, value)
		return value
	}
	if instanceProvider, existsProvider := container.binds[key]; existsProvider {
		return container.StaticCallByArguments(instanceProvider, arguments)[0]
	}
	return nil
}

// StaticCall 静态调用，直接传静态化的方法
func (container *Container) StaticCall(magicalFn contracts.MagicalFunc, args ...interface{}) []interface{} {
	return container.StaticCallByArguments(magicalFn, NewArgumentsTypeMap(append(args, container)))
}

// StaticCallByArguments 静态调用，直接传静态化的方法和处理好的参数
func (container *Container) StaticCallByArguments(magicalFn contracts.MagicalFunc, arguments ArgumentsTypeMap) []interface{} {
	fnArgs := make([]reflect.Value, 0)

	for _, arg := range magicalFn.Arguments() {
		key := utils.GetTypeKey(arg)
		fnArgs = append(fnArgs, reflect.ValueOf(container.findArg(key, arg, arguments)))
	}

	results := make([]interface{}, 0)

	for _, result := range magicalFn.Call(fnArgs) {
		results = append(results, result.Interface())
	}

	return results
}

func (container *Container) Call(fn interface{}, args ...interface{}) []interface{} {
	if magicalFn, isMagicalFunc := fn.(contracts.MagicalFunc); isMagicalFunc {
		return container.StaticCall(magicalFn, args...)
	}
	return container.StaticCall(NewMagicalFunc(fn), args...)
}

func (container *Container) findArg(key string, p reflect.Type, arguments ArgumentsTypeMap) (result interface{}) {
	for _, provider := range container.argProviders {
		if value := provider(key, p, arguments); value != nil {
			return value
		}
	}
	return
}

func (container *Container) DIByArguments(object interface{}, arguments ArgumentsTypeMap) {
	if component, ok := object.(contracts.Component); ok {
		component.Construct(container)
		return
	}

	objectValue := reflect.ValueOf(object)

	switch objectValue.Kind() {
	case reflect.Ptr:
		if objectValue.Elem().Kind() != reflect.Struct {
			exceptions.Throw(errors.New("参数必须是结构体指针"))
		}
		objectValue = objectValue.Elem()
	default:
		exceptions.Throw(errors.New("参数必须是结构体指针"))
	}

	valueType := objectValue.Type()

	var (
		fieldNum  = objectValue.NumField()
		tempValue = reflect.New(valueType).Elem()
	)

	tempValue.Set(objectValue)

	// 遍历所有字段
	for i := 0; i < fieldNum; i++ {
		var (
			field          = valueType.Field(i)
			key            = utils.GetTypeKey(field.Type)
			fieldTags      = utils.ParseStructTag(field.Tag)
			fieldValue     = tempValue.Field(i)
			fieldInterface interface{}
		)

		if di, existsDiTag := fieldTags["di"]; existsDiTag { // 配置了 fieldTags tag，优先用 tag 的配置
			if len(di) > 0 { // 如果指定某 di 值，优先取这个值
				fieldInterface = container.Get(di[0])
			}
			if fieldInterface == nil {
				fieldInterface = container.findArg(key, field.Type, arguments)
			}
		}

		if fieldInterface != nil {
			fieldType := reflect.TypeOf(fieldInterface)
			if fieldType.ConvertibleTo(field.Type) { // 可转换的类型
				value := reflect.ValueOf(fieldInterface)
				if key != utils.GetTypeKey(fieldType) { // 如果不是同一种类型，就转换一下
					value = value.Convert(field.Type)
				}
				fieldValue.Set(value)
			} else {
				exceptions.Throw(errors.New(fmt.Sprintf("无法注入 %s ，因为类型不一致，目标类型为 %s，而将注入的类型为 %s", field.Name, field.Type.String(), fieldType.String())))
			}
		}
	}

	objectValue.Set(tempValue)

	return
}

func (container *Container) DI(object interface{}, args ...interface{}) {
	container.DIByArguments(object, NewArgumentsTypeMap(append(args, container)))
}
