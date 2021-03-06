package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	. "github.com/franela/goblin"
	"github.com/kristofferahl/go-centry/internal/pkg/io"
	test "github.com/kristofferahl/go-centry/internal/pkg/test"
)

func TestMain(t *testing.T) {
	g := Goblin(t)

	// Esuring the workdir is the root of the repo
	os.Chdir("../../")

	g.Describe("runtime", func() {
		g.It("returns error when manifest fails to load", func() {
			context := NewContext(CLI, io.Headless())
			runtime, err := NewRuntime([]string{}, context)
			g.Assert(runtime == nil).IsTrue("expected runtime to be nil")
			g.Assert(err != nil).IsTrue("expected error")
			g.Assert(strings.HasPrefix(err.Error(), "Failed to read manifest file.")).IsTrue("expected error message")
		})
	})

	g.Describe("scripts", func() {
		g.It("loads script in the expected order", func() {
			os.Setenv("OUTPUT_DEBUG", "true")
			g.Assert(strings.HasPrefix(execQuiet("get").Stdout, "Loading init.sh\nLoading helpers.sh\n")).IsTrue()
			os.Unsetenv("OUTPUT_DEBUG")
		})
	})

	g.Describe("options", func() {
		g.Describe("invoke without option", func() {
			g.It("should pass arguments", func() {
				g.Assert(execQuiet("test args foo bar").Stdout).Equal("test:args (foo bar)\n")
			})

			g.It("should have default value for environment variable set", func() {
				test.AssertKeyValueExists(g, "STRINGOPT", "foobar", execQuiet("test env").Stdout)
			})
		})

		g.Describe("invoke with single option", func() {
			g.It("should have arguments passed", func() {
				g.Assert(execQuiet("--staging test args foo bar").Stdout).Equal("test:args (foo bar)\n")
			})

			g.It("should have environment set to staging", func() {
				test.AssertKeyValueExists(g, "CONTEXT", "staging", execQuiet("--staging test env").Stdout)
			})

			g.It("should have environment set to last option with same env_name (production)", func() {
				test.AssertKeyValueExists(g, "CONTEXT", "production", execQuiet("--staging=false --production test env").Stdout)
			})

			g.It("should have environment set to last option with same env_name (staging)", func() {
				test.AssertKeyValueExists(g, "CONTEXT", "staging", execQuiet("--production=false --staging test env").Stdout)
			})
		})

		g.Describe("invoke with multiple options", func() {
			g.It("should have arguments passed", func() {
				g.Assert(execQuiet("--staging --stringopt=baz test args foo bar").Stdout).Equal("test:args (foo bar)\n")
			})

			g.It("should have multipe environment variables set", func() {
				out := execQuiet("--staging --stringopt=bazer --boolopt test env").Stdout
				test.AssertKeyValueExists(g, "CONTEXT", "staging", out)
				test.AssertKeyValueExists(g, "STRINGOPT", "bazer", out)
				test.AssertKeyValueExists(g, "BOOLOPT", "true", out)
			})
		})

		g.Describe("invoke with invalid option", func() {
			g.It("should return error", func() {
				res := execQuiet("--invalidoption test args foo bar")
				g.Assert(res.Stdout).Equal("")
				g.Assert(res.Error.Error()).Equal("flag provided but not defined: -invalidoption")
			})
		})
	})

	g.Describe("commands", func() {
		g.Describe("invoking command", func() {
			g.Describe("with arguments", func() {
				g.It("should have arguments passed", func() {
					g.Assert(execQuiet("get foo bar").Stdout).Equal("get (foo bar)\n")
				})
			})

			g.Describe("without arguments", func() {
				g.It("should have no arguments passed", func() {
					g.Assert(execQuiet("get").Stdout).Equal("get ()\n")
				})
			})
		})

		g.Describe("invoking sub command", func() {
			g.Describe("with arguments", func() {
				g.It("should have arguments passed", func() {
					g.Assert(execQuiet("get sub foo bar").Stdout).Equal("get:sub (foo bar)\n")
				})
			})

			g.Describe("without arguments", func() {
				g.It("should have no arguments passed", func() {
					g.Assert(execQuiet("get sub").Stdout).Equal("get:sub ()\n")
				})
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
				expected := `Commands:
    delete    Deletes stuff
    get       Gets stuff
    post      Creates stuff
    put       Creates/Updates stuff`

				g.Assert(strings.Contains(result.Stderr, expected)).IsTrue("\n\nEXPECTED:\n\n", expected, "\n\nTO BE FOUND IN:\n\n", result.Stderr)
			})

			g.It("should display global options", func() {
				expected := `Global options:
    --boolopt, -B         A custom option
    --config.log.level    Overrides the log level
    --help, -h            Displays help
    --production          Sets the context to production
    --quiet, -q           Disables logging
    --staging             Sets the context to staging
    --stringopt, -S       A custom option
    --version, -v         Displays the version of the cli`

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

	g.Describe("global options", func() {
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
	Source   string
	ExitCode int
	Error    error
	Stdout   string
	Stderr   string
}

func execQuiet(source string) *execResult {
	return execCentry(source, true)
}

func execWithLogging(source string) *execResult {
	return execCentry(source, false)
}

func execCentry(source string, quiet bool) *execResult {
	var exitCode int
	var runtimeErr error

	out := test.CaptureOutput(func() {
		if quiet {
			source = fmt.Sprintf("--quiet %s", source)
		}
		context := NewContext(CLI, io.Headless())
		runtime, err := NewRuntime(strings.Split(fmt.Sprintf("test/data/main_test.yaml %s", source), " "), context)
		if err != nil {
			exitCode = 1
			runtimeErr = err
		} else {
			exitCode = runtime.Execute()
		}
	})

	return &execResult{
		Source:   source,
		ExitCode: exitCode,
		Error:    runtimeErr,
		Stdout:   out.Stdout,
		Stderr:   out.Stderr,
	}
}
