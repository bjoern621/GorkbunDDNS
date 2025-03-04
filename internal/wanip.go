package internal

import (
	"bytes"
	"encoding/xml"
	"fmt"
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
		GetExternalIPAddressResponse struct {
			NewExternalIPAddress string `xml:"NewExternalIPv6Address"`
		} `xml:"X_AVM_DE_GetExternalIPv6Address"`
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
		return "", fmt.Errorf("error sending request %w", err)
	}
	defer resp.Body.Close()

	var response _IPv4ResponseEnvelope
	err = xml.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("couldn't parse XML %w", err)
	}

	return response.Body.GetExternalIPAddressResponse.NewExternalIPAddress, nil
}
