package core

import (
	"context"
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
)

func Test_autocomplete(t *testing.T) {

	type TestCase struct {
		Suggestions []string
	}

	fileDescSet := descriptorpb.FileDescriptorSet{}
	err := proto.Unmarshal(rawProto, &fileDescSet)
	require.NoError(t, err)
	files, err := protodesc.NewFiles(&fileDescSet)
	require.NoError(t, err)

	ctx := context.Background()

	run := func(tc TestCase) func(t *testing.T) {
		return func(t *testing.T) {
			name := strings.Replace(t.Name(), "Test_autocomplete/", "", -1)
			name = strings.Replace(name, "_", " ", -1)
			words := strings.Split(name, " ")
			wordToCompleteIndex := len(words) - 1
			leftWords := words[:wordToCompleteIndex]
			wordToComplete := words[wordToCompleteIndex]
			rightWord := words[wordToCompleteIndex+1:]

			suggestion := Autocomplete(ctx, files, leftWords, wordToComplete, rightWord)
			assert.Equal(t, tc.Suggestions, suggestion)

		}
	}

	t.Run("grpc-cli ", run(TestCase{Suggestions: []string{"rpc"}}))
	t.Run("grpc-cli rpc tes", run(TestCase{Suggestions: []string{"test.Api"}}))
	t.Run("grpc-cli rpc test.Api ", run(TestCase{Suggestions: []string{"Echo"}}))
	t.Run("grpc-cli rpc test.Api Echo s", run(TestCase{Suggestions: []string{"str=", "strs="}}))
	t.Run("grpc-cli rpc test.Api Echo u", run(TestCase{Suggestions: []string{"uint32=", "uint64="}}))
}
