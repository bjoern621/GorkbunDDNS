package wanip

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"bjoernblessin.de/gorkbunddns/util/assert"
)

type _IPv4ResponseEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		GetExternalIPAddressResponse struct {
			NewExternalIPAddress string `xml:"NewExternalIPAddress"`
		} `xml:"GetExternalIPAddressResponse"`
	} `xml:"Body"`
}

type _IPv6ResponseEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		X_AVM_DE_GetExternalIPv6AddressResponse struct {
			NewExternalIPv6Address string `xml:"NewExternalIPv6Address"`
		} `xml:"X_AVM_DE_GetExternalIPv6AddressResponse"`
	} `xml:"Body"`
}

// GetFromFritzBox sends a TR-064 SOAP request to the FRITZ!Box to retrieve the current WAN IP address.
//
// ipProtocol is either "ipv4" or "ipv6".
func GetFromFritzBox(ipProtocol string) (string, error) {
	assert.Assert(ipProtocol == "ipv4" || ipProtocol == "ipv6", "ipProtocol must be \"ipv4\" or \"ipv6\"")

	soapRequest := `<?xml version="1.0" encoding="utf-8"?>
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:wan="urn:schemas-upnp-org:service:WANIPConnection:1">
	   <soapenv:Header/>
	   <soapenv:Body>
	      <wan:GetExternalIPAddress/>
	   </soapenv:Body>
	</soapenv:Envelope>`

	request, err := http.NewRequest("POST", "http://fritz.box:49000/igdupnp/control/WANIPConn1", bytes.NewBuffer([]byte(soapRequest)))
	assert.IsNil(err)

	request.Header.Set("Content-Type", "text/xml; charset=utf-8")
	if ipProtocol == "ipv4" {
		request.Header.Set("SOAPACTION", "urn:schemas-upnp-org:service:WANIPConnection:1#GetExternalIPAddress")
	} else {
		request.Header.Set("SOAPACTION", "urn:schemas-upnp-org:service:WANIPConnection:1#X_AVM_DE_GetExternalIPv6Address")
	}

	resp, err := (&http.Client{}).Do(request)
	if err != nil {
		return "", fmt.Errorf("Error sending request %w", err)
	}
	defer resp.Body.Close()

	if ipProtocol == "ipv4" {
		var response _IPv4ResponseEnvelope

		err = xml.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return "", fmt.Errorf("Couldn't parse XML %w", err)
		}

		return response.Body.GetExternalIPAddressResponse.NewExternalIPAddress, nil
	} else {
		var response _IPv6ResponseEnvelope

		err = xml.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return "", fmt.Errorf("Couldn't parse XML %w", err)
		}

		return response.Body.X_AVM_DE_GetExternalIPv6AddressResponse.NewExternalIPv6Address, nil
	}
}

type _IPv6PrefixResponseEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		X_AVM_DE_GetIPv6PrefixResponse struct {
			NewIPv6Prefix string `xml:"NewIPv6Prefix"`
		} `xml:"X_AVM_DE_GetIPv6PrefixResponse"`
	} `xml:"Body"`
}

// GetIPv6PrefixFromFritzBox sends a TR-064 SOAP request to the FRITZ!Box to retrieve the IPv6 prefix for the local network.
// The prefix is in the form of "2001:db8:1234:5678::".
func GetIPv6PrefixFromFritzBox() (string, error) {
	soapRequest := `<?xml version="1.0" encoding="utf-8"?>
    <soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:wan="urn:schemas-upnp-org:service:WANIPConnection:1">
       <soapenv:Header/>
       <soapenv:Body>
          <wan:X_AVM_DE_GetIPv6Prefix/>
       </soapenv:Body>
    </soapenv:Envelope>`

	request, err := http.NewRequest("POST", "http://fritz.box:49000/igdupnp/control/WANIPConn1", bytes.NewBuffer([]byte(soapRequest)))
	assert.IsNil(err)

	request.Header.Set("Content-Type", "text/xml; charset=utf-8")
	request.Header.Set("SOAPACTION", "urn:schemas-upnp-org:service:WANIPConnection:1#X_AVM_DE_GetIPv6Prefix")

	resp, err := (&http.Client{}).Do(request)
	if err != nil {
		return "", fmt.Errorf("Error sending request %w", err)
	}
	defer resp.Body.Close()

	value, _ := io.ReadAll(resp.Body)
	fmt.Printf("%s\n", value)

	var response _IPv6PrefixResponseEnvelope

	err = xml.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("Couldn't parse XML %w", err)
	}

	return response.Body.X_AVM_DE_GetIPv6PrefixResponse.NewIPv6Prefix, nil
}
