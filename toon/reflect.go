package toon

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

// Unmarshal decodes TOON into the value pointed to by v.
func Unmarshal(data []byte, v any, opts ...DecodeOptions) error {
	decoded, err := Decode(data, opts...)
	if err != nil {
		return err
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &json.InvalidUnmarshalError{Type: reflect.TypeOf(v)}
	}
	return assignValue(rv.Elem(), decoded)
}

func normalizeForEncode(v any) (any, error) {
	if v == nil {
		return nil, nil
	}
	switch val := v.(type) {
	case Object:
		return val, nil
	case []any:
		return val, nil
	case bool, string, json.Number, int, int32, int64, float32, float64:
		return val, nil
	}
	return reflectToValue(reflect.ValueOf(v))
}

func reflectToValue(rv reflect.Value) (any, error) {
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Bool, reflect.String:
		return rv.Interface(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	case reflect.Interface:
		if rv.IsNil() {
			return nil, nil
		}
		return reflectToValue(rv.Elem())
	case reflect.Slice:
		if rv.Type() == reflect.TypeOf(Object(nil)) {
			return rv.Interface().(Object), nil
		}
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return nil, &json.UnsupportedTypeError{Type: rv.Type()}
		}
		if rv.IsNil() {
			return []any{}, nil
		}
		out := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			val, err := reflectToValue(rv.Index(i))
			if err != nil {
				return nil, err
			}
			out[i] = val
		}
		return out, nil
	case reflect.Array:
		out := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			val, err := reflectToValue(rv.Index(i))
			if err != nil {
				return nil, err
			}
			out[i] = val
		}
		return out, nil
	case reflect.Map:
		if rv.IsNil() {
			return Object{}, nil
		}
		out := make(Object, 0, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			key := iter.Key()
			if key.Kind() != reflect.String {
				return nil, &json.UnsupportedTypeError{Type: rv.Type()}
			}
			val, err := reflectToValue(iter.Value())
			if err != nil {
				return nil, err
			}
			out = append(out, Field{Key: key.String(), Value: val})
		}
		return out, nil
	case reflect.Struct:
		return structToObject(rv)
	default:
		return nil, &json.UnsupportedTypeError{Type: rv.Type()}
	}
}

func structToObject(rv reflect.Value) (Object, error) {
	rt := rv.Type()
	out := make(Object, 0, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		if sf.PkgPath != "" {
			continue
		}
		name := structFieldName(sf)
		if name == "" {
			continue
		}
		val, err := reflectToValue(rv.Field(i))
		if err != nil {
			return nil, err
		}
		out = append(out, Field{Key: name, Value: val})
	}
	return out, nil
}

func structFieldName(sf reflect.StructField) string {
	if tag := sf.Tag.Get("toon"); tag != "" {
		return parseStructTag(tag)
	}
	if tag := sf.Tag.Get("json"); tag != "" {
		return parseStructTag(tag)
	}
	return sf.Name
}

func parseStructTag(tag string) string {
	if tag == "-" {
		return ""
	}
	if idx := strings.IndexByte(tag, ','); idx >= 0 {
		tag = tag[:idx]
	}
	return tag
}

func assignValue(rv reflect.Value, decoded any) error {
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}

	if decoded == nil {
		rv.SetZero()
		return nil
	}

	switch rv.Kind() {
	case reflect.Interface:
		if rv.NumMethod() == 0 {
			rv.Set(reflect.ValueOf(decoded))
			return nil
		}
		return mismatchError(rv.Type(), decoded)
	case reflect.Bool:
		v, ok := decoded.(bool)
		if !ok {
			return mismatchError(rv.Type(), decoded)
		}
		rv.SetBool(v)
		return nil
	case reflect.String:
		v, ok := decoded.(string)
		if !ok {
			return mismatchError(rv.Type(), decoded)
		}
		rv.SetString(v)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := asInt64(decoded)
		if err != nil {
			return err
		}
		rv.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := asInt64(decoded)
		if err != nil {
			return err
		}
		rv.SetUint(uint64(n))
		return nil
	case reflect.Float32, reflect.Float64:
		n, err := asFloat64(decoded)
		if err != nil {
			return err
		}
		rv.SetFloat(n)
		return nil
	case reflect.Slice:
		arr, ok := decoded.([]any)
		if !ok {
			return mismatchError(rv.Type(), decoded)
		}
		slice := reflect.MakeSlice(rv.Type(), len(arr), len(arr))
		for i, item := range arr {
			if err := assignValue(slice.Index(i), item); err != nil {
				return err
			}
		}
		rv.Set(slice)
		return nil
	case reflect.Array:
		arr, ok := decoded.([]any)
		if !ok {
			return mismatchError(rv.Type(), decoded)
		}
		if len(arr) != rv.Len() {
			return mismatchError(rv.Type(), decoded)
		}
		for i, item := range arr {
			if err := assignValue(rv.Index(i), item); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		m, ok := decoded.(map[string]any)
		if !ok {
			return mismatchError(rv.Type(), decoded)
		}
		if rv.IsNil() {
			rv.Set(reflect.MakeMap(rv.Type()))
		}
		for key, val := range m {
			elem := reflect.New(rv.Type().Elem()).Elem()
			if err := assignValue(elem, val); err != nil {
				return err
			}
			rv.SetMapIndex(reflect.ValueOf(key), elem)
		}
		return nil
	}

	if rv.Kind() == reflect.Struct {
		m, ok := decoded.(map[string]any)
		if !ok {
			return mismatchError(rv.Type(), decoded)
		}
		return assignStruct(rv, m)
	}

	return mismatchError(rv.Type(), decoded)
}

func assignStruct(rv reflect.Value, m map[string]any) error {
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		if sf.PkgPath != "" {
			continue
		}
		name := structFieldName(sf)
		if name == "" {
			continue
		}
		val, ok := m[name]
		if !ok {
			continue
		}
		if err := assignValue(rv.Field(i), val); err != nil {
			return err
		}
	}
	return nil
}

func asInt64(v any) (int64, error) {
	switch n := v.(type) {
	case int:
		return int64(n), nil
	case int64:
		return n, nil
	case int32:
		return int64(n), nil
	case float64:
		return int64(n), nil
	case json.Number:
		return n.Int64()
	case string:
		return strconv.ParseInt(n, 10, 64)
	default:
		return 0, mismatchError(reflect.TypeOf(int64(0)), v)
	}
}

func asFloat64(v any) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case int:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case json.Number:
		return n.Float64()
	case string:
		return strconv.ParseFloat(n, 64)
	default:
		return 0, mismatchError(reflect.TypeOf(float64(0)), v)
	}
}

func mismatchError(dst reflect.Type, src any) error {
	return &json.UnmarshalTypeError{
		Value:  typeName(src),
		Type:   dst,
		Offset: 0,
	}
}

func typeName(v any) string {
	if v == nil {
		return "null"
	}
	return reflect.TypeOf(v).String()
}
