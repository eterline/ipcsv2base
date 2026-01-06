package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/netip"
	"os"

	"github.com/eterline/ipcsv2base"
	"github.com/spf13/cobra"
)

type IPVersion string

const (
	IPv4 IPVersion = "ipv4"
	IPv6 IPVersion = "ipv6"
	All  IPVersion = "all"
)

func (v *IPVersion) Set(s string) error {
	switch s {
	case "ipv4", "IPv4":
		*v = IPv4
	case "ipv6", "IPv6":
		*v = IPv6
	case "all", "ALL":
		*v = All
	default:
		return fmt.Errorf("invalid version %q, must be one of: ipv4, ipv6, all", s)
	}
	return nil
}

func (v *IPVersion) Type() string {
	return "ipversion"
}

func (v *IPVersion) String() string {
	return string(*v)
}

func init() {
	rootCmd.AddCommand(country2ipCmd)
	country2ipCmd.PersistentFlags().StringP("input", "i", "ip-to-country.csv", "CSV file")
	country2ipCmd.PersistentFlags().StringP("output", "o", "country2ip.bin", "Output directory")
	country2ipCmd.PersistentFlags().BoolP("read", "r", false, "Testing reading base")
	country2ipCmd.PersistentFlags().StringSliceP("codes", "c", []string{}, "Testing reading codes to read")
	country2ipCmd.PersistentFlags().StringP("version", "v", "all", "Testing reading IP version")
}

var country2ipCmd = &cobra.Command{
	Use:   "country2ip",
	Short: "country2ip",
	Run: func(cmd *cobra.Command, args []string) {

		input, _ := cmd.Flags().GetString("input")
		output, _ := cmd.Flags().GetString("output")

		read, _ := cmd.Flags().GetBool("read")
		if read {
			ipVer, _ := cmd.Flags().GetString("version")
			codes, _ := cmd.Flags().GetStringSlice("codes")

			if err := readCountryBase(input, ipVer, codes); err != nil {
				log.Fatal(err)
			}
			return
		}

		log.Printf("[INIT] Input file: %s", input)
		log.Printf("[INIT] Output base: %s", output)

		src, err := openInputCSV(input)
		if err != nil {
			log.Fatal(err)
		}
		defer src.Close()

		base, err := createCountryBase(output)
		if err != nil {
			log.Fatal(err)
		}
		defer base.Close()

		if err := processCSVToCountryBase(src, base); err != nil {
			log.Fatal(err)
		}

		log.Printf(
			"[DONE] Writing finished successfully, records written: %d",
			base.Writes(),
		)
	},
}

func readCountryBase(path, ipver string, codes []string) error {
	log.Printf("[INIT] Opening country base: %s", path)

	base, err := ipcsv2base.OpenNetworkCountryBase(path, parseIpVer(ipver), codes...)
	if err != nil {
		return fmt.Errorf(
			"[ERROR] Failed to open country base %q: %w",
			path, err,
		)
	}
	defer base.Close()

	log.Println("[START] Reading country base")

	var i int = 0

	for {
		rec, err := base.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf(
				"[ERROR] Read error at record %d: %w",
				i, err,
			)
		}

		i++
		log.Printf(
			"[REC %d] %s - %s",
			i,
			rec.Country(),
			rec.Prefix(),
		)
	}

	log.Printf("[DONE] Reading finished, records read: %d", i)
	return nil
}

func openInputCSV(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf(
			"[ERROR] Failed to open source CSV file %q: %w",
			path, err,
		)
	}
	log.Println("[OK] Source CSV file opened")
	return f, nil
}

func createCountryBase(path string) (*ipcsv2base.NetworkCountryWriter, error) {
	base, err := ipcsv2base.CreateNetworkCountryBase(path)
	if err != nil {
		return nil, fmt.Errorf(
			"[ERROR] Failed to create country base %q: %w",
			path, err,
		)
	}
	log.Println("[OK] Country base created")
	return base, nil
}

func processCSVToCountryBase(src *os.File, base *ipcsv2base.NetworkCountryWriter) error {

	rd := csv.NewReader(src)
	rd.FieldsPerRecord = 4 // network, continent, country_code, country_name

	log.Println("[INFO] CSV reader initialized")
	log.Println("[INFO] Skipping CSV header")

	if _, err := rd.Read(); err != nil {
		return fmt.Errorf(
			"[ERROR] Failed to read CSV header: %w",
			err,
		)
	}

	log.Println("[START] Writing country records to base")

	var processed int

	for {
		rec, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf(
				"[ERROR] CSV read error at record %d: %w",
				processed, err,
			)
		}

		network := rec[0]
		code := rec[2]

		pfx, err := netip.ParsePrefix(network)
		if err != nil {
			return fmt.Errorf(
				"[ERROR] Invalid network prefix %q at record %d: %w",
				network, processed, err,
			)
		}

		entry := ipcsv2base.NewNetworkCountry(pfx, code)
		if _, err := base.Write(entry); err != nil {
			return fmt.Errorf(
				"[ERROR] Failed to write record %d: %w",
				processed, err,
			)
		}

		processed++
		if processed%256_000 == 0 {
			log.Printf("[PROGRESS] %d records written", processed)
		}
	}

	log.Printf("[END] CSV processing finished, total records: %d, base size: %dB", processed, base.Size())
	return nil
}

func parseIpVer(ipver string) ipcsv2base.NetworkVersion {
	if ipver == "v4" {
		return ipcsv2base.IPv4
	}
	if ipver == "v6" {
		return ipcsv2base.IPv6
	}
	return ipcsv2base.IPv4IPv6
}
