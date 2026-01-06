package ipcsv2base

import (
	"bytes"
	"compress/flate"
	"io"
	"os"
)

type BaseReader struct {
	file  *os.File
	flate io.ReadCloser
	buf   []byte
	recs  int64
}

type selfSizeAllocer interface {
	selfSize() int
}

func OpenBase(path string, ssa selfSizeAllocer) (*BaseReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fr := flate.NewReader(f)

	return &BaseReader{
		file:  f,
		flate: fr,
		buf:   make([]byte, ssa.selfSize()),
	}, nil
}

type BaseReaderRecord interface {
	readFields(r io.Reader) error
	setPrefix(prefixVector)
}

func (br *BaseReader) Next(rec BaseReaderRecord) error {
	var pfx prefixVector

	_, err := io.ReadFull(br.flate, br.buf)
	if err != nil {
		return err
	}

	copy(pfx[:], br.buf)
	rec.setPrefix(pfx)

	if err := rec.readFields(bytes.NewReader(br.buf[24:])); err != nil {
		return err
	}

	br.recs++
	return nil
}

func (br *BaseReader) Reset() error {
	if br.flate != nil {
		_ = br.flate.Close()
	}
	if _, err := br.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	br.flate = flate.NewReader(br.file)
	return nil
}

func (br *BaseReader) Close() error {
	if br.flate != nil {
		_ = br.flate.Close()
	}
	return br.file.Close()
}
