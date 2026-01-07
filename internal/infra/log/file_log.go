package log

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func InitFileLogWriter(file string) (wr io.WriteCloser, err error, wrEnable bool) {
	if strings.ToLower(file) == "none" {
		return nil, nil, false
	}

	out := newStdoutWrap(os.Stdout)

	if strings.ToLower(file) == "stdout" {
		return out, nil, true
	}

	dir := filepath.Dir(file)
	ds, err := os.Lstat(dir)
	if err != nil {
		return out, err, true
	}

	// symlink attack prevent
	if (ds.Mode() & os.ModeSymlink) != 0 {
		return out, err, true
	}

	out, err = os.Open(file)
	if err != nil {
		return out, err, true
	}

	return out, nil, true
}

// stdout proxy without close implementation
type stdoutWrap struct {
	stdout *os.File
}

func newStdoutWrap(stdout *os.File) io.WriteCloser {
	return &stdoutWrap{
		stdout: stdout,
	}
}

func (sw *stdoutWrap) Write(p []byte) (n int, err error) {
	return sw.stdout.Write(p)
}

func (sw *stdoutWrap) Close() error {
	return nil
}
