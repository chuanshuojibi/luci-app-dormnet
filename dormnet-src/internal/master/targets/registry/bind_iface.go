//go:generate go tool go-enum --marshal

package registry

import (
	"reflect"
	"strconv"

	"github.com/openwrt-dormnet/dormnet/shared/errx"
)

type DormnetClientBindIface struct {
	Iface     string `json:"iface"`
	ExtraArgs any    `json:"extra_args"`
}

// ENUM(Flag, Value, ListValue)
type ExtraArgType string

type ExtraArgCandidate struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ExtraArgInfo struct {
	Type       ExtraArgType        `json:"type"`
	Candidates []ExtraArgCandidate `json:"candidates"`
	Id         string              `json:"id"`
	Title      string              `json:"title"`
	Desc       string              `json:"desc"`
	Default    any                 `json:"default"`
	IsPassword bool                `json:"is_pwd"`
	Required   bool                `json:"required"`
	ModalOnly  bool                `json:"modalonly"`

	field reflect.Value
}

func FormatExtraArgInfo(extraArgs any) (args []ExtraArgInfo, err errx.Exception) {
	args = make([]ExtraArgInfo, 0)

	if extraArgs != nil {
		clazz := reflect.ValueOf(extraArgs)
		if clazz.Kind() != reflect.Pointer || clazz.Elem().Kind() != reflect.Struct {
			panic(errx.NewException("you should always return a pointer reference to a struct in method DormnetClient#ExtraArgInfo()"))
		}
		clazz = clazz.Elem()
		typ := clazz.Type()
		for i := 0; i < clazz.NumField(); i++ {
			fieldOfStruct := typ.Field(i)
			if !fieldOfStruct.IsExported() {
				continue
			}
			field := clazz.FieldByName(fieldOfStruct.Name)
			tags := fieldOfStruct.Tag

			isPassword := lookupForBoolTag(tags, "is_pwd", false)
			required := lookupForBoolTag(tags, "required", false)
			modalonly := lookupForBoolTag(tags, "modalonly", false)

			argType := ExtraArgType(tags.Get("type"))

			candidates := make([]ExtraArgCandidate, 0)
			if argType == ExtraArgTypeListValue {
				fieldType := fieldOfStruct.Type
				valuesMtd := field.MethodByName("AllValues")
				var values reflect.Value
				if retVal := valuesMtd.Call([]reflect.Value{}); len(retVal) != 1 {
					panic(errx.IllegalStateError())
				} else {
					values = retVal[0]
				}
				for _, key := range values.MapKeys() {
					enumValue := values.MapIndex(key).Convert(fieldType)
					nameMtd := enumValue.MethodByName("Name")
					name := nameMtd.Call([]reflect.Value{})[0].String()
					candidate := ExtraArgCandidate{
						Name:  name,
						Value: key.String(),
					}
					candidates = append(candidates, candidate)
				}
			}

			args = append(args, ExtraArgInfo{
				Type:       argType,
				Id:         tags.Get("json"),
				Title:      tags.Get("title"),
				Desc:       tags.Get("desc"),
				Default:    field.Interface(),
				IsPassword: isPassword,
				Required:   required,
				ModalOnly:  modalonly,
				Candidates: candidates,

				field: field,
			})
		}
	}

	return
}

func lookupForBoolTag(tags reflect.StructTag, key string, defVal bool) bool {
	realVal := defVal
	if req, ok := tags.Lookup(key); ok {
		converted, err := strconv.ParseBool(req)
		if err == nil {
			realVal = converted
		}
	}
	return realVal
}
