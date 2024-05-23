// urlHandler
package urlHandler

import (
	"crypto/rand"
	"encoding/base64"
	"net/url"
	"regexp"
	"strings"
)

func GenerateShortUrl() (string, error) {

	length := int((11 * 3) / 4)

	bytes := make([]byte, length)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil

}

func IsValidUrl(longUrl *string) bool {

	*longUrl = strings.ToLower(*longUrl)

	//if *longUrl > 2048 char not valide
	if len(*longUrl) > 2048 || len(*longUrl) == 0 {
		return false
	}

	if !strings.HasPrefix(*longUrl, "http://") && !strings.HasPrefix(*longUrl, "https://") {
		*longUrl = "http://" + *longUrl
	}

	parsedURL, err := url.Parse(*longUrl)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}
	if parsedURL.Host == "" {
		return false
	}

	hostPattern := regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !hostPattern.MatchString(parsedURL.Host) {
		return false
	}
	return true

}
