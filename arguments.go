package container

import (
	"github.com/goal-web/supports/utils"
	"reflect"
)

type ArgumentsTypeMap map[string][]interface{}

func NewArgumentsTypeMap(args []interface{}) ArgumentsTypeMap {
	argsTypeMap := ArgumentsTypeMap{}
	for _, arg := range args {
		argTypeKey := utils.GetTypeKey(reflect.TypeOf(arg))
		argsTypeMap[argTypeKey] = append(argsTypeMap[argTypeKey], arg)
	}
	return argsTypeMap
}

func (this ArgumentsTypeMap) Pull(key string) (arg interface{}) {
	if item, exits := this[key]; exits && len(item) >= 1 {
		arg = item[0]
		this[key] = item[1:]
		return
	}
	return nil
}

// FindConvertibleArg 找到可转换的参数
func (this ArgumentsTypeMap) FindConvertibleArg(targetKey string, targetType reflect.Type) interface{} {
	for key, args := range this {
		for _, arg := range args {
			if reflect.TypeOf(arg).ConvertibleTo(targetType) {
				if key != targetKey {
					return reflect.ValueOf(arg).Convert(targetType).Interface()
				}
				return arg
			}
		}
	}
	return nil
}
