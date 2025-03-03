package records

import (
	"log"
	"os"
	"regexp"
	"strings"

	"bjoernblessin.de/gorkbunddns/util/assert"
	"bjoernblessin.de/gorkbunddns/util/env"
	"bjoernblessin.de/gorkbunddns/util/logger"
)

const DomainsEnvKey = "DOMAINS"
const mulRecordsEnvKey = "MULTIPLE_RECORDS"
const mulRecordsUnifyValue = "unify"
const ipv4EnvKey = "IPV4"
const ipv6EnvKey = "IPV6"

func Update(apikey string, secretkey string) {
	domainsString, present := os.LookupEnv(DomainsEnvKey)
	assert.Assert(present == true, "env should be present here because it's checked in main.validateEnvironment()")

	domains := strings.Split(domainsString, ",")

	// todo: get current ip
	currentIP := "2.2.2.2"

	recordTypes := []string{}
	if value, present := env.ReadOptionalEnv(ipv4EnvKey); value == "true" || present == false {
		recordTypes = append(recordTypes, "A") // Either user set IPV4=true or he didn't set it at all (standard value)
	}
	if value, _ := env.ReadOptionalEnv(ipv6EnvKey); value == "true" {
		recordTypes = append(recordTypes, "AAAA") // The user set IPV6=true explicitly
	}

	for _, fqdn := range domains {
		valid, subdomain, rootDomain := isFQDNValid(fqdn)
		if !valid {
			logger.Warnf("%s is not a valid domain, skipping.", fqdn)
			return
		}

		for _, recordType := range recordTypes {
			activeRecordIDs, success := retrieveRecords(subdomain, rootDomain, recordType, apikey, secretkey)
			if success == false {
				logger.Warnf("Skipping %s-Record update of %s because retrieval of active records failed.", recordType, fqdn)
				continue
			}

			switch len(activeRecordIDs) {
			case 0:
				createRecord(subdomain, rootDomain, recordType, currentIP, apikey, secretkey)
			case 1:
				if activeRecordIDs[0].IP == currentIP {
					log.Printf("%s-Record of %s.%s is up to date.", recordType, subdomain, rootDomain)
					continue
				}

				editRecord(subdomain, rootDomain, recordType, currentIP, apikey, secretkey, activeRecordIDs[0].ID)
			default:
				logger.Warnf("Multiple active %s-Records found for %s.%s. Please clean up the DNS records in the Porkbun WebGUI or set the environment variable %s=%s to automatically unify them.",
					recordType, subdomain, rootDomain, mulRecordsEnvKey, mulRecordsUnifyValue)
			}
		}
	}
}

// isFQDNValid checks if fqdn is likely a valid fully qualified domain name and if so, returns sub- and root domain.
// FQDN is guaranteed to match ^.*[a-zA-Z0-9]\\.[a-zA-Z]{2,}$.
func isFQDNValid(fqdn string) (valid bool, subdomain string, rootDomain string) {
	matched, err := regexp.MatchString("^.*[a-zA-Z0-9]\\.[a-zA-Z]{2,}$", fqdn)
	assert.IsNil(err, "regex match should only fail if regex is invalid")
	if !matched {
		return false, "", ""
	}

	domainParts := strings.Split(fqdn, ".")
	rootDomain = strings.Join(domainParts[len(domainParts)-2:], ".")
	subdomain = strings.Join(domainParts[:len(domainParts)-2], ".")

	return true, subdomain, rootDomain
}
