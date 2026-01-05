package ipcsv2base

import (
	"compress/flate"
	"fmt"
	"os"
)

type NetworkCompanyWriter struct {
	f      *os.File
	w      *flate.Writer
	buf    [companyRecSize]byte
	writes int64
}

func CreateNetworkCompanyBase(file string) (*NetworkCompanyWriter, error) {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create company base file: %w", err)
	}

	w, err := flate.NewWriter(f, flate.BestSpeed)
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to create flate writer: %w", err)
	}

	return &NetworkCompanyWriter{f: f, w: w}, nil
}

func (w *NetworkCompanyWriter) Write(n NetworkCompany) error {
	pfxVec, err := PrefixToVec24(n.Network())
	if err != nil {
		return err
	}

	// network (24)
	copy(w.buf[0:pfxSize], pfxVec[:])

	// name (256)
	name := n.Name()
	if len(name) > companyLineStr {
		name = name[:companyLineStr]
	}
	copy(w.buf[pfxSize:pfxSize+companyLineStr], name)

	// org (256)
	org := n.Org()
	if len(org) > companyLineStr {
		org = org[:companyLineStr]
	}
	copy(w.buf[pfxSize+companyLineStr:pfxSize+companyLineStr*2], org)

	if _, err = w.w.Write(w.buf[:]); err == nil {
		w.writes++
	}

	return err
}

func (w *NetworkCompanyWriter) Close() error {
	if w.w != nil {
		if err := w.w.Close(); err != nil {
			_ = w.f.Close()
			return err
		}
	}
	if w.f != nil {
		return w.f.Close()
	}
	return nil
}

func (w *NetworkCompanyWriter) Writes() int64 {
	return w.writes
}

func (w *NetworkCompanyWriter) Size() int64 {
	stat, err := w.f.Stat()
	if err != nil {
		return 0
	}
	return stat.Size()
}

// ===================
