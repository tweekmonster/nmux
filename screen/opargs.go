package screen

import (
	"fmt"
	"reflect"
)

// This makes it easier to get types out of RPC response arguments.  They
// shouldn't be used for much else.
type opArgs struct {
	args []interface{}
}

func anyInt(val interface{}) int64 {
	switch v := val.(type) {
	case int64:
		return v
	case uint64:
		return int64(v)
	case int32:
		return int64(v)
	case uint32:
		return int64(v)
	case int16:
		return int64(v)
	case uint16:
		return int64(v)
	case int8:
		return int64(v)
	case uint8:
		return int64(v)
	}

	return 0
}

func (r *opArgs) next() interface{} {
	if len(r.args) == 0 {
		return nil
	}
	a := r.args[0]
	r.args = r.args[1:]
	return a
}

func (r *opArgs) String() string {
	var str string

	for {
		a := r.next()
		if a == nil {
			return str
		}

		switch v := a.(type) {
		case string:
			str += v
		default:
			r.args = append([]interface{}{a}, r.args...)
			return str
		}
	}
}

func (r *opArgs) Int64() int64 {
	return anyInt(r.next())
}

func (r *opArgs) Uint() uint {
	return uint(anyInt(r.next()))
}

func (r *opArgs) Int() int {
	return int(anyInt(r.next()))
}

func (r *opArgs) Uint32() uint32 {
	return uint32(anyInt(r.next()))
}

func (r *opArgs) Map() opMap {
	a := r.next()
	switch v := a.(type) {
	case map[string]interface{}:
		return opMap{args: v}
	default:
		panic(fmt.Sprintf("Expected map value, got %s", reflect.TypeOf(v)))
	}
}

type opMap struct {
	args map[string]interface{}
}

func (o opMap) Bool(key string) (bool, bool) {
	val, ok := o.args[key]
	if !ok {
		return false, false
	}

	switch v := val.(type) {
	case bool:
		return v, true
	default:
		panic("Map expected bool value")
	}
}

func (o opMap) Int64(key string) (int64, bool) {
	val, ok := o.args[key]
	if !ok {
		return 0, false
	}

	return anyInt(val), true
}

func (o opMap) Uint(key string) (uint, bool) {
	v, ok := o.Int64(key)
	return uint(v), ok
}

func (o opMap) Uint32(key string) (uint32, bool) {
	v, ok := o.Int64(key)
	return uint32(v), ok
}

func (o opMap) Int(key string) (int, bool) {
	v, ok := o.Int64(key)
	return int(v), ok
}
