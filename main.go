package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"bjoernblessin.de/gorkbunddns/util/assert"
	"bjoernblessin.de/gorkbunddns/util/env"
	"bjoernblessin.de/gorkbunddns/util/logger"
)

const timeoutSecondsEnvKey string = "TIMEOUT"
const apikeyEnvKey string = "APIKEY"
const secretkeyEnvKey string = "SECRETKEY"

func main() {
	log.Println("Running...")

	apikey, secretkey, timeoutSeconds := validateEnvironment()

	// Program never exits on its own after this point

	runLoop(apikey, secretkey, timeoutSeconds)
}

// ValidateEnvironment checks all required environment variables and returns them for further use.
func validateEnvironment() (string, string, int) {
	apikey := env.ReadNonEmptyRequiredEnv(apikeyEnvKey)
	secretkey := env.ReadNonEmptyRequiredEnv(secretkeyEnvKey)

	testApiKeys(apikey, secretkey)

	timeout := env.ReadNonEmptyRequiredEnv(timeoutSecondsEnvKey)
	timeoutSeconds, err := strconv.Atoi(timeout)
	if err != nil {
		logger.Errorf("Environment variable %s must be a number. Was: %s", timeoutSecondsEnvKey, timeout)
		assert.Never()
	}

	if timeoutSeconds <= 0 {
		logger.Errorf("Environment variable %s must be greater than 0. Was: %d", timeoutSecondsEnvKey, timeoutSeconds)
		assert.Never()
	}

	return apikey, secretkey, timeoutSeconds
}

// RunLoop indefinitely executes the DNS updates.
func runLoop(apikey string, secretkey string, timeoutSeconds int) {
	for {
		log.Printf("Updating all DNS records.")
		UpdateDNSRecords(apikey, secretkey)

		log.Printf("Sleeping for %d seconds.", timeoutSeconds)
		time.Sleep(time.Duration(timeoutSeconds * int(time.Second)))
	}
}

type PingRequest struct {
	SecretAPIKey string `json:"secretapikey"`
	APIKey       string `json:"apikey"`
}

// TestApiKeys pings the Porkbun server and validates the provided API keys.
// Stops execution if something fails.
func testApiKeys(apikey string, secretkey string) {
	requestBody := PingRequest{SecretAPIKey: secretkey, APIKey: apikey}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logger.Errorf("Cannot json encode environment variables %s and %s, please check them.", apikeyEnvKey, secretkeyEnvKey)
		assert.Never()
	}

	resp, err := http.Post("https://api.porkbun.com/api/json/v3/ping", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		logger.Errorf("Ping to the Porkbun server failed. This may be temporary, please try again later.")
		assert.Never()
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		prettyJSON := JSONResponseBodyToPrettyByteArray(resp.Body)

		logger.Errorf("Environment variable %s and/or %s is invalid:\n%s", apikeyEnvKey, secretkeyEnvKey, prettyJSON)
		assert.Never()
	}

	log.Printf("%s and %s successfully validated.", apikeyEnvKey, secretkeyEnvKey)
}

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
