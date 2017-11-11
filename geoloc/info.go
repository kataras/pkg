package geoloc

// Info contains the total information about a remote machine with a specific IP.
type Info struct {
	IP           string `sql:"ip" json:"ip,omitempty"`
	GeoInfo      `json:"geo"`
	LanguageInfo `json:"language"`
}

// GeoInfo contains the geo location information based on the IP of a remote machine.
type GeoInfo struct {
	Name        string  `sql:"name" json:"name,omitempty"`
	Country     string  `sql:"country" json:"country"`
	CountryCode string  `sql:"country_code" json:"countryCode"`
	Region      string  `sql:"region" json:"region"`
	RegionName  string  `sql:"region_name" json:"regionName"`
	City        string  `sql:"city" json:"city"`
	Zip         string  `sql:"zip" json:"zip"`
	Lat         float64 `sql:"lat" json:"lat"`
	Lon         float64 `sql:"lon" json:"lon"`
	Timezone    string  `sql:"timezone" json:"timezone"`
	ISP         string  `sql:"isp" json:"isp"`
	Org         string  `sql:"org" json:"org"`
	As          string  `sql:"as" json:"as"`
}

// LanguageInfo contains the Languge and the LanguageCode based on the IP of a remote machine.
type LanguageInfo struct {
	Language     string `sql:"language" json:"language"`
	LanguageCode string `sql:"language_code" json:"languageCode"`
}
