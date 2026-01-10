package config

import "github.com/go-playground/validator/v10"

func baseStructValidation(sl validator.StructLevel) {
	b := sl.Current().Interface().(Base)

	hasTSV := len(b.CountryTSV) > 0
	hasCSVPair := b.CountryCSV != "" && b.AsnCSV != ""

	switch {
	case hasTSV && hasCSVPair:
		sl.ReportError(
			b.CountryTSV,
			"CountryTSV",
			"country-tsvs",
			"mutuallyexclusive",
			"",
		)

	case !hasTSV && !hasCSVPair:
		sl.ReportError(
			b.CountryTSV,
			"CountryTSV",
			"country-tsvs",
			"required",
			"country-tsvs or (country-csv + asn-csv)",
		)
	}
}
