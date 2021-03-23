package args

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"

	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

//go:embed testdata/test.pb
var rawProto []byte

func TestUnmarshal(t *testing.T) {
	fileDescSet := descriptorpb.FileDescriptorSet{}
	err := proto.Unmarshal(rawProto, &fileDescSet)
	require.NoError(t, err)
	files, err := protodesc.NewFiles(&fileDescSet)
	require.NoError(t, err)

	t.Run("Simple", func(t *testing.T) {

		desc, err := files.FindDescriptorByName("test.Simple")
		require.NoError(t, err)
		message := dynamicpb.NewMessage(desc.(protoreflect.MessageDescriptor))
		err = Unmarshal([]string{
			"str=abc",
			"int32=32",
			"int64=64",
			"uint32=32",
			"uint64=64",
			"double=6.4",
			"bool=true",
			"enum=enum_value2",
		}, message)
		require.NoError(t, err)
		res, err := protojson.MarshalOptions{Indent: ""}.Marshal(message)
		require.NoError(t, err)
		assert.Equal(t, `{"str":"abc","int32":32,"int64":"64","uint32":32,"uint64":"64","double":6.4,"bool":true,"enum":"enum_value2"}`, string(res))

	})

	t.Run("Repeated", func(t *testing.T) {

		desc, err := files.FindDescriptorByName("test.Simple")
		require.NoError(t, err)
		message := dynamicpb.NewMessage(desc.(protoreflect.MessageDescriptor))
		err = Unmarshal([]string{
			"strs.0=abc1",
			"strs.1=abc2",
			"int32s.0=21",
			"int32s.1=42",
			"int64s.0=21",
			"int64s.1=42",
			"uint32s.0=21",
			"uint32s.1=42",
			"uint64s.0=21",
			"uint64s.1=42",
			"doubles.0=2.1",
			"doubles.1=4.2",
			"bools.0=true",
			"bools.1=false",
			"enums.0=enum_value1",
			"enums.1=enum_value2",
		}, message)
		require.NoError(t, err)
		res, err := protojson.MarshalOptions{Indent: ""}.Marshal(message)
		require.NoError(t, err)
		assert.Equal(t, `{"strs":["abc1","abc2"],"int32s":[21,42],"int64s":["21","42"],"uint32s":[21,42],"uint64s":["21","42"],"doubles":[2.1,4.2],"bools":[true,false],"enums":["enum_value1","enum_value2"]}`, string(res))
	})

}
