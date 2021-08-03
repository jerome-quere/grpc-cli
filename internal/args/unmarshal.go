package args

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Unmarshal(args []string, dest protoreflect.Message) error {

	// Map arg names to their values.
	// ["arg1=1", "arg2=2", "arg3"] => [ ["arg1","1"], ["arg2","2"], ["arg3",""] ]
	argsSlice := SplitRaw(args)

	processedArgNames := make(map[string]bool)

	// Loop through all arguments
	for _, kv := range argsSlice {
		argName, argValue := kv[0], kv[1]
		argNameWords := strings.Split(argName, ".")

		if processedArgNames[argName] {
			return &DuplicateArgError{
				ArgName: argName,
			}
		}

		err := unmarshalMessage(dest, argNameWords, argValue)
		if err != nil {
			return &UnmarshalArgError{
				ArgName:  argName,
				ArgValue: argValue,
				Err:      err,
			}
		}

	}

	return nil

	//f := msg.Descriptor().Fields().ByName("organization_id")
	//
	//v := protoreflect.ValueOf("test")
	//
	//ff := dynamicpb.NewMessage(f.Message())
	//ffd := ff.Descriptor().Fields().ByName("value")
	//ff.Set(ffd, v)
	//
	//msg.Set(f, protoreflect.ValueOfMessage(ff))
	//return nil
}

func unmarshalMessage(dest protoreflect.Message, argNameWords []string, value string) error {
	if unmarshal, hasUnmarshalFunc := unmarshalFuncs[dest.Descriptor().FullName()]; hasUnmarshalFunc {
		// TODO add if len(argNameWords) > 0
		return unmarshal(value, dest)
	}

	if len(argNameWords) == 0 {
		return fmt.Errorf("trying to set a value to a message not a field")
	}

	field := dest.Descriptor().Fields().ByName(protoreflect.Name(argNameWords[0]))
	if field == nil {
		return fmt.Errorf("unknown field %s", argNameWords[0])
	}
	return set(field, dest, argNameWords[1:], value)
}

func set(field protoreflect.FieldDescriptor, dest protoreflect.Message, argNameWords []string, value string) error {

	switch {
	case field.IsList():
		// If type is a slice:

		// We cannot handle slice without an index notation.
		if len(argNameWords) == 0 {
			return fmt.Errorf("missing index on array")
		}

		// We check if argNameWords[0] is a positive integer to handle cases like keys.0.value=12
		index, err := strconv.ParseUint(argNameWords[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid index array %s", argNameWords[0])
		}

		list := dest.Mutable(field).List()
		// Make sure array is big enough to access the correct index.
		diff := int(index) - list.Len()
		switch {
		case diff > 0:
			return fmt.Errorf("missing index %d in array", index)
		case diff == 0:
			// Append one element to our slice.
			list.Append(list.NewElement())
		case diff < 0:
			// Element already exist at current index.
		}

		item := list.Get(int(index))

		if field.Kind() == protoreflect.MessageKind {
			return unmarshalMessage(item.Message(), argNameWords[1:], value)
		}

		err = unmarshalValue(field, &item, argNameWords[1:], value)
		if err != nil {
			return err
		}
		list.Set(int(index), item)

	//case protoreflect.Map:
	//
	//	// If map is nil we create it.
	//	if dest.IsNil() {
	//		dest.Set(reflect.MakeMap(dest.Type()))
	//	}
	//	if len(argNameWords) == 0 {
	//		return &MissingMapKeyError{}
	//	}
	//
	//	// Create a new value if it does not exist, then call set and add result in the map
	//	mapKey := reflect.ValueOf(argNameWords[0])
	//	mapValue := dest.MapIndex(mapKey)
	//
	//	if !mapValue.IsValid() {
	//		mapValue = reflect.New(dest.Type().Elem()).Elem()
	//	}
	//	err := set(mapValue, argNameWords[1:], value)
	//	dest.SetMapIndex(mapKey, mapValue)
	//
	//	return err

	default:

		if field.Kind() == protoreflect.MessageKind {
			return unmarshalMessage(dest.Mutable(field).Message(), argNameWords, value)
		}

		v := protoreflect.Value{}
		err := unmarshalValue(field, &v, argNameWords, value)
		if err != nil {
			return err
		}
		dest.Set(field, v)
	}

	return nil
}

var scalarKinds = map[protoreflect.Kind]bool{
	protoreflect.Int32Kind:  true,
	protoreflect.Int64Kind:  true,
	protoreflect.Uint32Kind: true,
	protoreflect.Uint64Kind: true,
	protoreflect.DoubleKind: true,
	protoreflect.BoolKind:   true,
	protoreflect.StringKind: true,
}

// A type is unmarshalable if:
// - it implement Unmarshaler
// - it has an unmarshalFunc
// - it is a scalar type
func isUnmarshalableValue(field protoreflect.FieldDescriptor) bool {

	if field.Kind() == protoreflect.EnumKind {
		return true
	}

	_, isScalar := scalarKinds[field.Kind()]
	if isScalar {
		return true
	}

	if field.Kind() != protoreflect.MessageKind {
		return false
	}

	_, hasUnmarshalFunc := unmarshalFuncs[field.Message().FullName()]
	return hasUnmarshalFunc
}

func unmarshalValue(field protoreflect.FieldDescriptor, dest *protoreflect.Value, argNameWords []string, value string) error {

	if len(argNameWords) > 0 {
		return fmt.Errorf("cannot set nested field %s", argNameWords[0])
	}

	if field.Kind() == protoreflect.EnumKind {
		enumValue := field.Enum().Values().ByName(protoreflect.Name(value))
		if enumValue == nil {
			return fmt.Errorf("unknown enum value %s", value)
		}
		*dest = protoreflect.ValueOfEnum(enumValue.Number())
		return nil
	}

	return unmarshalScalar(field, dest, value)
}

// unmarshalScalar handles unmarshaling from a string to a scalar type .
// It handles transformation like Atoi if dest is an Integer.
func unmarshalScalar(field protoreflect.FieldDescriptor, dest *protoreflect.Value, value string) error {

	switch field.Kind() {
	case protoreflect.Int32Kind:
		i, err := strconv.ParseInt(value, 0, 32)
		*dest = protoreflect.ValueOfInt32(int32(i))
		return err
	case protoreflect.Int64Kind:
		i, err := strconv.ParseInt(value, 0, 64)
		*dest = protoreflect.ValueOfInt64(int64(i))
		return err
	case protoreflect.Uint32Kind:
		i, err := strconv.ParseUint(value, 0, 32)
		*dest = protoreflect.ValueOfUint32(uint32(i))
		return err
	case protoreflect.Uint64Kind:
		i, err := strconv.ParseUint(value, 0, 64)
		*dest = protoreflect.ValueOfUint64(i)
		return err
	case protoreflect.DoubleKind:
		f, err := strconv.ParseFloat(value, 64)
		*dest = protoreflect.ValueOfFloat64(f)
		return err
	case protoreflect.BoolKind:
		switch value {
		case "true":
			*dest = protoreflect.ValueOfBool(true)
		case "false":
			*dest = protoreflect.ValueOfBool(false)
		default:
			return fmt.Errorf("%s is not a valid boolean a=value", value)
		}
		return nil
	case protoreflect.StringKind:
		*dest = protoreflect.ValueOfString(value)
		return nil
	default:
		return fmt.Errorf("don't know how to unmarshal %s", field.Kind())
	}
}

type UnmarshalFunc func(value string, message protoreflect.Message) error

var unmarshalFuncs = map[protoreflect.FullName]UnmarshalFunc{
	"google.protobuf.StringValue": func(value string, dest protoreflect.Message) error {
		f := dest.Descriptor().Fields().ByName("value")
		dest.Set(f, protoreflect.ValueOfString(value))
		return nil
	},
	"google.protobuf.Timestamp": func(value string, dest protoreflect.Message) error {
		date, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return err
		}
		t := timestamppb.New(date)

		seconds := dest.Descriptor().Fields().ByName("seconds")
		nanos := dest.Descriptor().Fields().ByName("nanos")
		dest.Set(seconds, protoreflect.ValueOfInt64(t.Seconds))
		dest.Set(nanos, protoreflect.ValueOfInt32(t.Nanos))
		return nil
	},
	"google.protobuf.Int32Value": func(value string, dest protoreflect.Message) error {
		f := dest.Descriptor().Fields().ByName("value")
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return err
		}
		dest.Set(f, protoreflect.ValueOfInt32(int32(v)))
		return nil
	},
	"google.protobuf.UInt32Value": func(value string, dest protoreflect.Message) error {
		f := dest.Descriptor().Fields().ByName("value")
		v, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return err
		}
		dest.Set(f, protoreflect.ValueOfUint32(uint32(v)))
		return nil
	},
	"google.protobuf.Int64Value": func(value string, dest protoreflect.Message) error {
		f := dest.Descriptor().Fields().ByName("value")
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		dest.Set(f, protoreflect.ValueOfInt64(v))
		return nil
	},
	"google.protobuf.UInt64Value": func(value string, dest protoreflect.Message) error {
		f := dest.Descriptor().Fields().ByName("value")
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		dest.Set(f, protoreflect.ValueOfUint64(v))
		return nil
	},
}
