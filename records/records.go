package records

import (
	"log"
	"os"
	"regexp"
	"strings"

	"bjoernblessin.de/gorkbunddns/util/assert"
	"bjoernblessin.de/gorkbunddns/util/env"
	"bjoernblessin.de/gorkbunddns/util/logger"
	"bjoernblessin.de/gorkbunddns/wanip"
)

const DomainsEnvKey = "DOMAINS"
const mulRecordsEnvKey = "MULTIPLE_RECORDS"
const mulRecordsUnifyValue = "unify"
const IPv4EnvKey = "IPV4"
const IPv6EnvKey = "IPV6"

func Update(apikey string, secretkey string) {
	domainsString, present := os.LookupEnv(DomainsEnvKey)
	assert.Assert(present == true, "env should be present here because it's checked in main.validateEnvironment()")

	domains := strings.Split(domainsString, ",")

	selectedIPVersionInfos := []struct {
		ipProtocol string
		recordType string
		currentIP  string
	}{}
	if value, present := env.ReadOptionalEnv(IPv4EnvKey); value == "true" || present == false {
		// Either user set IPV4=true or he didn't set it at all (standard value)
		currentIP, err := wanip.GetFromFritzBox("ipv4")
		if err != nil {
			logger.Warnf("Retrieving current WAN IPv4 via FRITZ!Box failed.")
		} else {
			selectedIPVersionInfos = append(selectedIPVersionInfos, struct {
				ipProtocol string
				recordType string
				currentIP  string
			}{ipProtocol: "ipv4", recordType: "A", currentIP: currentIP})
		}
	}
	if value, _ := env.ReadOptionalEnv(IPv6EnvKey); value == "true" {
		// The user set IPV6=true explicitly
		currentIP, err := wanip.GetFromFritzBox("ipv6")
		if err != nil {
			logger.Warnf("Retrieving current WAN IPv6 via FRITZ!Box failed.")
		} else {
			selectedIPVersionInfos = append(selectedIPVersionInfos, struct {
				ipProtocol string
				recordType string
				currentIP  string
			}{ipProtocol: "ipv6", recordType: "AAAA", currentIP: currentIP})
		}
	}

	for _, fqdn := range domains {
		valid, subdomain, rootDomain := isFQDNValid(fqdn)
		if !valid {
			logger.Warnf("%s is not a valid domain.", fqdn)
			return
		}

		for _, IPVersionInfo := range selectedIPVersionInfos {
			activeRecordIDs, err := retrieveRecords(subdomain, rootDomain, IPVersionInfo.recordType, apikey, secretkey)
			if err != nil {
				logger.Warnf("Skipping %s-Record update of %s because retrieval of active records failed.", IPVersionInfo.recordType, fqdn)
				continue
			}

			switch len(activeRecordIDs) {
			case 0:
				createRecord(subdomain, rootDomain, IPVersionInfo.recordType, IPVersionInfo.currentIP, apikey, secretkey)
			case 1:
				oldRecord := activeRecordIDs[0]
				if oldRecord.IP == IPVersionInfo.currentIP {
					// TODO: %s.%s returns .example.de for [] and example.de, expected: example.com, fix with getFQDNString(subdomain, rootDomain)
					log.Printf("%s-Record of %s is up to date.", IPVersionInfo.recordType, fqdn)
					continue
				}

				editRecord(subdomain, rootDomain, IPVersionInfo.recordType, IPVersionInfo.currentIP, apikey, secretkey, oldRecord.ID, oldRecord.IP)
			default:
				logger.Warnf("Multiple active %s-Records found for %s. Please clean up the DNS records in the Porkbun WebGUI or set the environment variable %s=%s to automatically unify them.",
					IPVersionInfo.recordType, fqdn, mulRecordsEnvKey, mulRecordsUnifyValue)
			}
		}
	}
}

// isFQDNValid checks if fqdn is likely a valid fully qualified domain name and if so, returns sub- and root domain.
// FQDN is guaranteed to match ^.*[a-zA-Z0-9]\\.[a-zA-Z]{2,}$.
func isFQDNValid(fqdn string) (valid bool, subdomain string, rootDomain string) {
	matched, err := regexp.MatchString("^.*[a-zA-Z0-9]\\.[a-zA-Z]{2,}$", fqdn)
	assert.IsNil(err)
	if !matched {
		return false, "", ""
	}

	domainParts := strings.Split(fqdn, ".")
	rootDomain = strings.Join(domainParts[len(domainParts)-2:], ".")
	subdomain = strings.Join(domainParts[:len(domainParts)-2], ".")

	return true, subdomain, rootDomain
}

func getFQDNString(subdomain string, rootDomain string) string {
	return "abc"
}
