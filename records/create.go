package records

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"bjoernblessin.de/gorkbunddns/shared"
	"bjoernblessin.de/gorkbunddns/util/assert"
	"bjoernblessin.de/gorkbunddns/util/logger"
)

type createRequest struct {
	shared.RequestCredentials
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

// createRecord request the Porkbun server to create a specific record.
//
// Valid recordTypes are "A", "MX", "CNAME", "ALIAS", "TXT", "NS", "AAAA", "SRV", "TLSA", "CAA", "HTTPS", "SVCB"
func createRecord(subdomain string, rootDomain string, recordType string, newIP string, apikey string, secretkey string) {
	requestBody := createRequest{RequestCredentials: shared.RequestCredentials{SecretAPIKey: secretkey, APIKey: apikey}, Name: subdomain, Type: recordType, Content: newIP}
	jsonBody, err := json.Marshal(requestBody)
	assert.IsNil(err)

	resp, err := http.Post(fmt.Sprintf("https://api.porkbun.com/api/json/v3/dns/create/%s", rootDomain), "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		logger.Warnf("Could not create %s-Record for %s.%s.", recordType, subdomain, rootDomain)
		return
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warnf("Could not create %s-Record for %s.%s.", recordType, subdomain, rootDomain)
		return
	}

	log.Printf("%s-Record for %s.%s created. New IP: %s.", recordType, subdomain, rootDomain, newIP)
}
