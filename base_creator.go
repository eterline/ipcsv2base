package ipcsv2base

import (
	"bytes"
	"compress/flate"
	"io"
	"os"
)

type BaseCreator struct {
	buf   *bytes.Buffer
	file  *os.File
	flate *flate.Writer
	recs  int64
}

func CreateBase(path string) (*BaseCreator, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	flate, err := flate.NewWriter(f, flate.BestSpeed)
	if err != nil {
		return nil, err
	}

	return &BaseCreator{
		file:  f,
		flate: flate,
		buf:   &bytes.Buffer{},
	}, nil
}

type BaseRecorder interface {
	prefixVec() prefixVector
	writeFields(io.Writer)
}

func (bc *BaseCreator) Add(r BaseRecorder) (n int, err error) {
	bc.buf.Reset()

	pfx := r.prefixVec()
	bc.buf.Write(pfx[:])
	r.writeFields(bc.buf)

	n, err = bc.flate.Write(bc.buf.Bytes())
	if err != nil {
		return 0, err
	}

	bc.recs++
	return n, nil
}

func (bc *BaseCreator) Close() error {
	if bc.flate != nil {
		_ = bc.flate.Close()
		return bc.file.Close()
	}
	return bc.file.Close()
}

func (bc *BaseCreator) Writes() int64 {
	return bc.recs
}

func (bc *BaseCreator) Size() int64 {
	stat, err := bc.file.Stat()
	if err != nil {
		return 0
	}
	return stat.Size()
}
