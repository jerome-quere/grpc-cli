package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func Autocomplete(ctx context.Context, files *protoregistry.Files, leftWords []string, wordToComplete string, rightWords []string) []string {
	// We only autocomplete rpc command
	if len(leftWords) == 1 {
		return autocompleteFilter([]string{"rpc"}, wordToComplete)
	}
	if leftWords[1] != "rpc" {
		return nil
	}

	if len(leftWords) == 2 {
		return autocompleteService(files, wordToComplete)
	}
	if len(leftWords) == 3 {
		return autocompleteMethod(files, leftWords[2], wordToComplete)
	}

	return nil
}

func autocompleteFilter(strs []string, wordToComplete string) []string {
	res := []string(nil)
	for _, s := range strs {
		if strings.HasPrefix(s, wordToComplete) {
			res = append(res, s)
		}
	}
	return res
}

func autocompleteMethod(files *protoregistry.Files, serviceName string, wordToComplete string) []string {
	desc, err := files.FindDescriptorByName(protoreflect.FullName(serviceName))
	if err != nil {
		return nil
	}
	res := []string(nil)
	service := desc.(protoreflect.ServiceDescriptor)
	for i := 0; i < service.Methods().Len(); i++ {
		method := service.Methods().Get(i)
		methodName := string(method.Name())
		if strings.HasPrefix(methodName, wordToComplete) {
			res = append(res, methodName)
		}
	}
	return res
}

func autocompleteService(files *protoregistry.Files, wordToComplete string) []string {
	res := []string(nil)
	files.RangeFiles(func(descriptor protoreflect.FileDescriptor) bool {
		for i := 0; i < descriptor.Services().Len(); i++ {
			service := descriptor.Services().Get(i)
			serviceName := string(service.FullName())
			if strings.HasPrefix(serviceName, wordToComplete) {
				res = append(res, serviceName)
			}
		}
		return true
	})
	return res
}

func AutocompleteCommand(ctx context.Context, files *protoregistry.Files) *cobra.Command {
	cmd := &cobra.Command{
		Use: "autocomplete",
	}

	completeCmd := &cobra.Command{
		Use:    "complete",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "fish":
				leftWords := args[4:]
				wordToComplete := args[3]
				rightWords := []string(nil) // TODO

				res := Autocomplete(ctx, files, leftWords, wordToComplete, rightWords)

				// TODO: decide if we want to add descriptions
				// see https://stackoverflow.com/a/20879411
				// "followed optionally by a tab and a short description."
				fmt.Println(strings.Join(res, "\n"))
				return nil
			}

			fmt.Println(args)

			// TODO
			return nil
		},
	}
	cmd.AddCommand(completeCmd)

	scriptCmd := &cobra.Command{
		Use:    "script",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			binaryName := CtxBinaryName(ctx)
			switch args[0] {
			case "fish":
				fmt.Printf(`
complete --erase --command %[1]s;
complete --command %[1]s --no-files;
complete --command %[1]s --arguments '(%[1]s autocomplete complete fish -- (commandline) (commandline --cursor) (commandline --current-token) (commandline --current-process --tokenize --cut-at-cursor))';
		`, binaryName)
			}
			return nil
		},
	}
	cmd.AddCommand(scriptCmd)

	return cmd
}
