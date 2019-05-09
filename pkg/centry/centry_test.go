package centry

import (
	"fmt"
	"os"
	"strings"
	"testing"

	. "github.com/franela/goblin"
	th "github.com/kristofferahl/go-centry/pkg/testing"
)

func TestMain(t *testing.T) {
	g := Goblin(t)

	g.Describe("scripts", func() {
		g.It("loads script in the expected order", func() {
			os.Setenv("OUTPUT_DEBUG", "true")
			g.Assert(strings.HasPrefix(execQuiet("get").Stdout, "Loading init.sh\nLoading helpers.sh\n")).IsTrue()
			os.Unsetenv("OUTPUT_DEBUG")
		})
	})

	g.Describe("commands", func() {
		g.Describe("call without argument", func() {
			g.It("should have no arguments passed", func() {
				g.Assert(execQuiet("get").Stdout).Equal("get ()\n")
			})
		})

		g.Describe("call with argument", func() {
			g.It("should have arguments passed", func() {
				g.Assert(execQuiet("get foobar").Stdout).Equal("get (foobar)\n")
			})
		})
	})

	g.Describe("help", func() {
		g.Describe("call with no arguments", func() {
			result := execQuiet("")

			g.It("should display help", func() {
				expected := `Usage: centry`
				g.Assert(strings.Contains(result.Stderr, expected)).IsTrue()
			})
		})

		g.Describe("call with -h", func() {
			result := execQuiet("-h")

			g.It("should display help", func() {
				expected := `Usage: centry`
				g.Assert(strings.Contains(result.Stderr, expected)).IsTrue()
			})
		})

		g.Describe("call with --help", func() {
			result := execQuiet("--help")

			g.It("should display help", func() {
				expected := `Usage: centry`
				g.Assert(strings.Contains(result.Stderr, expected)).IsTrue()
			})
		})

		g.Describe("output", func() {
			result := execQuiet("")

			g.It("should display available commands", func() {
				expected := `Available commands are:
    delete    Deletes stuff
    get       Gets stuff
    put       Creates stuff`

				g.Assert(strings.Contains(result.Stderr, expected)).IsTrue("\n\nEXPECTED:\n\n", expected, "\n\nTO BE FOUND IN:\n\n", result.Stderr)
			})

			g.It("should display global options", func() {
				expected := `Global options are:
       | --config.log.level    Overrides the log level
       | --custom              A custom option with default value
    -h | --help                Displays help
    -q | --quiet               Disables logging
    -v | --version             Displays the version fo the cli`

				g.Assert(strings.Contains(result.Stderr, expected)).IsTrue("\n\nEXPECTED:\n\n", expected, "\n\nTO BE FOUND IN:\n\n", result.Stderr)
			})
		})

		g.Describe("call without arguments", func() {
			result := execQuiet("")

			g.It("should display help text", func() {
				g.Assert(strings.Contains(result.Stderr, "Usage: centry")).IsTrue(result.Stderr)
			})
		})
	})

	g.Describe("version", func() {
		g.Describe("--version", func() {
			result := execQuiet("--version")

			g.It("should display version", func() {
				expected := `1.0.0`
				g.Assert(strings.Contains(result.Stderr, expected)).IsTrue()
			})
		})

		g.Describe("-v", func() {
			result := execQuiet("-v")

			g.It("should display version", func() {
				expected := `1.0.0`
				g.Assert(strings.Contains(result.Stderr, expected)).IsTrue()
			})
		})
	})

	g.Describe("global options", func() {
		g.Describe("quiet", func() {
			g.Describe("--quiet", func() {
				result := execWithLogging("--quiet")

				g.It("should disable logging", func() {
					expected := `Changing loglevel to panic (from debug)`
					g.Assert(strings.Contains(result.Stderr, expected)).IsTrue(result.Stderr)
				})
			})

			g.Describe("-q", func() {
				result := execWithLogging("-q")

				g.It("should disable logging", func() {
					expected := `Changing loglevel to panic (from debug)`
					g.Assert(strings.Contains(result.Stderr, expected)).IsTrue(result.Stderr)
				})
			})
		})

		g.Describe("--config.log.level=info", func() {
			result := execWithLogging("--config.log.level=info")

			g.It("should change log level to info", func() {
				expected := `Changing loglevel to info (from debug)`
				g.Assert(strings.Contains(result.Stderr, expected)).IsTrue(result.Stderr)
			})
		})

		g.Describe("--config.log.level=error", func() {
			result := execWithLogging("--config.log.level=error")

			g.It("should change log level to error", func() {
				expected := `Changing loglevel to error (from debug)`
				g.Assert(strings.Contains(result.Stderr, expected)).IsTrue(result.Stderr)
			})
		})
	})
}

type execResult struct {
	Source         string
	Stdout         string
	Stderr         string
	CombinedOutput string
}

func execQuiet(source string) *execResult {
	return execCentry(source, true)
}

func execWithLogging(source string) *execResult {
	return execCentry(source, false)
}

func execCentry(source string, quiet bool) *execResult {
	out := th.CaptureOutput(func() {
		if quiet {
			source = fmt.Sprintf("--quiet %s", source)
		}
		RunOnce(strings.Split(fmt.Sprintf("./centry ../../test/data/main_test.yaml %s", source), " "))
	})

	return &execResult{
		Source:         source,
		Stdout:         out.Stdout,
		Stderr:         out.Stderr,
		CombinedOutput: fmt.Sprintf("\nStdout:\n%s\nStderr:\n%s", out.Stderr, out.Stdout),
	}
}