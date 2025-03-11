package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"bjoernblessin.de/gorkbunddns/src/records"
	"bjoernblessin.de/gorkbunddns/src/shared"
	"bjoernblessin.de/gorkbunddns/src/util/assert"
	"bjoernblessin.de/gorkbunddns/src/util/env"
	"bjoernblessin.de/gorkbunddns/src/util/logger"
)

const timeoutSecondsEnvKey string = "TIMEOUT"
const apikeyEnvKey string = "APIKEY"
const secretkeyEnvKey string = "SECRETKEY"
const defaultTimeoutSeconds int = 600

func main() {
	log.Println("Running...")

	apikey, secretkey, timeoutSeconds := validateEnvironment()

	// Program never exits on its own after this point

	runLoop(apikey, secretkey, timeoutSeconds)
}

// validateEnvironment checks environment variables for misconfiguration.
// If one was found, an error message is printed and the program exits.
func validateEnvironment() (apikey string, secretkey string, timeoutSeconds int) {
	apikey = env.ReadNonEmptyRequiredEnv(apikeyEnvKey)
	secretkey = env.ReadNonEmptyRequiredEnv(secretkeyEnvKey)

	testApiKeys(apikey, secretkey)

	timeout, present := env.ReadOptionalEnv(timeoutSecondsEnvKey)
	if present {
		var err error
		timeoutSeconds, err = strconv.Atoi(timeout)
		if err != nil {
			logger.Errorf("Environment variable %s must be a number. Was: %s", timeoutSecondsEnvKey, timeout)
			assert.Never()
		}

		if timeoutSeconds <= 0 {
			logger.Errorf("Environment variable %s must be greater than 0. Was: %d", timeoutSecondsEnvKey, timeoutSeconds)
			assert.Never()
		}
	} else {
		timeoutSeconds = defaultTimeoutSeconds
	}

	env.ReadNonEmptyRequiredEnv(records.DomainsEnvKey)

	IPv4Value := env.ReadValidEnv(records.IPv4EnvKey, []string{"", "true", "false"})
	IPv6Value := env.ReadValidEnv(records.IPv6EnvKey, []string{"", records.IPv6PrefixOnlyValue, records.IPv6HostIPValue, records.IPv6FritzBoxIPValue, "false"})
	if IPv4Value == "false" && (IPv6Value == "" || IPv6Value == "false") {
		logger.Errorf("Both IPv4 and IPv6 updates are disabled. No updates will be performed, so execution is unnecessary.")
		assert.Never()
	}

	return apikey, secretkey, timeoutSeconds
}

// runLoop indefinitely executes the DNS updates.
func runLoop(apikey string, secretkey string, timeoutSeconds int) {
	for {
		records.Update(apikey, secretkey)

		log.Printf("Sleeping for %d seconds.", timeoutSeconds)
		time.Sleep(time.Duration(timeoutSeconds * int(time.Second)))
	}
}

// testApiKeys pings the Porkbun server and validates the provided API keys.
// Stops execution if something fails.
func testApiKeys(apikey string, secretkey string) {
	pingRequest := shared.RequestCredentials{SecretAPIKey: secretkey, APIKey: apikey}
	jsonBody, err := json.Marshal(pingRequest)
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

	if resp.StatusCode != http.StatusOK {
		prettyJSON := _JSONResponseBodyToPrettyByteArray(resp.Body)

		logger.Errorf("Environment variable %s or %s is invalid:\n%s", apikeyEnvKey, secretkeyEnvKey, prettyJSON)
		assert.Never()
	}

	log.Printf("%s and %s successfully validated.", apikeyEnvKey, secretkeyEnvKey)
}

func _JSONResponseBodyToPrettyByteArray(reader io.Reader) []byte {
	responseBody, err := io.ReadAll(reader)
	assert.IsNil(err)

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
