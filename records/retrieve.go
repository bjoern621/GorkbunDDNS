package records

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"bjoernblessin.de/gorkbunddns/shared"
	"bjoernblessin.de/gorkbunddns/util/assert"
)

type retrieveResponse struct {
	Status  string   `json:"status"`
	Records []record `json:"records"`
}

type record struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Ttl     string `json:"ttl"`
	Prio    string `json:"prio"`
	Notes   string `json:"notes"`
}

type Record struct {
	ID string
	IP string
}

// retrieveRecords gets the active record IDs and their associtated IPs for a given FQDN and record type.
// There may be zero, one, or multiple active records, each with different answers.
// If something fails, success is set to false.
func retrieveRecords(subdomain string, rootDomain string, recordType string, apikey string, secretkey string) (records []Record, success bool) {
	requestBody := shared.RequestCredentials{SecretAPIKey: secretkey, APIKey: apikey}
	jsonBody, err := json.Marshal(requestBody)
	assert.IsNil(err, "could not encode data to JSON")

	resp, err := http.Post(fmt.Sprintf("https://api.porkbun.com/api/json/v3/dns/retrieveByNameType/%s/%s/%s",
		rootDomain, recordType, subdomain), "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		// logger.Warnf("Could not retrieve currently active %s-Records for %s.%s.", recordType, subdomain, rootDomain)
		return []Record{}, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// logger.Warnf("Something unexpected happened while retrieving active %s-Records for %s.%s.", recordType, subdomain, rootDomain)
		return []Record{}, false
	}

	var response retrieveResponse
	if json.NewDecoder(resp.Body).Decode(&response) != nil {
		// logger.Warnf("Porkbun server returned invalid JSON format while retrieving active %s-Records for %s.%s.", recordType, subdomain, rootDomain)
		return []Record{}, false
	}

	for _, record := range response.Records {
		records = append(records, Record{ID: record.Id, IP: record.Content})
	}

	return records, true
}
