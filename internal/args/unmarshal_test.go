package args

import (
	"bytes"
	_ "embed"
	"strings"
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

	run := func(args []string, expected string) {
		desc, err := files.FindDescriptorByName("test.Simple")
		require.NoError(t, err)
		message := dynamicpb.NewMessage(desc.(protoreflect.MessageDescriptor))
		err = Unmarshal(args, message)
		require.NoError(t, err)

		res, err := protojson.MarshalOptions{
			UseProtoNames: true,
			Indent:        "\t",
			Multiline:     true,
		}.Marshal(message)
		require.NoError(t, err)

		// Don't know why but protojson sometime add extra space in the json output
		// To make result consitent we remove all double space
		res = bytes.ReplaceAll(res, []byte("  "), []byte(" "))

		assert.Equal(t, formatExpected(expected), string(res))
	}

	t.Run("Simple", func(t *testing.T) {
		run([]string{
			"str=abc",
			"int32=32",
			"int64=64",
			"uint32=32",
			"uint64=64",
			"double=6.4",
			"bool=true",
			"enum=enum_value2",
			"nested.str=nested_abc",
			"nested.strs.0=nested_str_1",
			"nested.strs.1=nested_str_2",
			"wrapper_str=wrapper_str",
		}, `
			{
				"str": "abc",
				"int32": 32,
				"int64": "64",
				"uint32": 32,
				"uint64": "64",
				"double": 6.4,
				"bool": true,
				"enum": "enum_value2",
				"nested": {
					"str": "nested_abc",
					"strs": [
						"nested_str_1",
						"nested_str_2"
					]
				},
				"wrapper_str": "wrapper_str"
			}
		`)
	})

	t.Run("Repeated", func(t *testing.T) {

		run([]string{
			"strs.0=abc1",
			"strs.1=abc2",
			"enums.0=enum_value1",
			"enums.1=enum_value2",
			"nesteds.0.str=nested_abc_1",
			"nesteds.1.str=nested_abc_2",
			"nesteds.0.strs.0=nested_abc_nested_1",
			"nesteds.1.strs.0=nested_abc_nested_2",
			"wrapper_strs.0=wrapper_str_1",
			"wrapper_strs.1=wrapper_str_2",
		}, `
			{
				"strs": [
					"abc1",
					"abc2"
				],
				"enums": [
					"enum_value1",
					"enum_value2"
				],
				"nesteds": [
					{
						"str": "nested_abc_1",
						"strs": [
							"nested_abc_nested_1"
						]
					},
					{
						"str": "nested_abc_2",
						"strs": [
							"nested_abc_nested_2"
						]
					}
				],
				"wrapper_strs": [
					"wrapper_str_1",
					"wrapper_str_2"
				]
			}
		`)
	})

}

func formatExpected(str string) string {
	// Remove First \n
	var res []string
	lines := strings.Split(str, "\n")
	lines = lines[1 : len(lines)-1]
	indent := strings.Index(lines[0], "{")
	for _, line := range lines {
		res = append(res, line[indent:])
	}
	return strings.Join(res, "\n")
}
