package reflectutil

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

func CopySameFields(src, dest interface{}) error {
	srcVal := DeepGetElemVal(reflect.ValueOf(src))
	destVal := DeepGetElemVal(reflect.ValueOf(dest))

	if !destVal.CanAddr() {
		return errors.New("copy to value is unaddressable")
	}

	srcType := srcVal.Type()
	destType := destVal.Type()

	if srcType.Kind() != reflect.Struct && srcType.ConvertibleTo(destType) {
		destVal.Set(srcVal)
		return nil
	}

	if srcType.Kind() != reflect.Struct || destType.Kind() != reflect.Struct {
		return nil
	}

	if !srcVal.IsValid() {
		destVal.Set(reflect.Zero(srcType))
		return nil
	}

	var tryCopyField func(srcField, destField reflect.Value) bool

	tryCopyField = func(srcField, destField reflect.Value) bool {
		if !srcField.IsValid() || srcField.IsZero() {
			destField.Set(reflect.Zero(destField.Type()))
			return true
		}

		if destField.Kind() == reflect.Ptr {
			destField = destField.Elem()
		}

		if srcField.Type().ConvertibleTo(destField.Type()) {
			destField.Set(srcField)
			return true
		}

		if scanner, ok := destField.Addr().Interface().(sql.Scanner); ok {
			if err := scanner.Scan(srcField); err != nil {
				return false
			}
			return true
		}

		if srcField.Kind() == reflect.Ptr {
			return tryCopyField(srcField, srcField.Elem())
		}

		return true
	}

	srcFields, err := DeepGetStructFields(srcVal.Type())
	if err != nil {
		return err
	}

	for _, srcField := range srcFields {
		destField := destVal.FieldByName(srcField.Name)
		if !destField.CanSet() {
			continue
		}

		tryCopyField(srcVal.FieldByName(srcField.Name), destField)
	}

	return nil
}

func DeepGetElemType(reflectType reflect.Type) reflect.Type {
	for reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}
	return reflectType
}

func DeepGetElemVal(reflectVal reflect.Value) reflect.Value {
	for reflectVal.Kind() == reflect.Ptr {
		reflectVal = reflectVal.Elem()
	}
	return reflectVal
}

func DeepGetStructFields(reflectType reflect.Type) ([]reflect.StructField, error) {
	if elemType := DeepGetElemType(reflectType); elemType.Kind() != reflect.Struct {
		return nil, errors.New("not struct elem")
	}

	var fields []reflect.StructField
	structFieldNum := reflectType.NumField()
	for i := 0; i < structFieldNum; i++ {
		field := reflectType.Field(i)
		if field.Anonymous {
			anonymousFields, err := DeepGetStructFields(field.Type)
			if err != nil {
				return nil, err
			}
			fields = append(fields, anonymousFields...)
		} else {
			fields = append(fields, field)
		}
	}

	return fields, nil
}

func pluck(list interface{}, fieldName string, defaultVal interface{}) interface{} {
	reflectVal := reflect.ValueOf(list)
	switch reflectVal.Kind() {
	case reflect.Array, reflect.Slice:
		if reflectVal.Len() == 0 {
			return defaultVal
		}

		valElem := reflectVal.Type().Elem()
		for valElem.Kind() == reflect.Ptr {
			valElem = valElem.Elem()
		}

		if valElem.Kind() != reflect.Struct {
			panic("list element is not struct")
		}

		field, ok := valElem.FieldByName(fieldName)
		if !ok {
			panic(fmt.Sprintf("field %s not found", fieldName))
		}

		result := reflect.MakeSlice(reflect.SliceOf(field.Type), reflectVal.Len(), reflectVal.Len())

		for i := 0; i < reflectVal.Len(); i++ {
			ev := reflectVal.Index(i)
			for ev.Kind() == reflect.Ptr {
				ev = ev.Elem()
			}
			if ev.Kind() != reflect.Struct {
				panic("element is not a struct")
			}
			if !ev.IsValid() {
				continue
			}
			result.Index(i).Set(ev.FieldByIndex(field.Index))
		}

		return result.Interface()
	default:
		panic("list must be an array or slice")
	}
}

func PluckInt(list interface{}, fieldName string) []int {
	return pluck(list, fieldName, []int{}).([]int)
}

func PluckInt32(list interface{}, fieldName string) []int32 {
	return pluck(list, fieldName, []int32{}).([]int32)
}

func PluckUint32(list interface{}, fileName string) []uint32 {
	return pluck(list, fileName, []uint32{}).([]uint32)
}

func PluckUint64(list interface{}, fieldName string) []uint64 {
	return pluck(list, fieldName, []uint64{}).([]uint64)
}

func PluckString(list interface{}, fieldName string) []string {
	return pluck(list, fieldName, []string{}).([]string)
}

func MapByKey(list interface{}, fieldName string) interface{} {
	reflectVal := reflect.ValueOf(list)

	switch reflectVal.Kind() {
	case reflect.Slice, reflect.Array:
	default:
		panic("list required slice or array type")
	}

	valElem := reflectVal.Type().Elem()
	deepValElem := valElem
	for deepValElem.Kind() == reflect.Ptr {
		deepValElem = deepValElem.Elem()
	}

	if deepValElem.Kind() != reflect.Struct {
		panic("element not struct")
	}

	field, ok := deepValElem.FieldByName(fieldName)
	if !ok {
		panic(fmt.Sprintf("field %s not found", fieldName))
	}

	m := reflect.MakeMapWithSize(reflect.MapOf(field.Type, valElem), reflectVal.Len())
	for i := 0; i < reflectVal.Len(); i++ {
		elem := reflectVal.Index(i)
		elemStruct := elem
		for elemStruct.Kind() == reflect.Ptr {
			elemStruct = elemStruct.Elem()
		}

		// 如果是nil的，意味着key和value同时不存在，所以跳过不处理
		if !elemStruct.IsValid() {
			continue
		}

		if elemStruct.Kind() != reflect.Struct {
			panic("element not struct")
		}

		m.SetMapIndex(elemStruct.FieldByIndex(field.Index), elem)
	}

	return m.Interface()
}

func IsBlank(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}
	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}
