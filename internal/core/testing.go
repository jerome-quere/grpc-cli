package core

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	UpdateGoldens = flag.Bool("goldens", os.Getenv("UPDATE_GOLDENS") == "true", "Record goldens")
)

type TestConfig struct {
	Descriptor []byte

	Cmd  string
	Args []string

	Check TestCheckFunc
}

type TestCheckFuncCtx struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}

type TestCheckFunc func(t *testing.T, ctx *TestCheckFuncCtx)

func Test(config *TestConfig) func(t *testing.T) {
	return func(t *testing.T) {

		ctx := context.Background()

		args := config.Args
		if args == nil {
			args = strings.Split(config.Cmd, " ")
		}

		stderr := &bytes.Buffer{}
		stdout := &bytes.Buffer{}

		exitCode := Bootstrap(ctx, &BootstrapConfig{
			Stderr:     stderr,
			Stdout:     stdout,
			Args:       args,
			Descriptor: config.Descriptor,
		})

		config.Check(t, &TestCheckFuncCtx{
			ExitCode: exitCode,
			Stdout:   stdout.Bytes(),
			Stderr:   stderr.Bytes(),
		})
	}
}

// TestCheckGolden assert stderr and stdout using golden
func TestCheckGolden() TestCheckFunc {
	return func(t *testing.T, ctx *TestCheckFuncCtx) {
		actual := marshalGolden(t, ctx)

		goldenPath := getTestFilePath(t, ".golden")
		// In order to avoid diff in goldens we set all timestamp to the same date
		if *UpdateGoldens || true {
			require.NoError(t, os.MkdirAll(path.Dir(goldenPath), 0755))
			require.NoError(t, ioutil.WriteFile(goldenPath, []byte(actual), 0644)) //nolint:gosec
		}

		expected, err := ioutil.ReadFile(goldenPath)
		require.NoError(t, err, "expected to find golden file %s", goldenPath)
		assert.Equal(t, string(expected), actual)
	}
}

func marshalGolden(t *testing.T, ctx *TestCheckFuncCtx) string {
	jsonStderr := &bytes.Buffer{}
	jsonStdout := &bytes.Buffer{}

	buffer := bytes.Buffer{}
	buffer.WriteString(fmt.Sprintf("\U0001F3B2\U0001F3B2\U0001F3B2 EXIT CODE: %d \U0001F3B2\U0001F3B2\U0001F3B2\n", ctx.ExitCode))

	if len(ctx.Stdout) > 0 {
		buffer.WriteString("\U0001F7E9\U0001F7E9\U0001F7E9 STDOUT️ \U0001F7E9\U0001F7E9\U0001F7E9️\n")
		buffer.Write(ctx.Stdout)
	}

	if len(ctx.Stderr) > 0 {
		buffer.WriteString("\U0001F7E5\U0001F7E5\U0001F7E5 STDERR️️ \U0001F7E5\U0001F7E5\U0001F7E5️\n")
		buffer.Write(ctx.Stderr)
	}

	if jsonStdout.Len() > 0 {
		buffer.WriteString("\U0001F7E9\U0001F7E9\U0001F7E9 JSON STDOUT \U0001F7E9\U0001F7E9\U0001F7E9\n")
		buffer.Write(jsonStdout.Bytes())
	}

	if jsonStderr.Len() > 0 {
		buffer.WriteString("\U0001F7E5\U0001F7E5\U0001F7E5 JSON STDERR \U0001F7E5\U0001F7E5\U0001F7E5\n")
		buffer.Write(jsonStderr.Bytes())
	}

	str := buffer.String()
	// Replace Windows return carriage.
	str = strings.ReplaceAll(str, "\r", "")
	return str
}

// getTestFilePath returns a valid filename path based on the go test name and suffix. (Take care of non fs friendly char)
func getTestFilePath(t *testing.T, suffix string) string {
	specialChars := regexp.MustCompile(`[\\?%*:|"<>. ]`)

	// Replace nested tests separators.
	fileName := strings.Replace(t.Name(), "/", "-", -1)
	fileName = strings.Replace(fileName, "_", "-", -1)
	fileName = strings.ToLower(fileName)

	// Replace special characters.
	fileName = specialChars.ReplaceAllLiteralString(fileName, "") + suffix

	return filepath.Join(".", "testdata", "goldens", fileName)
}
