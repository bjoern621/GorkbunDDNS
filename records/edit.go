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

type _DNSEditRequest struct {
	shared.RequestCredentials
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

// editRecord updates the record matching id.
// The subdomain, ?rootDomain? and IP will be changed accordingly.
// After execution one record will point the IP. Note: this doesn't mean, that the edit was successful, neither that the record matching id will point to the IP.
func editRecord(subdomain string, rootDomain string, recordType string, newIP string, apikey string, secretkey string, id string) {
	requestBody := _DNSEditRequest{RequestCredentials: shared.RequestCredentials{SecretAPIKey: secretkey, APIKey: apikey}, Name: subdomain, Type: recordType, Content: newIP}
	jsonBody, err := json.Marshal(requestBody)
	assert.IsNil(err)

	var resp *http.Response
	totalTries := 3

	for i := 1; i <= totalTries; i++ {
		resp, err = http.Post(fmt.Sprintf("https://api.porkbun.com/api/json/v3/dns/edit/%s/%s", rootDomain, id), "application/json", bytes.NewReader(jsonBody))
		if err != nil {
			logger.Warnf("Edit attempt %d/%d failed: %v", i, totalTries, err)
		} else {
			resp.Body.Close()
			break
		}
	}

	if err != nil {
		logger.Warnf("All attempts failed.")
		return
	}

	if resp.StatusCode != http.StatusOK {
		logger.Warnf("Could not update %s-Record of %s.%s.", recordType, subdomain, rootDomain)
		return
	}

	log.Printf("%s-Record of %s.%s updated to %s.", recordType, subdomain, rootDomain, newIP)
}
