package baseapi

import (
	"net/netip"
	"time"

	"github.com/eterline/ipcsv2base/internal/model"
)

// IPMetadataDTO - Flat DTO for API responses.
type IPMetadataDTO struct {
	LookupDurationMs int64  `json:"lookup_duration_ms"`
	Success          bool   `json:"success"`
	RequestIP        string `json:"request_ip"`
	NetworkType      string `json:"network_type"`
	Network          string `json:"network,omitempty"`
	ContinentCode    string `json:"continent_code,omitempty"`
	CountryCode      string `json:"country_code,omitempty"`
	CountryName      string `json:"country_name,omitempty"`
	ASN              int32  `json:"asn,omitempty"`
	ASNName          string `json:"asn_name,omitempty"`
	ASNOrg           string `json:"asn_org,omitempty"`
	ASNCountryCode   string `json:"asn_country_code,omitempty"`
	Domain           string `json:"domain,omitempty"`
}

func domain2IPMetadataDTO(m *model.IPMetadata, dur time.Duration, reqip netip.Addr) *IPMetadataDTO {
	return &IPMetadataDTO{
		Success:          true,
		RequestIP:        reqip.String(),
		LookupDurationMs: dur.Milliseconds(),
		NetworkType:      m.Type.String(),
		Network:          m.Network.String(),
		ContinentCode:    m.Geo.ContinentCode.String(),
		CountryCode:      m.Geo.CountryCode.String(),
		CountryName:      m.Geo.CountryName,
		ASN:              m.ASN.ASN,
		ASNName:          m.ASN.Name,
		ASNOrg:           m.ASN.Org,
		ASNCountryCode:   m.ASN.CountryCode.String(),
		Domain:           m.ASN.Domain,
	}
}
