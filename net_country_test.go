package ipcsv2base_test

import (
	"path/filepath"
	"testing"

	"github.com/eterline/ipcsv2base"
)

var (
	asn2ipSrc     = filepath.Join("asn2ip.bin")
	country2ipSrc = filepath.Join("country2ip.bin")
)

func BenchmarkNetworkCountryReader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		base, err := ipcsv2base.OpenNetworkCountryBase(country2ipSrc, ipcsv2base.IPv4IPv6)
		if err != nil {
			b.Fatal(err)
		}

		b.StartTimer()
		if _, err := base.Container(); err != nil {
			b.Fatal(err)
		}

		b.StopTimer()
		base.Close()
		b.StartTimer()
	}
}
