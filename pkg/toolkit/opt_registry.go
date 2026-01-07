// Copyright (c) 2025 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package toolkit

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// CommandFunc defines a function type for registered commands.
type CommandFunc func(args ...string) error

// registry holds registered command functions.
var (
	commandRegistry = make(map[string]CommandFunc)
	mu              sync.RWMutex
)

// RegisterCommand registers a command function with a unique name.
// Returns an error if the name is already used.
func RegisterCommand(name string, fn CommandFunc) error {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := commandRegistry[name]; exists {
		return fmt.Errorf("command %s already registered", name)
	}

	commandRegistry[name] = fn
	return nil
}

// ExecuteCommand executes a registered command by name with arguments.
// Returns error if command is not found or function itself returns error.
func ExecuteCommand(name string, args ...string) (error, bool) {
	if strings.HasPrefix(name, "-") {
		return nil, false
	}

	mu.RLock()
	fn, ok := commandRegistry[name]
	mu.RUnlock()

	if !ok {
		return fmt.Errorf("command %s not found", name), false
	}

	return fn(args...), true
}

// ListCommands returns a list of all registered command names.
func ListCommands() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(commandRegistry))
	for k := range commandRegistry {
		names = append(names, k)
	}
	return names
}

func RunAdditional() (error, bool) {
	args := os.Args[1:]
	if len(args) == 0 {
		return nil, false // no additional command
	}

	cmdName := args[0]
	cmdArgs := args[1:]

	// check if command is registered
	err, ok := ExecuteCommand(cmdName, cmdArgs...)
	if err != nil {
		return fmt.Errorf("failed to execute command '%s': %w", cmdName, err), false
	}

	return nil, ok
}
