package main

import (
	"github.com/kristofferahl/go-centry/internal/pkg/config"
	"github.com/kristofferahl/go-centry/internal/pkg/io"
	"github.com/kristofferahl/go-centry/internal/pkg/log"
)

// Executor is the name of the executor
type Executor string

// CLI executor
var CLI Executor = "CLI"

// API Executor
var API Executor = "API"

// Context defines the current context
type Context struct {
	executor           Executor
	io                 io.InputOutput
	log                *log.Manager
	manifest           *config.Manifest
	commandEnabledFunc func(config.Command) bool
	optionEnabledFunc  func(config.Option) bool
}

// NewContext creates a new context
func NewContext(executor Executor, io io.InputOutput) *Context {
	return &Context{
		executor: executor,
		io:       io,
	}
}
