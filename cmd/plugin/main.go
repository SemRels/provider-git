// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The provider-git Authors

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	semrelplugin "github.com/SemRels/provider-git/internal/plugin"
)

type gitClient interface {
	PushTag(ctx context.Context, tagName, remote string) error
	PushBranch(ctx context.Context, branch, remote string) error
}

func main() {
	if err := run(context.Background(), os.Getenv, os.Stdout, semrelplugin.NewClient(semrelplugin.ConfigFromEnv(os.Getenv), semrelplugin.ExecRunner{})); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, getenv func(string) string, stdout io.Writer, client gitClient) error {
	if getenv == nil {
		getenv = os.Getenv
	}

	config := semrelplugin.ConfigFromEnv(getenv)
	tagName := strings.TrimSpace(getenv("SEMREL_TAG_NAME"))
	if tagName == "" {
		return errors.New("SEMREL_TAG_NAME is required")
	}

	branch := strings.TrimSpace(getenv("SEMREL_BRANCH"))
	if branch == "" {
		return errors.New("SEMREL_BRANCH is required")
	}

	if strings.EqualFold(strings.TrimSpace(getenv("SEMREL_DRY_RUN")), "true") {
		_, err := fmt.Fprintf(stdout, "dry run: would push tag %s and branch %s to %s\n", tagName, branch, config.Remote)
		return err
	}

	if _, err := fmt.Fprintf(stdout, "pushing tag %s to %s\n", tagName, config.Remote); err != nil {
		return err
	}

	if err := client.PushTag(ctx, tagName, config.Remote); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "pushing branch %s to %s\n", branch, config.Remote); err != nil {
		return err
	}

	return client.PushBranch(ctx, branch, config.Remote)
}
