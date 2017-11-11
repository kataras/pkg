package geoloc

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/kataras/chronos/ext/http"
	"github.com/kataras/golog"
)

// Fetch returns the total geo location information.
func Fetch(ip string) (Info, bool) {
	info := Info{IP: ip}

	err := FetchGeo(ip, &info.GeoInfo)
	if err != nil {
		golog.Errorf("error while trying to fetch the geo location information from '%s': %v", ip, err)
		return info, false
	}

	info.Name, err = FetchNameFromIP(ip)
	if err != nil {
		golog.Errorf("error while trying to fetch the server name from '%s': %v", ip, err)
		return info, false
	}

	err = FetchLanguage(info.CountryCode, &info.LanguageInfo)
	if err != nil {
		golog.Errorf("error while trying to fetch the language information from country code: '%s' from '%s':%v", info.CountryCode, ip, err)
		return info, false
	}

	return info, true
}

// limits: 150 requests per minute.
var geohttp = http.New(150, 1*time.Minute)

func geoServiceFromIP(ip string) string {
	return fmt.Sprintf("http://ip-api.com/json/%s", ip)
}

// FetchGeo accepts a pointer to a Server table
// and sets its values based on a geo location service.
// This checks the limit (150 per minute).
// http://ip-api.com/json/$IP
func FetchGeo(ip string, geoInfo *GeoInfo) error {
	return geohttp.ReadJSON(geoServiceFromIP(ip), geoInfo)
}

// FetchNameFromIP gets the lookup name based on the ip
// and returns it.
func FetchNameFromIP(ip string) (string, error) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("nslookup", ip, "8.8.8.8") // the 8.8.8.8 is the google server.
		output, err := cmd.Output()
		if err != nil {
			return "", err
		}
		out := string(output)
		nameKey := "Name:"
		idx := strings.Index(out, nameKey)
		if idx <= 0 {
			return "", nil // not found
		}

		addrKey := "Address:"
		end := strings.LastIndex(out, addrKey)
		if end <= 0 {
			return "", nil
		}

		idx += len(nameKey)
		if end <= idx {
			return "", nil // we should not panic..
		}

		out = out[idx:end]
		out = strings.TrimSpace(out)
		return out, nil
	}

	return "", nil
}

// no limits but we limit it to 42 million requests per minute.
var languagehttp = http.New(42000000, 1*time.Minute)

func languageServiceFromCountryCode(countryCode string) string {
	return fmt.Sprintf("https://restcountries.eu/rest/v2/alpha/%s?fields=languages;", countryCode)
}

type language struct {
	ISO6391 string `json:"iso639_1"`
	Name    string `json:"name"`
}

// FetchLanguage accepts a pointer to a Server table
// and sets its srv.Language and srv.LanguageCode based on the srv.CountryCode.
// https://restcountries.eu/rest/v2/alpha/$COUNTRYCODE?fields=languages;
func FetchLanguage(countryCode string, langInfo *LanguageInfo) error {
	if countryCode == "" {
		return nil
	}

	apiURL := languageServiceFromCountryCode(countryCode)

	var response = struct {
		Languages []language `json:"languages"`
		// it has a Name as well but we take it from the languages.
	}{}

	if err := languagehttp.ReadJSON(apiURL, &response); err != nil {
		return err
	}

	if len(response.Languages) == 0 {
		return nil
	}

	// we get only the first one.
	lang := response.Languages[0]
	langInfo.Language = lang.Name
	langInfo.LanguageCode = lang.ISO6391
	return nil
}
