package mapper

import (
	"reflect"

	"github.com/bytedance/sonic"
	"github.com/jinzhu/copier"
)

// Copy - copy struct to struct
func Copy(dest, src interface{}) error {
	return copier.Copy(dest, src)
}

// CopyIgnoreEmpty - copy struct to struct ignore zero value
func CopyIgnoreEmpty(dest, src interface{}) error {
	return copier.CopyWithOption(dest, src, copier.Option{IgnoreEmpty: true})
}

// BindingStruct - biding struct to struct
func BindingStruct(src interface{}, desc interface{}) error {
	byteSrc, err := sonic.Marshal(src)
	if err != nil {
		return err
	}
	return sonic.Unmarshal(byteSrc, desc)
}

func BindingAndValidate[T any](detail interface{}, validator func(interface{}) error) (T, error) {
	var model T
	if err := BindingStruct(detail, &model); err != nil {
		return model, err
	}

	if err := validator(model); err != nil {
		return model, err
	}
	return model, nil
}

func StructToMap(input interface{}, ignoreNilFiled bool) map[string]interface{} {
	result := make(map[string]interface{})
	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("json") == "" {
			continue
		}

		fv := v.Field(i)
		if fv.Kind() == reflect.Pointer {
			if fv.IsNil() {
				if ignoreNilFiled {
					continue
				}
				result[field.Tag.Get("json")] = nil
				continue
			}
			fv = fv.Elem()
		}

		result[field.Tag.Get("json")] = fv.Interface()
	}
	return result
}

// GetJsonStringify converts a struct to a JSON string, excluding specified fields.
func GetJsonStringify(src interface{}) string {
	byteData, err := sonic.Marshal(src)
	if err != nil {
		return ""
	}
	return string(byteData)
}

func ConvertMapToString(data map[string]interface{}) string {
	byteData, err := sonic.Marshal(data)
	if err != nil {
		return ""
	}
	return string(byteData)
}
