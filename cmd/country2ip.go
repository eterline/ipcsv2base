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

func init() {
	rootCmd.AddCommand(country2ipCmd)
	country2ipCmd.PersistentFlags().StringP("input", "i", "ip-to-country.csv", "CSV file")
	country2ipCmd.PersistentFlags().StringP("output", "o", "country2ip.bin", "Output directory")
	country2ipCmd.PersistentFlags().BoolP("read", "r", false, "Testing reading base")
	country2ipCmd.PersistentFlags().StringSliceP("codes", "c", []string{}, "Testing reading codes to read")
}

var country2ipCmd = &cobra.Command{
	Use:   "country2ip",
	Short: "country2ip",
	Run: func(cmd *cobra.Command, args []string) {

		input, _ := cmd.Flags().GetString("input")
		output, _ := cmd.Flags().GetString("output")

		read, _ := cmd.Flags().GetBool("read")
		if read {
			codes, _ := cmd.Flags().GetStringSlice("codes")
			if err := readCountryBase(input, codes); err != nil {
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

func readCountryBase(path string, codes []string) error {
	log.Printf("[INIT] Opening country base: %s", path)

	base, err := ipcsv2base.OpenNetworkCountryBase(path, ipcsv2base.IPv6, codes...)
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
			rec.Code(),
			rec.Network(),
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

	// skip header
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
		if err := base.Write(entry); err != nil {
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
