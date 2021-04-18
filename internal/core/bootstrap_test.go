package core

import (
	_ "embed"
	"testing"
)

//go:embed testdata/test.pb
var rawProto []byte

func TestBootstrap_usage(t *testing.T) {

	t.Run("root", Test(&TestConfig{
		Descriptor: rawProto,
		Cmd:        "grpc-cli -h",
		Check:      TestCheckGolden(),
	}))

	t.Run("rpc", Test(&TestConfig{
		Descriptor: rawProto,
		Cmd:        "grpc-cli rpc -h",
		Check:      TestCheckGolden(),
	}))

	t.Run("rpc test.Api", Test(&TestConfig{
		Descriptor: rawProto,
		Cmd:        "grpc-cli rpc test.Api -h",
		Check:      TestCheckGolden(),
	}))

	t.Run("rpc test.Api Echo", Test(&TestConfig{
		Descriptor: rawProto,
		Cmd:        "grpc-cli rpc test.Api Echo -h",
		Check:      TestCheckGolden(),
	}))

}
