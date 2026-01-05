package ipcsv2base

import (
	"compress/flate"
	"fmt"
	"io"
	"net/netip"
	"os"

	"go4.org/netipx"
)

type NetworkCountryWriter struct {
	f      *os.File
	w      *flate.Writer
	buf    [countryRecSize]byte
	writes int64
}

func CreateNetworkCountryBase(file string) (*NetworkCountryWriter, error) {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create country base file: %w", err)
	}

	w, err := flate.NewWriter(f, flate.BestSpeed)
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to create flate writer: %w", err)
	}

	return &NetworkCountryWriter{f: f, w: w}, nil
}

func (w *NetworkCountryWriter) Write(n NetworkCountry) error {
	pfxVec, err := PrefixToVec24(n.Network())
	if err != nil {
		return err
	}

	w.buf = [countryRecSize]byte{}    // clear buffer
	copy(w.buf[0:pfxSize], pfxVec[:]) // write network vector

	code := n.Code()
	if len(code) != codeSize {
		return fmt.Errorf("invalid country code")
	}

	// write code vector in uppercase
	w.buf[24] = code[0]
	w.buf[25] = code[1]

	if _, err = w.w.Write(w.buf[:]); err == nil {
		w.writes++
	}

	return err
}

func (w *NetworkCountryWriter) Save() error {
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

func (w *NetworkCountryWriter) Close() error {
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

func (w *NetworkCountryWriter) Writes() int64 {
	return w.writes
}

func (w *NetworkCountryWriter) Size() int64 {
	stat, err := w.f.Stat()
	if err != nil {
		return 0
	}
	return stat.Size()
}

// ===================

type NetworkCountryReader struct {
	f       *os.File
	r       io.ReadCloser
	buf     [26]byte
	codes   map[uint16]struct{}
	testVer func(netip.Addr) bool
}

func OpenNetworkCountryBase(path string, ver NetworkVersion, codes ...string) (*NetworkCountryReader, error) {

	var codesMap map[uint16]struct{}

	if len(codes) > 0 {
		codesMap = make(map[uint16]struct{})
		for _, code := range codes {
			if len(code) != codeSize {
				return nil, fmt.Errorf("invalid country code: %s", code)
			}
			code := codeToUint16LE([]byte(code)[0], []byte(code)[1])
			codesMap[code] = struct{}{}
		}
	} else {
		codesMap = nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r := flate.NewReader(f)

	var testVer func(netip.Addr) bool
	switch ver {
	case IPv4:
		testVer = func(a netip.Addr) bool { return a.Is4In6() || a.Is4() }
	case IPv6:
		testVer = func(a netip.Addr) bool { return a.Is6() }
	default:
		testVer = func(a netip.Addr) bool { return true }
	}

	return &NetworkCountryReader{f: f, r: r, codes: codesMap, testVer: testVer}, nil
}

func (r *NetworkCountryReader) Next() (NetworkCountry, error) {
	for {
		_, err := io.ReadFull(r.r, r.buf[:])
		if err != nil {
			return NetworkCountry{}, err
		}

		if r.codes != nil {
			code := codeToUint16LE(r.buf[24], r.buf[25])
			if _, ok := r.codes[code]; !ok {
				continue
			}
		}

		code := string(r.buf[24:26])

		var tmp [24]byte
		copy(tmp[:], r.buf[:24])

		pfx, err := Vec24ToPrefix(tmp)
		if err != nil {
			return NetworkCountry{}, err
		}

		if !r.testVer(pfx.Addr()) {
			continue
		}

		return NewNetworkCountry(pfx, code), nil
	}
}

func (r *NetworkCountryReader) All() ([]NetworkCountry, error) {
	var nets []NetworkCountry

	for {
		rec, err := r.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		nets = append(nets, rec)
	}

	return nets, nil
}

func (r *NetworkCountryReader) Close() error {
	if r.r != nil {
		if err := r.r.Close(); err != nil {
			_ = r.f.Close()
			return err
		}
	}
	if r.f != nil {
		return r.f.Close()
	}
	return nil
}

func (r *NetworkCountryReader) Container() (Container, error) {
	setBuild := &netipx.IPSetBuilder{}

	for {
		rec, err := r.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		setBuild.AddPrefix(rec.Network())
	}

	return setBuild.IPSet()
}
