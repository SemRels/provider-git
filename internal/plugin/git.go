// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The provider-git Authors

package plugin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const defaultRemote = "origin"

// execTextFileBusyRetries bounds how many times ExecRunner.Run retries a
// command that fails immediately with ETXTBSY. This race is a well-known
// Go/OS quirk (see golang/go#22315, golang/go#62221): a freshly written
// executable can transiently report "text file busy" if it is exec'd before
// the OS/filesystem (notably overlayfs, as used by GitHub Actions/Docker
// runners) has finished committing the write. A short bounded retry with
// backoff is the standard workaround recommended by the Go team.
const execTextFileBusyRetries = 5

const execTextFileBusyBackoff = 20 * time.Millisecond

// Config controls how git push operations are executed.
type Config struct {
	Remote string
}

// Runner executes subprocess commands.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) error
}

// ExecRunner runs commands with os/exec in an optional working directory.
type ExecRunner struct {
	Dir string
}

func (r ExecRunner) Run(ctx context.Context, name string, args ...string) error {
	var (
		output []byte
		err    error
	)

	for attempt := 0; ; attempt++ {
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = r.Dir

		output, err = cmd.CombinedOutput()
		if err == nil {
			return nil
		}

		if !isTextFileBusy(err) || attempt >= execTextFileBusyRetries {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(execTextFileBusyBackoff):
		}
	}

	command := strings.TrimSpace(strings.Join(append([]string{name}, args...), " "))
	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Errorf("run %q: %w", command, err)
	}

	return fmt.Errorf("run %q: %w: %s", command, err, message)
}

// isTextFileBusy reports whether err is (or wraps) a syscall.ETXTBSY error,
// which os/exec surfaces when a freshly written executable is launched
// before the OS has released its write handle on the file.
func isTextFileBusy(err error) bool {
	var errno syscall.Errno

	return errors.As(err, &errno) && errno == syscall.ETXTBSY
}

// Client pushes SemRel-generated refs to a configured git remote.
type Client struct {
	config Config
	runner Runner
}

func ConfigFromEnv(lookupEnv func(string) string) Config {
	if lookupEnv == nil {
		lookupEnv = os.Getenv
	}

	remote := strings.TrimSpace(lookupEnv("SEMREL_PLUGIN_REMOTE"))
	if remote == "" {
		remote = strings.TrimSpace(lookupEnv("SEMREL_REMOTE"))
	}
	if remote == "" {
		remote = defaultRemote
	}

	return Config{Remote: remote}
}

func NewClient(config Config, runner Runner) *Client {
	if strings.TrimSpace(config.Remote) == "" {
		config.Remote = defaultRemote
	}

	if runner == nil {
		runner = ExecRunner{}
	}

	return &Client{config: config, runner: runner}
}

func (c *Client) PushTag(ctx context.Context, tagName, remote string) error {
	tagName = strings.TrimSpace(tagName)
	if tagName == "" {
		return errors.New("tag name is required")
	}

	return c.runner.Run(ctx, "git", "push", c.resolveRemote(remote), tagName)
}

func (c *Client) PushBranch(ctx context.Context, branch, remote string) error {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return errors.New("branch is required")
	}

	return c.runner.Run(ctx, "git", "push", c.resolveRemote(remote), "HEAD:"+branch)
}

func (c *Client) resolveRemote(remote string) string {
	remote = strings.TrimSpace(remote)
	if remote != "" {
		return remote
	}

	remote = strings.TrimSpace(c.config.Remote)
	if remote != "" {
		return remote
	}

	return defaultRemote
}
