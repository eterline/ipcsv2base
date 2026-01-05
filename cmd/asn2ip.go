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
	rootCmd.AddCommand(asn2ipCmd)
	asn2ipCmd.PersistentFlags().StringP("input", "i", "ip-to-country.csv", "CSV file")
	asn2ipCmd.PersistentFlags().StringP("output", "o", "asn2ip.bin", "Output directory")
	asn2ipCmd.PersistentFlags().StringP("read", "r", "country2ip.bin", "Testing reading base")
}

var asn2ipCmd = &cobra.Command{
	Use:   "asn2ip",
	Short: "asn2ip",
	Run: func(cmd *cobra.Command, args []string) {

		read, _ := cmd.Flags().GetString("read")
		if read != "" {
			return
		}

		input, _ := cmd.Flags().GetString("input")
		output, _ := cmd.Flags().GetString("output")

		log.Printf("[INIT] Input file: %s", input)
		log.Printf("[INIT] Output base: %s", output)

		src, err := openInputCSV(input)
		if err != nil {
			log.Fatal(err)
		}
		defer src.Close()

		base, err := createAsnBase(output)
		if err != nil {
			log.Fatal(err)
		}
		defer base.Close()

		if err := processCSVToAsnBase(src, base); err != nil {
			log.Fatal(err)
		}

		log.Printf(
			"[DONE] Writing finished successfully, records written: %d",
			base.Writes(),
		)
	},
}

func createAsnBase(path string) (*ipcsv2base.NetworkCompanyWriter, error) {
	base, err := ipcsv2base.CreateNetworkCompanyBase(path)
	if err != nil {
		return nil, fmt.Errorf(
			"[ERROR] Failed to create asn base %q: %w",
			path, err,
		)
	}
	log.Println("[OK] ASN base created")
	return base, nil
}

func processCSVToAsnBase(src *os.File, base *ipcsv2base.NetworkCompanyWriter) error {

	rd := csv.NewReader(src)
	rd.FieldsPerRecord = 6 // network, asn, country_code, name, org, domain

	log.Println("[INFO] CSV reader initialized")
	log.Println("[INFO] Skipping CSV header")

	// skip header
	if _, err := rd.Read(); err != nil {
		return fmt.Errorf(
			"[ERROR] Failed to read CSV header: %w",
			err,
		)
	}

	log.Println("[START] Writing asn records to base")

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
		name := rec[3]
		org := rec[4]

		pfx, err := netip.ParsePrefix(network)
		if err != nil {
			return fmt.Errorf(
				"[ERROR] Invalid network prefix %q at record %d: %w",
				network, processed, err,
			)
		}

		entry := ipcsv2base.NewNetworkCompany(pfx, name, org)
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
