package shared

type RequestCredentials struct {
	SecretAPIKey string `json:"secretapikey"`
	APIKey       string `json:"apikey"`
}
