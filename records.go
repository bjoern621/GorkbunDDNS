package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"bjoernblessin.de/gorkbunddns/util/assert"
	"bjoernblessin.de/gorkbunddns/util/env"
	"bjoernblessin.de/gorkbunddns/util/logger"
)

type DNSEditRequest struct {
	SecretAPIKey string `json:"secretapikey"`
	APIKey       string `json:"apikey"`
	Content      string `json:"content"`
}

const domainsEnvKey = "DOMAINS"

func UpdateDNSRecords(apikey string, secretkey string) {
	domainsString := env.ReadNonEmptyRequiredEnv(domainsEnvKey)
	domains := strings.Split(domainsString, ",")

	// for _, fqdn := range domains {
	// 	checkRecord()
	// }

	log.Printf("%v", domains)

	domain := domains[0]
	updateDNSRecord(domain, apikey, secretkey)
}

// IsFQDNValid checks if fqdn is likely a valid fully qualified domain name and if so, returns sub- and root domain.
// FQDN is guaranteed to match ^.*[a-zA-Z0-9]\\.[a-zA-Z]{2,}$.
func isFQDNValid(fqdn string) (valid bool, subdomain string, rootDomain string) {
	matched, err := regexp.MatchString("^.*[a-zA-Z0-9]\\.[a-zA-Z]{2,}$", fqdn)
	assert.IsNil(err, "Regex match should not fail if regex is valid.")
	if !matched {
		return false, "", ""
	}

	domainParts := strings.Split(fqdn, ".")
	rootDomain = strings.Join(domainParts[len(domainParts)-2:], ".")
	subdomain = strings.Join(domainParts[:len(domainParts)-2], ".")

	return true, subdomain, rootDomain
}

func updateDNSRecord(fqdn string, apikey string, secretkey string) {
	valid, subdomain, rootDomain := isFQDNValid(fqdn)
	if !valid {
		logger.Warnf("%s is not a valid domain, skipping.", fqdn)
		return
	}

	log.Printf("%s", rootDomain)
	log.Printf("%s", subdomain)

	log.Printf("[DEBUG] Updating %s", fqdn)

	recordType := "A"

	requestBody := DNSEditRequest{SecretAPIKey: secretkey, APIKey: apikey, Content: "2.2.2.2"}
	jsonBody, err := json.Marshal(requestBody)
	assert.IsNil(err, "Could not encode data to JSON.")

	log.Printf("%s", jsonBody)

	var resp *http.Response
	totalTries := 3

	for i := 1; i <= totalTries; i++ {
		resp, err = http.Post(fmt.Sprintf("https://api.porkbun.com/api/json/v3/dns/editByNameType/%s/%s/%s", rootDomain, recordType, subdomain), "application/json", bytes.NewReader(jsonBody))
		if err != nil {
			logger.Warnf("Attempt %d/%d failed: %v", i, totalTries, err)
		} else {
			defer resp.Body.Close()
			break
		}
	}

	if err != nil {
		logger.Warnf("All attempts failed. Skipping (sub)domain.")
		return
	}

	if resp.StatusCode != 200 {
		prettyJSON := JSONResponseBodyToPrettyByteArray(resp.Body)

		logger.Warnf("Could not update A-Record of %s:\n%s\nSkipping (sub)domain.", fqdn, prettyJSON)
		return
	}
	prettyJSON := JSONResponseBodyToPrettyByteArray(resp.Body)
	log.Printf("A-Record of %s updated to 1.1.1.1\n%s", fqdn, prettyJSON)
}

func isRecordPresent(subdomain string, rootDomain string) {

}

func editRecord(subdomain string, rootDomain string, recordType, newIp string) {

}

func createRecord(subdomain string, rootDomain string, recordType, newIp string) {

}
