package zerocheck

import "reflect"

var emptyIn = []reflect.Value{}

// IsZero returns true if a value is nil, remember boolean's false is zero.
// Remember; fields to be checked should be exported otherwise it returns false.
func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Struct:
		zero := true
		for i := 0; i < v.NumField(); i++ {
			zero = zero && IsZero(v.Field(i))
		}

		if typ := v.Type(); typ != nil && v.IsValid() {
			f, ok := typ.MethodByName("IsZero")
			// if not found
			// if has input arguments (1 is for the value receiver, so > 1 for the actual input args)
			// if output argument is not boolean
			// then skip this IsZero user-defined function.
			if !ok || f.Type.NumIn() > 1 || f.Type.NumOut() != 1 && f.Type.Out(0).Kind() != reflect.Bool {
				return zero
			}

			method := v.Method(f.Index)
			// no needed check but:
			if method.IsValid() && !method.IsNil() {
				// it shouldn't panic here.
				zero = method.Call(emptyIn)[0].Interface().(bool)
			}
		}

		return zero
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		zero := true
		for i := 0; i < v.Len(); i++ {
			zero = zero && IsZero(v.Index(i))
		}
		return zero
	}
	// if not any special type then use the reflect's .Zero
	// usually for fields, but remember if it's boolean and it's false
	// then it's zero, even if set-ed.

	if !v.CanInterface() {
		// if can't interface, i.e return value from unexported field or method then return false
		return false
	}
	zero := reflect.Zero(v.Type())
	return v.Interface() == zero.Interface()
}
