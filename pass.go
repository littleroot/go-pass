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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Options struct {
	StoreDir string
}

// Init is equivalent to the "init" subcommand.
func Init(gpgID, subfolder string, opts *Options) error {
	var args []string
	if subfolder != "" {
		args = append(args, subfolder)
	}
	args = append(args, gpgID)

	_, err := execCommand("init", args, nil, opts)
	if err != nil {
		return fmt.Errorf("exec init: %s", err)
	}
	return nil
}

// Show is equivalent to  the "ls" subcommand.
// Unlike the original subcommand, this function does not follow and
// list the contents of symbolic links.
func List(subfolder string, opts *Options) ([]string, error) {
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
		rel, _ := filepath.Rel(storeDir, p)
		ret = append(ret, strings.TrimSuffix(rel, ".gpg"))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Show is equivalent to the "show" subcommand.
func Show(name string, opts *Options) ([]byte, error) {
	content, err := execCommand("show", []string{name}, nil, opts)
	if err != nil {
		return nil, fmt.Errorf("exec show: %s", err)
	}
	return ioutil.ReadAll(content)
}

// Insert is equivalent to the "insert" subcommand.
func Insert(name string, content []byte, force bool, opts *Options) error {
	var args []string
	if force {
		args = append(args, "--force")
	}
	args = append(args, "--multiline") // always use so we can set stdin
	args = append(args, name)

	_, err := execCommand("insert", args, bytes.NewReader(content), opts)
	if err != nil {
		return fmt.Errorf("exec insert: %s", err)
	}
	return nil
}

// Remove is equivalent to the "rm" subcommand.
func Remove(name string, recursive, force bool, opts *Options) error {
	var args []string
	if recursive {
		args = append(args, "--recursive")
	}
	if force {
		args = append(args, "--force")
	}
	args = append(args, name)

	_, err := execCommand("rm", args, nil, opts)
	if err != nil {
		return fmt.Errorf("exec rm: %s", err)
	}
	return nil
}

// Move is equivalent to the "mv" subcommand.
func Move(oldPath, newPath string, force bool, opts *Options) error {
	var args []string
	if force {
		args = append(args, "--force")
	}
	args = append(args, oldPath)
	args = append(args, newPath)

	_, err := execCommand("mv", args, nil, opts)
	if err != nil {
		return fmt.Errorf("exec mv: %s", err)
	}
	return nil
}

// Copy is equivalent to the "cp" subcommand.
func Copy(oldPath, newPath string, force bool, opts *Options) error {
	var args []string
	if force {
		args = append(args, "--force")
	}
	args = append(args, oldPath)
	args = append(args, newPath)

	_, err := execCommand("cp", args, nil, opts)
	if err != nil {
		return fmt.Errorf("exec cp: %s", err)
	}
	return nil
}

// Git is equivalent to the "git" subcommand.
func Git(gitArgs []string, opts *Options) error {
	_, err := execCommand("git", gitArgs, nil, opts)
	if err != nil {
		return fmt.Errorf("exec git subcommand: %s", err)
	}
	return nil
}

func execCommand(subcommand string, args []string, in io.Reader, opts *Options) (out io.Reader, err error) {
	allArgs := []string{subcommand}
	allArgs = append(allArgs, args...)

	var env []string
	if opts != nil && opts.StoreDir != "" {
		env = append(env, fmt.Sprintf("PASSWORD_STORE_DIR=%s", opts.StoreDir))
	}

	var buf bytes.Buffer

	cmd := exec.Command("pass", allArgs...)
	cmd.Env = env
	cmd.Stdout = &buf
	if in != nil {
		cmd.Stdin = in
	}

	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return &buf, nil
}
