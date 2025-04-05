package records

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"

	"bjoernblessin.de/gorkbunddns/src/shared"
	"bjoernblessin.de/gorkbunddns/src/util/assert"
	"bjoernblessin.de/gorkbunddns/src/util/env"
	"bjoernblessin.de/gorkbunddns/src/util/logger"
	"bjoernblessin.de/gorkbunddns/src/wanip"
)

const DomainsEnvKey = "DOMAINS"
const mulRecordsEnvKey = "MULTIPLE_RECORDS"
const mulRecordsUnifyValue = "unify"
const IPv4EnvKey = "IPV4"
const IPv6EnvKey = "IPV6"
const IPv6PrefixOnlyValue = "prefix-only"
const IPv6HostIPValue = "host-ip"
const IPv6FritzBoxIPValue = "fritzbox-ip"

func Update(apikey string, secretkey string) {
	domainsString, present := os.LookupEnv(DomainsEnvKey)
	assert.Assert(present, "env should be present here because it's checked in main.validateEnvironment()")

	domains := strings.Split(domainsString, ",")

	IPv4Value, IPv4ValuePresent := env.ReadOptionalEnv(IPv4EnvKey)
	var currentIPv4 string
	var IPv4Err error

	if IPv4Value == "true" || !IPv4ValuePresent {
		// Either user set IPV4=true or he didn't set it at all (standard value)
		currentIPv4, IPv4Err = wanip.GetFromFritzBox("ipv4")
		if IPv4Err != nil {
			logger.Warnf("Retrieving current WAN IPv4 via FRITZ!Box failed.")
		}
	}

	IPv6Value, _ := env.ReadOptionalEnv(IPv6EnvKey)
	var currentFritzboxIPv6, currentHostIPv6, currentIPv6Prefix string
	var IPv6Err error

	if IPv6Value == IPv6FritzBoxIPValue {
		// The user set IPV6=fritzbox-ip explicitly
		currentFritzboxIPv6, IPv6Err = wanip.GetFromFritzBox("ipv6")
		if IPv6Err != nil {
			logger.Warnf("Retrieving current WAN IPv6 of FRITZ!Box failed.")
		}
	} else if IPv6Value == IPv6HostIPValue {
		// The user set IPV6=host-ip explicitly
		currentHostIPv6, IPv6Err = wanip.GetGlobalUnicastIPv6()
		if IPv6Err != nil {
			logger.Warnf("Retrieving current host IPv6 failed. Is the host running on a (Docker) network with IPv6 support?")
		}
	} else if IPv6Value == IPv6PrefixOnlyValue {
		// The user set IPV6=prefix-only explicitly
		currentIPv6Prefix, IPv6Err = wanip.GetIPv6PrefixFromFritzBox()
		if IPv6Err != nil {
			logger.Warnf("Retrieving current IPv6 prefix via FRITZ!Box failed.")
		}
	}

	for _, fqdn := range domains {
		if !isFQDNValid(fqdn) {
			logger.Warnf("%s is not a valid domain.", fqdn)
			return
		}

		subdomain, rootDomain := getSubAndRootDomain(fqdn)

		if (IPv4Value == "true" || !IPv4ValuePresent) && IPv4Err == nil {
			assert.Assert(currentIPv4 != "", "currentIPv4 should be set here because it's checked in the beginning of this function")
			tryUpdateRecordWithConstIP(currentIPv4, "A", fqdn, subdomain, rootDomain, apikey, secretkey)
		}

		if IPv6Value == IPv6FritzBoxIPValue && IPv6Err == nil {
			assert.Assert(currentFritzboxIPv6 != "", "currentFritzboxIPv6 should be set here because it's checked in the beginning of this function")
			tryUpdateRecordWithConstIP(currentFritzboxIPv6, "AAAA", fqdn, subdomain, rootDomain, apikey, secretkey)
		} else if IPv6Value == IPv6HostIPValue && IPv6Err == nil {
			assert.Assert(currentHostIPv6 != "", "currentHostIPv6 should be set here because it's checked in the beginning of this function")
			tryUpdateRecordWithConstIP(currentHostIPv6, "AAAA", fqdn, subdomain, rootDomain, apikey, secretkey)
		} else if IPv6Value == IPv6PrefixOnlyValue && IPv6Err == nil {
			assert.Assert(currentIPv6Prefix != "", "currentIPv6Prefix should be set here because it's checked in the beginning of this function")
			tryUpdateRecordWithIPv6Prefix(currentIPv6Prefix, fqdn, subdomain, rootDomain, apikey, secretkey)
		}
	}
}

func tryUpdateRecordWithConstIP(currentIP string, recordType string, fqdn string, subdomain string, rootDomain string, apikey string, secretkey string) {
	retrievedRecords, err := retrieveRecords(subdomain, rootDomain, recordType, apikey, secretkey)
	if err != nil {
		logger.Warnf("Skipping %s-Record update of %s because retrieval of active records failed. %s", recordType, fqdn, err)
		return
	}

	switch len(retrievedRecords) {
	case 0:
		createRecord(subdomain, rootDomain, recordType, currentIP, apikey, secretkey)
	case 1:
		oldRecord := retrievedRecords[0]
		if oldRecord.IP == currentIP {
			log.Printf("%s-Record of %s is up to date.", recordType, fqdn)
			return
		}

		editRecord(subdomain, rootDomain, recordType, currentIP, apikey, secretkey, oldRecord.ID, oldRecord.IP)
	default:
		logger.Warnf("Multiple active %s-Records found for %s. Please clean up the DNS records in the Porkbun WebGUI or set the environment variable %s=%s to automatically unify them.",
			recordType, fqdn, mulRecordsEnvKey, mulRecordsUnifyValue)
	}
}

func tryUpdateRecordWithIPv6Prefix(currentIPv6Prefix string, fqdn string, subdomain string, rootDomain string, apikey string, secretkey string) {
	recordType := "AAAA"

	retrievedRecords, err := retrieveRecords(subdomain, rootDomain, recordType, apikey, secretkey)
	if err != nil {
		logger.Warnf("Skipping %s-Record update of %s because retrieval of active records failed.", recordType, fqdn)
		return
	}

	switch len(retrievedRecords) {
	case 0:
		logger.Warnf("No %s-Record found for %s. Can only edit existing %[1]s-Records with %[3]s=%s.", recordType, fqdn, IPv6EnvKey, IPv6PrefixOnlyValue)
	case 1:
		oldRecord := retrievedRecords[0]

		IPv6Addr := combineIPv6PrefixAndInterfaceID(currentIPv6Prefix, oldRecord.IP)

		if oldRecord.IP == IPv6Addr {
			log.Printf("%s-Record of %s is up to date.", recordType, fqdn)
			return
		}

		editRecord(subdomain, rootDomain, recordType, IPv6Addr, apikey, secretkey, oldRecord.ID, oldRecord.IP)
	default:
		logger.Warnf("Multiple active %s-Records found for %s. Can only edit existing %[1]s-Records with %[3]s=%s.",
			recordType, fqdn, IPv6EnvKey, IPv6PrefixOnlyValue)
	}
}

// combineIPv6PrefixAndInterfaceID combines an the prefix of an IPv6 address and the interface ID of another IPv6 address to a combined IPv6 address.
// The IPv6 addresses should be RFC 5952 ("2001:db8::1") compliant.
// The returned IPv6 address is also RFC 5952 compliant.
// Example: combineIPv6PrefixAndInterfaceID("2001:db8::", "fe80:efef:db8:1234:5678:90ab:cdef:0123") returns "2001:db8::5678:90ab:cdef:123".
func combineIPv6PrefixAndInterfaceID(prefixIPv6 string, interfaceIDIPv6 string) string {
	prefixAddr := net.ParseIP(prefixIPv6)
	assert.Assert(prefixAddr != nil, "prefixIPv6 should be a valid IP address")
	prefixAddr = prefixAddr.To16()
	assert.Assert(prefixAddr != nil, "prefixIPv6 should be a valid IPv6 address")

	prefix := fmt.Sprintf("%x:%x:%x:%x",
		uint16(prefixAddr[0])<<8|uint16(prefixAddr[1]),
		uint16(prefixAddr[2])<<8|uint16(prefixAddr[3]),
		uint16(prefixAddr[4])<<8|uint16(prefixAddr[5]),
		uint16(prefixAddr[6])<<8|uint16(prefixAddr[7]),
	)

	interfaceIDAddr := net.ParseIP(interfaceIDIPv6)
	assert.Assert(interfaceIDAddr != nil, "interfaceIDIPv6 should be a valid IP address")
	interfaceIDAddr = interfaceIDAddr.To16()
	assert.Assert(interfaceIDAddr != nil, "interfaceIDIPv6 should be a valid IPv6 address")

	// Get the interface ID from the IPv6 address
	// Example: 2001:db8:abcd:1234:5678:90ab:cdef:0123
	interfaceID := fmt.Sprintf("%x:%x:%x:%x",
		uint16(interfaceIDAddr[8])<<8|uint16(interfaceIDAddr[9]), // uint16(addr[8])<<8 = 0x5600, addr[9] = 0x78, => 0x5678
		uint16(interfaceIDAddr[10])<<8|uint16(interfaceIDAddr[11]),
		uint16(interfaceIDAddr[12])<<8|uint16(interfaceIDAddr[13]),
		uint16(interfaceIDAddr[14])<<8|uint16(interfaceIDAddr[15]),
	)

	netIP := net.ParseIP(fmt.Sprintf("%s:%s", prefix, interfaceID))
	assert.Assert(netIP != nil, "combined IPv6 address should be a valid IP address")

	return netIP.String()
}

// getSubAndRootDomain splits a fully qualified domain name into subdomain and root domain.
// getSubAndRootDomain("sub.example.com") returns "sub" and "example.com".
func getSubAndRootDomain(fqdn string) (subdomain string, rootDomain string) {
	domainParts := strings.Split(fqdn, ".")
	rootDomain = strings.Join(domainParts[len(domainParts)-2:], ".")
	subdomain = strings.Join(domainParts[:len(domainParts)-2], ".")
	return subdomain, rootDomain
}

// isFQDNValid checks if fqdn is likely a valid fully qualified domain name.
// FQDN is guaranteed to match ^.*[a-zA-Z0-9]\\.[a-zA-Z]{2,}$.
func isFQDNValid(fqdn string) bool {
	matched, err := regexp.MatchString("^.*[a-zA-Z0-9]\\.[a-zA-Z]{2,}$", fqdn)
	assert.IsNil(err)

	return matched
}

type retrievedRecord struct {
	ID string
	IP string
}

// retrieveRecords gets the active record IDs and their associated IPs for a given FQDN and record type.
// There may be zero, one, or multiple active records, each with different answers.
func retrieveRecords(subdomain string, rootDomain string, recordType string, apikey string, secretkey string) ([]retrievedRecord, error) {
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

	requestBody := shared.RequestCredentials{SecretAPIKey: secretkey, APIKey: apikey}
	jsonBody, err := json.Marshal(requestBody)
	assert.IsNil(err)

	resp, err := http.Post(fmt.Sprintf("https://api.porkbun.com/api/json/v3/dns/retrieveByNameType/%s/%s/%s",
		rootDomain, recordType, subdomain), "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return []retrievedRecord{}, fmt.Errorf("Could not retrieve currently active %s-Records for %s.%s. %w", recordType, subdomain, rootDomain, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// May happen! For example: 503 Service Temporarily Unavailable
		return []retrievedRecord{}, fmt.Errorf("Something unexpected happened while retrieving active %s-Records for %s.%s.", recordType, subdomain, rootDomain)
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

// createRecord request the Porkbun server to create a specific record.
//
// Valid recordTypes are "A", "MX", "CNAME", "ALIAS", "TXT", "NS", "AAAA", "SRV", "TLSA", "CAA", "HTTPS", "SVCB"
func createRecord(subdomain string, rootDomain string, recordType string, newIP string, apikey string, secretkey string) {
	type createRequest struct {
		shared.RequestCredentials
		Name    string `json:"name"`
		Type    string `json:"type"`
		Content string `json:"content"`
	}

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

// editRecord updates the record matching id.
// The subdomain, ?rootDomain? and IP will be changed accordingly.
// After execution and if the Porkbun server accepted the request, one record will point the IP. Note: this does not mean, that the edit was successful, neither that the record matching id will point to the IP.
func editRecord(subdomain string, rootDomain string, recordType string, newIP string, apikey string, secretkey string, id string, oldIP string) {
	type editRequest struct {
		shared.RequestCredentials
		Name    string `json:"name"`
		Type    string `json:"type"`
		Content string `json:"content"`
	}

	requestBody := editRequest{RequestCredentials: shared.RequestCredentials{SecretAPIKey: secretkey, APIKey: apikey}, Name: subdomain, Type: recordType, Content: newIP}
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

	log.Printf("%s-Record of %s.%s updated: %s -> %s.", recordType, subdomain, rootDomain, oldIP, newIP)
}
