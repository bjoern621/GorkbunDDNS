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
	Status  string `json:"status"`
	Records []struct {
		Id      string `json:"id"`
		Name    string `json:"name"`
		Type    string `json:"type"`
		Content string `json:"content"`
		Ttl     string `json:"ttl"`
		Prio    string `json:"prio"`
		Notes   string `json:"notes"`
	} `json:"records"`
}

type retrievedRecord struct {
	ID string
	IP string
}

// retrieveRecords gets the active record IDs and their associtated IPs for a given FQDN and record type.
// There may be zero, one, or multiple active records, each with different answers.
// If something fails, success is set to false.
func retrieveRecords(subdomain string, rootDomain string, recordType string, apikey string, secretkey string) ([]retrievedRecord, error) {
	requestBody := shared.RequestCredentials{SecretAPIKey: secretkey, APIKey: apikey}
	jsonBody, err := json.Marshal(requestBody)
	assert.IsNil(err)

	resp, err := http.Post(fmt.Sprintf("https://api.porkbun.com/api/json/v3/dns/retrieveByNameType/%s/%s/%s",
		rootDomain, recordType, subdomain), "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return []retrievedRecord{}, fmt.Errorf("could not retrieve currently active %s-Records for %s.%s. %w", recordType, subdomain, rootDomain, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []retrievedRecord{}, fmt.Errorf("something unexpected happened while retrieving active %s-Records for %s.%s.", recordType, subdomain, rootDomain)
	}

	var response retrieveResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return []retrievedRecord{}, fmt.Errorf("Porkbun server returned invalid JSON format while retrieving active %s-Records for %s.%s. %w", recordType, subdomain, rootDomain, err)
	}

	var records []retrievedRecord
	for _, record := range response.Records {
		records = append(records, retrievedRecord{ID: record.Id, IP: record.Content})
	}

	return records, nil
}
