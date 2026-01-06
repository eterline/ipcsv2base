package ipcsv2base_test

import (
	"io"
	"math/rand"
	"net/netip"
	"path/filepath"
	"testing"

	"github.com/eterline/ipcsv2base"
)

var (
	asn2ipSrc     = filepath.Join("asn2ip.bin")
	country2ipSrc = filepath.Join("country2ip.bin")
)

func TestCreateNetworkCountryBase(t *testing.T) {
	base, err := ipcsv2base.CreateNetworkCountryBase(country2ipSrc)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()

	for i := 0; i < 256; i++ {
		ip := netip.AddrFrom4([4]byte{
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
			0,
			0,
		})
		record := ipcsv2base.NewNetworkCountry(netip.PrefixFrom(ip, 16), "RU")
		_, err := base.Write(record)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func BenchmarkCreateNetworkCountryBase(b *testing.B) {
	base, err := ipcsv2base.CreateNetworkCountryBase(country2ipSrc)
	if err != nil {
		b.Fatal(err)
	}
	defer base.Close()

	const recordsCount = 1024
	records := make([]ipcsv2base.NetworkCountry, recordsCount)

	r := rand.New(rand.NewSource(1))

	for i := 0; i < recordsCount; i++ {
		ip := netip.AddrFrom4([4]byte{
			byte(r.Intn(256)),
			byte(r.Intn(256)),
			0,
			0,
		})

		pfx := netip.PrefixFrom(ip, 16)

		records[i] = ipcsv2base.NewNetworkCountry(
			pfx,
			"RU",
		)
	}

	mask := recordsCount - 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := records[i&mask]

		n, err := base.Write(rec)
		if err != nil {
			b.Fatal(err)
		}

		b.SetBytes(int64(n))
	}
}

func BenchmarkCreateNetworkCountryBaseAllocable(b *testing.B) {
	base, err := ipcsv2base.CreateNetworkCountryBase(country2ipSrc)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record := ipcsv2base.NewNetworkCountry(netip.MustParsePrefix("10.192.0.0/24"), "RU")
		n, err := base.Write(record)
		if err != nil {
			b.Fatal(err)
		}

		b.SetBytes(int64(n))
	}

	b.StopTimer()
	base.Close()
	b.StartTimer()
}

func BenchmarkNetworkCountryBaseNext(b *testing.B) {
	base, err := ipcsv2base.OpenNetworkCountryBase(country2ipSrc, ipcsv2base.IPv4IPv6)
	if err != nil {
		b.Fatal(err)
	}
	defer base.Close()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := base.Next()
		if err == io.EOF {
			if err := base.Reset(); err != nil {
				b.Fatal(err)
			}
			continue
		}
		if err != nil {
			b.Fatal(err)
		}
	}
}
