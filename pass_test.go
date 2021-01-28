package pass

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

const (
	testGpgID         = "0F5E1E3F3CE3019D9A3AD09313B82ACF5C4BAB55"
	testGpgPassphrase = "test_passphrase"

	tmpDirPrefix = "go-pass-test-"
)

func TestInit(t *testing.T) {
	storeDir, err := ioutil.TempDir("", tmpDirPrefix)
	if err != nil {
		log.Fatalf("create tmp dir: %s", err)
	}

	opts := &Options{
		StoreDir: storeDir,
	}
	err = Init(testGpgID, "", opts)
	Ok(t, err)

	got, err := ioutil.ReadFile(filepath.Join(storeDir, ".gpg-id"))
	if err != nil {
		t.Errorf("unexpected error reading .gpg-id file: %s", err)
		return
	}
	got = bytes.TrimSpace(got)
	if string(got) != testGpgID {
		t.Errorf("wrong .gpg-id: expected: %s, got: %s", testGpgID, string(got))
		return
	}
}

func TestInsert(t *testing.T) {
	storeDir, err := ioutil.TempDir("", tmpDirPrefix)
	if err != nil {
		log.Fatalf("create tmp dir: %s", err)
	}

	opts := &Options{
		StoreDir: storeDir,
	}
	err = Init(testGpgID, "", opts)
	Ok(t, err)

	err = Insert("bar", []byte("mypassword"), false, opts)
	Ok(t, err)

	_, err = os.Stat(filepath.Join(storeDir, "bar.gpg"))
	Ok(t, err)
}

func TestCopy(t *testing.T) {
	storeDir, err := ioutil.TempDir("", tmpDirPrefix)
	if err != nil {
		log.Fatalf("create tmp dir: %s", err)
	}

	opts := &Options{
		StoreDir: storeDir,
	}
	err = Init(testGpgID, "", opts)
	Ok(t, err)

	err = Insert("bar", []byte("mypassword"), false, opts)
	Ok(t, err)

	err = Copy("bar", "baz", false, opts)
	Ok(t, err)

	_, err = os.Stat(filepath.Join(storeDir, "baz.gpg"))
	Ok(t, err)
}

func TestList(t *testing.T) {
	storeDir, err := ioutil.TempDir("", tmpDirPrefix)
	if err != nil {
		log.Fatalf("create tmp dir: %s", err)
	}

	opts := &Options{
		StoreDir: storeDir,
	}
	err = Init(testGpgID, "", opts)
	Ok(t, err)

	err = Insert("google.com/bar", []byte("mypassword"), false, opts)
	Ok(t, err)
	err = Insert("google.com/baz", []byte("mypassword"), false, opts)
	Ok(t, err)
	err = Insert("atlassian.com/baz", []byte("mypassword"), false, opts)
	Ok(t, err)

	ls, err := List("google.com", opts)
	Ok(t, err)
	if len(ls) != 2 {
		t.Errorf("expected 2 items, got %d", len(ls))
		return
	}
	Equal(t, "google.com/bar", ls[0])
	Equal(t, "google.com/baz", ls[1])

	ls, err = List("", opts)
	Ok(t, err)
	if len(ls) != 3 {
		t.Errorf("expected 3 items, got %d", len(ls))
		return
	}
	Equal(t, "atlassian.com/baz", ls[0])
	Equal(t, "google.com/bar", ls[1])
	Equal(t, "google.com/baz", ls[2])
}

func TestShow(t *testing.T) {
	storeDir, err := ioutil.TempDir("", tmpDirPrefix)
	if err != nil {
		log.Fatalf("create tmp dir: %s", err)
	}

	opts := &Options{
		StoreDir: storeDir,
	}
	err = Init(testGpgID, "", opts)
	Ok(t, err)

	err = Insert("google.com/bar", []byte("mypassword"), false, opts)
	Ok(t, err)

	c, err := Show("google.com/bar", testGpgPassphrase, opts)
	Ok(t, err)
	if string(c) != "mypassword" {
		t.Errorf("incorrect content: %s", string(c))
		return
	}
}

func Ok(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected: <nil> error, got: %s", err)
	}
}

func Equal(t *testing.T, expected, got string) {
	t.Helper()
	if expected != got {
		t.Errorf("expected: %s, got: %s", expected, got)
	}
}
