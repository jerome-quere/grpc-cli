package core

import (
	"context"
	"fmt"
	"strconv"
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

	if len(leftWords) > 3 {
		return autocompleteParams(files, leftWords[2], leftWords[3], wordToComplete)
	}

	return nil
}

func autocompleteParams(files *protoregistry.Files, serviceName string, methodName string, wordToComplete string) []string {
	desc, err := files.FindDescriptorByName(protoreflect.FullName(serviceName + "." + methodName))
	if err != nil {
		return nil
	}
	method := desc.(protoreflect.MethodDescriptor)
	reqDesc := method.Input()

	fields := reqDesc.Fields()
	res := []string(nil)

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if strings.HasPrefix(string(field.Name()), wordToComplete) {
			res = append(res, string(field.Name())+"=")
		}
	}
	return res
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

func AutocompleteCobraCommand(ctx context.Context, files *protoregistry.Files) *cobra.Command {
	cmd := &cobra.Command{
		Use: "autocomplete",
	}

	completeCmd := &cobra.Command{
		Use:    "complete",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {

			case "bash":
				wordIndex, err := strconv.Atoi(args[1])
				if err != nil {
					return err
				}
				words := args[2:]
				leftWords := words[:wordIndex]
				wordToComplete := words[wordIndex]
				rightWords := words[wordIndex+1:]

				// If the wordToComplete is an argument label (cf. `arg=`), remove
				// this prefix for all suggestions.
				res := Autocomplete(ctx, files, leftWords, wordToComplete, rightWords)
				if strings.Contains(wordToComplete, "=") {
					prefix := strings.SplitAfterN(wordToComplete, "=", 2)[0]
					for k, p := range res {
						res[k] = strings.TrimPrefix(p, prefix)
					}
				}

				fmt.Println(strings.Join(res, "\n"))
				return nil
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
			case "zsh":
				// First arg is the word index.
				wordIndex, err := strconv.Atoi(args[0])
				if err != nil {
					return err
				}
				wordIndex-- // In zsh word index starts at 1.

				// Other args are all the words.
				words := args[1:]
				if len(words) <= wordIndex {
					words = append(words, "") // Handle case when last word is empty.
				}

				leftWords := words[:wordIndex]
				wordToComplete := words[wordIndex]
				rightWords := words[wordIndex+1:]

				res := Autocomplete(ctx, files, leftWords, wordToComplete, rightWords)
				fmt.Println(strings.Join(res, "\n"))
				return nil

			default:
				return fmt.Errorf("unsuported shell type %s", args[0])
			}
		},
	}
	cmd.AddCommand(completeCmd)

	scriptCmd := &cobra.Command{
		Use:    "script",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			binaryName := CtxBinaryName(ctx)
			switch args[0] {
			case "bash":
				fmt.Printf(`
_%[1]s() {
	_get_comp_words_by_ref -n = cword words

	output=$(%[1]s autocomplete complete bash -- "$COMP_LINE" "$cword" "${words[@]}")
	COMPREPLY=($output)
	# apply compopt option and ignore failure for older bash versions
	[[ $COMPREPLY == *= ]] && compopt -o nospace 2> /dev/null || true
	return
}
complete -F _%[1]s %[1]s
`, binaryName)

			case "fish":
				fmt.Printf(`
complete --erase --command %[1]s;
complete --command %[1]s --no-files;
complete --command %[1]s --arguments '(%[1]s autocomplete complete fish -- (commandline) (commandline --cursor) (commandline --current-token) (commandline --current-process --tokenize --cut-at-cursor))';
`, binaryName)

			case "zsh":
				fmt.Printf(`
autoload -U compinit && compinit
_%[1]s () {
	output=($(%[1]s autocomplete complete zsh -- ${CURRENT} ${words}))
	opts=('-S' ' ')
	if [[ $output == *= ]]; then
		opts=('-S' '')
	fi
	compadd "${opts[@]}" -- "${output[@]}"
}
compdef _%[1]s %[1]s
`, binaryName)

			default:
				return fmt.Errorf("unsuported shell type %s", args[0])
			}
			return nil
		},
	}
	cmd.AddCommand(scriptCmd)

	return cmd
}
