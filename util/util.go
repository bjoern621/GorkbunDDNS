package util

import (
	"encoding/json"
	"io"

	"bjoernblessin.de/gorkbunddns/util/assert"
	"bjoernblessin.de/gorkbunddns/util/logger"
)

func JSONResponseBodyToPrettyByteArray(reader io.Reader) []byte {
	responseBody, err := io.ReadAll(reader)
	assert.IsNil(err, "Can't think of a way this fails here.")

	var responseJsonPretty []byte

	var responseJson map[string]any
	err = json.Unmarshal(responseBody, &responseJson)
	if err != nil {
		logger.Warnf("Porkbun server returned invalid JSON format while validating API keys.")
		responseJsonPretty = responseBody
	} else {
		responseJsonPretty, err = json.MarshalIndent(responseJson, "", "    ")
		if err != nil {
			assert.Never("JSON encoding (marshalling) failed. This should only happen with channel, complex and function values, which we don't use.", err, responseJson)
		}
	}

	return responseJsonPretty
}
