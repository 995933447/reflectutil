package reflectutil

import (
	"database/sql"
	"errors"
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
		if !srcField.IsValid() {
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
