package records

import (
	"fmt"
	"log"
	"net"
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
const IPv6PrefixOnlyValue = "prefix-only"
const IPv6HostIPValue = "host-ip"
const IPv6FritzBoxIPValue = "fritzbox-ip"

func Update(apikey string, secretkey string) {
	domainsString, present := os.LookupEnv(DomainsEnvKey)
	assert.Assert(present == true, "env should be present here because it's checked in main.validateEnvironment()")

	domains := strings.Split(domainsString, ",")

	if value, present := env.ReadOptionalEnv(IPv4EnvKey); value == "true" || present == false {
		// Either user set IPV4=true or he didn't set it at all (standard value)
		currentIPv4, err := wanip.GetFromFritzBox("ipv4")
		if err != nil {
			logger.Warnf("Retrieving current WAN IPv4 via FRITZ!Box failed.")
		} else {
			updateFinalIPRecords(domains, apikey, secretkey, currentIPv4, "A")
		}
	}

	if IPv6Value, _ := env.ReadOptionalEnv(IPv6EnvKey); IPv6Value == IPv6FritzBoxIPValue {
		// The user set IPV6=fritzbox-ip explicitly
		currentFritzboxIPv6, err := wanip.GetFromFritzBox("ipv6")
		if err != nil {
			logger.Warnf("Retrieving current WAN IPv6 of FRITZ!Box failed.")
		} else {
			updateFinalIPRecords(domains, apikey, secretkey, currentFritzboxIPv6, "AAAA")
		}
	} else if IPv6Value == IPv6HostIPValue {
		// The user set IPV6=host-ip explicitly
		currentHostIPv6, err := wanip.GetGlobalUnicastIPv6()
		if err != nil {
			logger.Warnf("Retrieving current host IPv6 failed. Is the host running on a (Docker) network with IPv6 support?")
		} else {
			updateFinalIPRecords(domains, apikey, secretkey, currentHostIPv6, "AAAA")
		}
	} else if IPv6Value == IPv6PrefixOnlyValue {
		// The user set IPV6=prefix-only explicitly
		currentIPv6Prefix, err := wanip.GetIPv6PrefixFromFritzBox()
		if err != nil {
			logger.Warnf("Retrieving current IPv6 prefix via FRITZ!Box failed.")
		} else {
			updateIPv6Prefix(domains, apikey, secretkey, currentIPv6Prefix)
		}
	}
}

func updateFinalIPRecords(domains []string, apikey string, secretkey string, currentIP string, recordType string) {
	for _, fqdn := range domains {
		if !isFQDNValid(fqdn) {
			logger.Warnf("%s is not a valid domain.", fqdn)
			return
		}

		subdomain, rootDomain := getSubAndRootDomain(fqdn)

		retrievedRecords, err := retrieveRecords(subdomain, rootDomain, recordType, apikey, secretkey)
		if err != nil {
			logger.Warnf("Skipping %s-Record update of %s because retrieval of active records failed.", recordType, fqdn)
			return
		}

		switch len(retrievedRecords) {
		case 0:
			createRecord(subdomain, rootDomain, recordType, currentIP, apikey, secretkey)
		case 1:
			oldRecord := retrievedRecords[0]
			if oldRecord.IP == currentIP {
				log.Printf("%s-Record of %s is up to date.", recordType, fqdn)
				continue
			}

			editRecord(subdomain, rootDomain, recordType, currentIP, apikey, secretkey, oldRecord.ID, oldRecord.IP)
		default:
			logger.Warnf("Multiple active %s-Records found for %s. Please clean up the DNS records in the Porkbun WebGUI or set the environment variable %s=%s to automatically unify them.",
				recordType, fqdn, mulRecordsEnvKey, mulRecordsUnifyValue)
		}
	}
}

func handleRetrievedRecordsConstIP() {

}

func handleRetrievedRecordsIPv6Prefix() {

}

func updateIPv6Prefix(domains []string, apikey string, secretkey string, currentIPv6Prefix string) {
	recordType := "AAAA"

	for _, fqdn := range domains {
		if !isFQDNValid(fqdn) {
			logger.Warnf("%s is not a valid domain.", fqdn)
			return
		}

		subdomain, rootDomain := getSubAndRootDomain(fqdn)

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
				continue
			}

			editRecord(subdomain, rootDomain, recordType, IPv6Addr, apikey, secretkey, oldRecord.ID, oldRecord.IP)
		default:
			logger.Warnf("Multiple active %s-Records found for %s. Can only edit existing %[1]s-Records with %[3]s=%s.",
				recordType, fqdn, IPv6EnvKey, IPv6PrefixOnlyValue)
		}
	}
}

// combineIPv6PrefixAndInterfaceID combines an IPv6 prefix and an abitrary IPv6 address to a full IPv6 address.
// Example: combineIPv6PrefixAndInterfaceID("2001:db8::", "::1234:5678:90ab:cdef:0123") returns "2001:db8::5678:90ab:cdef:0123".
// Example: combineIPv6PrefixAndInterfaceID("2001:db8::", "fe80:efef:db8:1234:5678:90ab:cdef:0123") returns "2001:db8::5678:90ab:cdef:0123".
// Example: combineIPv6PrefixAndInterfaceID("2001:efef:db8:1234:", "fe80:efef:db8:1234:5678:90ab:cdef:0123") returns "2001:efef:db8:1234:5678:90ab:cdef:0123".
// TODO: Add tests for this function. Change implementation (e.g. account for different prefix lengths).
func combineIPv6PrefixAndInterfaceID(prefix string, IPv6 string) string {
	assert.Assert(strings.HasSuffix(prefix, "::"), "prefix should end with '::'")

	addr := net.ParseIP(IPv6)
	assert.Assert(addr != nil, "IPv6 should be a valid IP address")

	addr = addr.To16()
	assert.Assert(addr != nil, "IPv6 should be a valid IPv6 address")

	// Get the interface ID from the IPv6 address
	// Example: 2001:db8:abcd:1234:5678:90ab:cdef:0123
	interfaceID := fmt.Sprintf("%x:%x:%x:%x",
		uint16(addr[8])<<8|uint16(addr[9]), // uint16(addr[8])<<8 = 0x5600, addr[9] = 0x78, => 0x5678
		uint16(addr[10])<<8|uint16(addr[11]),
		uint16(addr[12])<<8|uint16(addr[13]),
		uint16(addr[14])<<8|uint16(addr[15]),
	)

	// Prepend prefix to interface ID
	return fmt.Sprintf("%s%s", prefix[:len(prefix)-1], interfaceID)
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
