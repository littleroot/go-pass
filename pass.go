// Package pass provides a Go API for pass, the standard UNIX password manager
// (http://passwordstore.org).
//
// The API of the package aims to be close to each of the subcommands of the
// pass program. See the doc comments on each function for details and
// differences. Some of the subcommands are omitted from the API since I don't
// have a need for them currently.
package pass

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Options struct {
	StoreDir string //  Optional. The value of PASSWORD_STORE_DIR.
}

// Init is equivalent to the "init" subcommand.
func Init(ctx context.Context, gpgID, subfolder string, opts *Options) error {
	var args []string
	if subfolder != "" {
		args = append(args, subfolder)
	}
	args = append(args, gpgID)

	_, err := execCommand(ctx, "init", args, nil, nil, opts)
	if err != nil {
		return fmt.Errorf("exec init: %s", err)
	}
	return nil
}

// List is equivalent to the "ls" subcommand.
// Unlike the original subcommand, this function does not follow and
// list the contents of symbolic links.
func List(ctx context.Context, subfolder string, opts *Options) ([]string, error) {
	storeDir := filepath.Join(os.Getenv("HOME"), ".password-store")
	if opts != nil && opts.StoreDir != "" {
		storeDir = opts.StoreDir
	}

	targetDir := storeDir
	if subfolder != "" {
		targetDir = filepath.Join(storeDir, subfolder)
	}

	var ret []string

	err := filepath.Walk(targetDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".gpg") {
			return nil
		}
		rel, err := filepath.Rel(storeDir, p)
		if err != nil {
			panic(err) // should not happen
		}
		ret = append(ret, strings.TrimSuffix(rel, ".gpg"))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Show is equivalent to the "show" subcommand.
func Show(ctx context.Context, name, gpgPassphrase string, opts *Options) ([]byte, error) {
	env := []string{`PASSWORD_STORE_GPG_OPTS=--passphrase-fd=0 --pinentry-mode=loopback --batch`}
	output, err := execCommand(ctx, "show", []string{name}, strings.NewReader(gpgPassphrase), env, opts)
	if err != nil {
		return nil, fmt.Errorf("exec show: %s: %s", err, output)
	}
	return output, nil
}

// Insert is equivalent to the "insert" subcommand.
func Insert(ctx context.Context, name string, content []byte, force bool, opts *Options) error {
	var args []string
	if force {
		args = append(args, "--force")
	}
	args = append(args, "--multiline") // always use so we can set stdin
	args = append(args, name)

	_, err := execCommand(ctx, "insert", args, bytes.NewReader(content), nil, opts)
	if err != nil {
		return fmt.Errorf("exec insert: %s", err)
	}
	return nil
}

// Remove is equivalent to the "rm" subcommand.
func Remove(ctx context.Context, name string, recursive, force bool, opts *Options) error {
	var args []string
	if recursive {
		args = append(args, "--recursive")
	}
	if force {
		args = append(args, "--force")
	}
	args = append(args, name)

	_, err := execCommand(ctx, "rm", args, nil, nil, opts)
	if err != nil {
		return fmt.Errorf("exec rm: %s", err)
	}
	return nil
}

// Move is equivalent to the "mv" subcommand.
func Move(ctx context.Context, oldPath, newPath string, force bool, opts *Options) error {
	var args []string
	if force {
		args = append(args, "--force")
	}
	args = append(args, oldPath)
	args = append(args, newPath)

	_, err := execCommand(ctx, "mv", args, nil, nil, opts)
	if err != nil {
		return fmt.Errorf("exec mv: %s", err)
	}
	return nil
}

// Copy is equivalent to the "cp" subcommand.
func Copy(ctx context.Context, oldPath, newPath string, force bool, opts *Options) error {
	var args []string
	if force {
		args = append(args, "--force")
	}
	args = append(args, oldPath)
	args = append(args, newPath)

	_, err := execCommand(ctx, "cp", args, nil, nil, opts)
	if err != nil {
		return fmt.Errorf("exec cp: %s", err)
	}
	return nil
}

// Git is equivalent to the "git" subcommand.
func Git(ctx context.Context, gitArgs []string, opts *Options) error {
	_, err := execCommand(ctx, "git", gitArgs, nil, nil, opts)
	if err != nil {
		return fmt.Errorf("exec git: %s", err)
	}
	return nil
}

func execCommand(ctx context.Context, subcommand string, args []string, stdin io.Reader, extraEnv []string, opts *Options) (stdout []byte, err error) {
	allArgs := []string{subcommand}
	allArgs = append(allArgs, args...)

	var env []string
	if opts != nil && opts.StoreDir != "" {
		env = append(env, fmt.Sprintf("PASSWORD_STORE_DIR=%s", opts.StoreDir))
	}
	env = append(env, extraEnv...)

	cmd := exec.CommandContext(ctx, "pass", allArgs...)
	cmd.Env = env
	if stdin != nil {
		cmd.Stdin = stdin
	}

	return cmd.CombinedOutput()
}
