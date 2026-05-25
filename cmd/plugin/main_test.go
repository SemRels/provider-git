// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The provider-git Authors

package main

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunPushesTagAndBranch(t *testing.T) {
	t.Parallel()

	client := &fakeClient{}
	stdout := &bytes.Buffer{}

	err := run(context.Background(), mapEnv(map[string]string{
		"SEMREL_TAG_NAME":      "v1.0.0",
		"SEMREL_BRANCH":        "main",
		"SEMREL_PLUGIN_REMOTE": "upstream",
	}), stdout, client)

	require.NoError(t, err)
	require.Equal(t, []call{{kind: "tag", name: "v1.0.0", remote: "upstream"}, {kind: "branch", name: "main", remote: "upstream"}}, client.calls)
	require.Contains(t, stdout.String(), "pushing tag v1.0.0 to upstream")
	require.Contains(t, stdout.String(), "pushing branch main to upstream")
}

func TestRunSkipsPushesInDryRunMode(t *testing.T) {
	t.Parallel()

	client := &fakeClient{}
	stdout := &bytes.Buffer{}

	err := run(context.Background(), mapEnv(map[string]string{
		"SEMREL_TAG_NAME": "v1.0.0",
		"SEMREL_BRANCH":   "release",
		"SEMREL_DRY_RUN":  "true",
	}), stdout, client)

	require.NoError(t, err)
	require.Empty(t, client.calls)
	require.Contains(t, stdout.String(), "dry run: would push tag v1.0.0 and branch release to origin")
}

func TestRunRequiresTagName(t *testing.T) {
	t.Parallel()

	err := run(context.Background(), mapEnv(map[string]string{"SEMREL_BRANCH": "main"}), &bytes.Buffer{}, &fakeClient{})

	require.EqualError(t, err, "SEMREL_TAG_NAME is required")
}

func TestRunRequiresBranch(t *testing.T) {
	t.Parallel()

	err := run(context.Background(), mapEnv(map[string]string{"SEMREL_TAG_NAME": "v1.0.0"}), &bytes.Buffer{}, &fakeClient{})

	require.EqualError(t, err, "SEMREL_BRANCH is required")
}

func TestRunReturnsTagPushError(t *testing.T) {
	t.Parallel()

	client := &fakeClient{tagErr: errors.New("tag push failed")}

	err := run(context.Background(), mapEnv(map[string]string{
		"SEMREL_TAG_NAME": "v1.0.0",
		"SEMREL_BRANCH":   "main",
	}), &bytes.Buffer{}, client)

	require.EqualError(t, err, "tag push failed")
}

func TestRunReturnsBranchPushError(t *testing.T) {
	t.Parallel()

	client := &fakeClient{branchErr: errors.New("branch push failed")}

	err := run(context.Background(), mapEnv(map[string]string{
		"SEMREL_TAG_NAME": "v1.0.0",
		"SEMREL_BRANCH":   "main",
	}), &bytes.Buffer{}, client)

	require.EqualError(t, err, "branch push failed")
}

type call struct {
	kind   string
	name   string
	remote string
}

type fakeClient struct {
	calls     []call
	tagErr    error
	branchErr error
}

func (f *fakeClient) PushTag(_ context.Context, tagName, remote string) error {
	f.calls = append(f.calls, call{kind: "tag", name: tagName, remote: remote})
	return f.tagErr
}

func (f *fakeClient) PushBranch(_ context.Context, branch, remote string) error {
	f.calls = append(f.calls, call{kind: "branch", name: branch, remote: remote})
	return f.branchErr
}

func mapEnv(values map[string]string) func(string) string {
	return func(key string) string {
		return values[key]
	}
}
