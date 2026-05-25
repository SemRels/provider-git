// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The provider-git Authors

package plugin

import (
"context"
"fmt"
"os"
"os/exec"
"path/filepath"
"runtime"
"strings"
"testing"

"github.com/stretchr/testify/require"
)

func TestConfigFromEnvDefaultsToOrigin(t *testing.T) {
t.Parallel()

config := ConfigFromEnv(func(string) string { return "" })

require.Equal(t, "origin", config.Remote)
}

func TestConfigFromEnvUsesConfiguredRemote(t *testing.T) {
t.Parallel()

config := ConfigFromEnv(func(key string) string {
if key == "SEMREL_PLUGIN_REMOTE" {
return "upstream"
}
return ""
})
require.Equal(t, "upstream", config.Remote)
}

func TestPushTagPushesTagToRemoteRepository(t *testing.T) {
t.Parallel()

repoDir, remoteDir := createGitRepositories(t)
client := NewClient(Config{Remote: "origin"}, ExecRunner{Dir: repoDir})

require.NoError(t, client.PushTag(context.Background(), "v1.2.3", ""))
require.NoError(t, gitInDir(remoteDir, "show-ref", "--verify", "refs/tags/v1.2.3").Run())
}

func TestPushBranchPushesHeadToRequestedBranch(t *testing.T) {
t.Parallel()

repoDir, remoteDir := createGitRepositories(t)
client := NewClient(Config{Remote: "origin"}, ExecRunner{Dir: repoDir})

require.NoError(t, client.PushBranch(context.Background(), "release", ""))
require.NoError(t, gitInDir(remoteDir, "show-ref", "--verify", "refs/heads/release").Run())
}

func TestPushTagReturnsValidationError(t *testing.T) {
t.Parallel()

client := NewClient(Config{}, runnerFunc(func(context.Context, string, ...string) error { return nil }))

err := client.PushTag(context.Background(), "", "")

require.EqualError(t, err, "tag name is required")
}

func TestPushBranchReturnsValidationError(t *testing.T) {
t.Parallel()

client := NewClient(Config{}, runnerFunc(func(context.Context, string, ...string) error { return nil }))

err := client.PushBranch(context.Background(), "", "")

require.EqualError(t, err, "branch is required")
}

func TestPushBranchUsesExplicitRemoteOverride(t *testing.T) {
t.Parallel()

var got []string
client := NewClient(Config{Remote: "origin"}, runnerFunc(func(_ context.Context, name string, args ...string) error {
got = append([]string{name}, args...)
return nil
}))

require.NoError(t, client.PushBranch(context.Background(), "main", "fork"))
require.Equal(t, []string{"git", "push", "fork", "HEAD:main"}, got)
}

func TestExecRunnerIncludesCommandOutputInErrors(t *testing.T) {
t.Parallel()

dir := createTestDir(t)
err := ExecRunner{Dir: dir}.Run(context.Background(), "git", "not-a-command")

require.Error(t, err)
require.Contains(t, err.Error(), "git not-a-command")
}

type runnerFunc func(ctx context.Context, name string, args ...string) error

func (f runnerFunc) Run(ctx context.Context, name string, args ...string) error {
return f(ctx, name, args...)
}

func createGitRepositories(t *testing.T) (string, string) {
t.Helper()

baseDir := createTestDir(t)
repoDir := filepath.Join(baseDir, "repo")
remoteDir := filepath.Join(baseDir, "remote.git")

require.NoError(t, os.MkdirAll(repoDir, 0o755))
require.NoError(t, gitInDir(baseDir, "init", "--bare", remoteDir).Run())
require.NoError(t, gitInDir(baseDir, "init", repoDir).Run())
require.NoError(t, gitInDir(repoDir, "config", "user.email", "test@example.com").Run())
require.NoError(t, gitInDir(repoDir, "config", "user.name", "SemRel Test").Run())
require.NoError(t, os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("hello\n"), 0o644))
require.NoError(t, gitInDir(repoDir, "add", "README.md").Run())
require.NoError(t, gitInDir(repoDir, "commit", "-m", "initial").Run())
require.NoError(t, gitInDir(repoDir, "tag", "v1.2.3").Run())
require.NoError(t, gitInDir(repoDir, "remote", "add", "origin", remoteDir).Run())

return repoDir, remoteDir
}

func createTestDir(t *testing.T) string {
t.Helper()

wd, err := os.Getwd()
require.NoError(t, err)

baseDir, err := os.MkdirTemp(wd, "git-test-")
require.NoError(t, err)
t.Cleanup(func() {
require.NoError(t, os.RemoveAll(baseDir))
})

return baseDir
}

func gitInDir(dir string, args ...string) *exec.Cmd {
cmd := exec.Command("git", args...)
cmd.Dir = dir
cmd.Env = append(
os.Environ(),
"GIT_CONFIG_NOSYSTEM=1",
"GIT_CONFIG_COUNT=2",
"GIT_CONFIG_KEY_0=safe.bareRepository",
"GIT_CONFIG_VALUE_0=all",
"GIT_CONFIG_KEY_1=init.defaultBranch",
"GIT_CONFIG_VALUE_1=main",
)
cmd.Stderr = os.Stderr
return cmd
}

// createEchoScript returns a path to a platform-appropriate script that
// prints "runner-ok" and exits 0, and a failing script that exits 1.
func createEchoScript(t *testing.T, dir, name string, exitCode int) string {
t.Helper()
var path, content string
if runtime.GOOS == "windows" {
path = filepath.Join(dir, name+".bat")
if exitCode == 0 {
content = "@echo off\r\necho runner-ok\r\n"
} else {
content = "@echo off\r\necho fail-output\r\nexit /b 1\r\n"
}
require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
} else {
path = filepath.Join(dir, name+".sh")
if exitCode == 0 {
content = "#!/bin/sh\necho runner-ok\n"
} else {
content = "#!/bin/sh\necho fail-output\nexit 1\n"
}
require.NoError(t, os.WriteFile(path, []byte(content), 0o755))
}
return path
}

func TestExecRunnerRunsCommandSuccessfully(t *testing.T) {
t.Parallel()

dir := createTestDir(t)
script := createEchoScript(t, dir, "script", 0)

err := ExecRunner{Dir: dir}.Run(context.Background(), script)

require.NoError(t, err)
}

func TestExecRunnerIncludesStdoutAndStderrInErrors(t *testing.T) {
t.Parallel()

dir := createTestDir(t)
script := createEchoScript(t, dir, "fail", 1)

err := ExecRunner{Dir: dir}.Run(context.Background(), script)

require.Error(t, err)
require.True(t, strings.Contains(err.Error(), "fail-output") || strings.Contains(err.Error(), "runner"), fmt.Sprintf("unexpected error: %v", err))
}