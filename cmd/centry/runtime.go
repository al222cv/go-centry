package main

import (
	"strings"

	"github.com/kristofferahl/cli"
	"github.com/kristofferahl/go-centry/pkg/config"
	"github.com/kristofferahl/go-centry/pkg/log"
	"github.com/kristofferahl/go-centry/pkg/shell"
	"github.com/sirupsen/logrus"
)

// Runtime defines the runtime
type Runtime struct {
	context *Context
	cli     *cli.CLI
}

// NewRuntime builds a runtime based on the given arguments
func NewRuntime(inputArgs []string, context *Context) *Runtime {
	// Create the runtime
	runtime := &Runtime{}

	// Args
	file := ""
	args := []string{}
	if len(inputArgs) >= 1 {
		file = inputArgs[0]
		args = inputArgs[1:]
	}

	// Load manifest
	context.manifest = config.LoadManifest(file)

	// Create the log manager
	context.log = log.CreateManager(context.manifest.Config.Log.Level, context.manifest.Config.Log.Prefix, context.io)

	// Create global options
	options := createGlobalOptions(context.manifest)

	// Parse global options to get cli args
	args = options.Parse(args)

	// Initialize cli
	c := &cli.CLI{
		Name:    context.manifest.Config.Name,
		Version: context.manifest.Config.Version,

		Commands:   map[string]cli.CommandFactory{},
		Args:       args,
		HelpFunc:   cliHelpFunc(context.manifest, options),
		HelpWriter: context.io.Stderr,

		// Autocomplete:          true,
		// AutocompleteInstall:   "install-autocomplete",
		// AutocompleteUninstall: "uninstall-autocomplete",
	}

	// Override the current log level from options
	logLevel := options.GeString("config.log.level")
	if options.GetBool("quiet") {
		logLevel = "panic"
	}
	context.log.TrySetLogLevel(logLevel)

	logger := context.log.GetLogger()

	// Register builtin commands
	if context.executor == CLI {
		c.Commands["serve"] = func() (cli.Command, error) {
			return &ServeCommand{
				Manifest: context.manifest,
				Log: logger.WithFields(logrus.Fields{
					"command": "serve",
				}),
			}, nil
		}
	}

	// Build commands
	for _, cmd := range context.manifest.Commands {
		cmd := cmd

		if context.commandEnabled != nil && context.commandEnabled(cmd) == false {
			continue
		}

		script := createScript(cmd, context)

		funcs, err := script.Functions()
		if err != nil {
			logger.WithFields(logrus.Fields{
				"command": cmd.Name,
			}).Fatal(err)
		}

		for _, fn := range funcs {
			namespace := script.CreateFunctionNamespace(cmd.Name)

			if fn != cmd.Name && strings.HasPrefix(fn, namespace) == false {
				continue
			}

			cmdKey := strings.Replace(fn, script.FunctionNameSplitChar(), " ", -1)
			c.Commands[cmdKey] = func() (cli.Command, error) {
				return &ScriptCommand{
					Context:       context,
					Log:           logger.WithFields(logrus.Fields{}),
					GlobalOptions: options,
					Command:       cmd,
					Script:        script,
					Function:      fn,
				}, nil
			}

			logger.Debugf("Registered command \"%s\"", cmdKey)
		}
	}

	runtime.context = context
	runtime.cli = c

	return runtime
}

// Execute runs the CLI and exits with a code
func (runtime *Runtime) Execute() int {
	// Run cli
	exitCode, err := runtime.cli.Run()
	if err != nil {
		logger := runtime.context.log.GetLogger()
		logger.Error(err)
	}

	return exitCode
}

func createScript(cmd config.Command, context *Context) shell.Script {
	return &shell.BashScript{
		BasePath: context.manifest.BasePath,
		Path:     cmd.Path,
	}
}