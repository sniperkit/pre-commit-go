// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package internal contains internal code used by multiple packages within
// pre-commit-go.
//
// It contains a reimplementation of os.Remove() that works even with files
// with the read-only bit on Windows.
package internal

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Capture runs an executable from a directory returns the output, exit code
// and error if appropriate. It sets the environment variables specified.
func Capture(wd string, env []string, args ...string) (string, int, error) {
	exitCode := -1
	//log.Printf("Capture(%s, %s, %s)", wd, env, args)
	var c *exec.Cmd
	switch len(args) {
	case 0:
		return "", -1, errors.New("no command specified")
	case 1:
		c = exec.Command(args[0])
	default:
		c = exec.Command(args[0], args[1:]...)
	}
	if wd == "" {
		return "", -1, errors.New("wd is required")
	}
	c.Dir = wd
	procEnv := map[string]string{}
	for _, item := range os.Environ() {
		items := strings.SplitN(item, "=", 2)
		procEnv[items[0]] = items[1]
	}
	procEnv["LANG"] = "en_US.UTF-8"
	procEnv["LANGUAGE"] = "en_US.UTF-8"
	for _, item := range env {
		items := strings.SplitN(item, "=", 2)
		if items[1] == "" {
			// Remove it instead.
			delete(procEnv, items[0])
		} else {
			procEnv[items[0]] = items[1]
		}
	}
	c.Env = make([]string, 0, len(procEnv))
	for k, v := range procEnv {
		c.Env = append(c.Env, k+"="+v)
	}
	out, err := c.CombinedOutput()
	if c.ProcessState != nil {
		if waitStatus, ok := c.ProcessState.Sys().(syscall.WaitStatus); ok {
			exitCode = waitStatus.ExitStatus()
			if exitCode != 0 {
				err = nil
			}
		}
	}
	// TODO(maruel): Handle code page on Windows.
	return string(out), exitCode, err
}
