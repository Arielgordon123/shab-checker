package config

import (
	"encoding/json"
	"os"
)

type PreDefinedCells struct {
	// Sheet      string   `json:"sheet"`
	CellRange  string `json:"cellRange"`
	TitleRange string `json:"titleRange"`
	TimeRange  string `json:"timeRange"`
}

type Config struct {
	SpreadsheetIDs struct {
		Sheet1 string `json:"sheet1"`
		Sheet2 string `json:"sheet2"`
	} `json:"spreadsheetIDs"`
	CredentialsFile string `json:"credentialsFile"`

	PreDefinedCells []PreDefinedCells `json:"preDefinedCells"`
	TgServiceURL    string            `json:"tgServiceURL"`
}

type Credentials struct {
	Type                    string `json:"type"`
	ProjectId               string `json:"project_id"`
	PrivateKeyId            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientId                string `json:"client_id"`
	AuthUri                 string `json:"auth_uri"`
	TokenUri                string `json:"token_uri"`
	AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
	ClientX509CertUrl       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

func LoadConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func LoadCredentials(path string) (Credentials, error) {

	file, err := os.Open(path)
	if err != nil {
		return Credentials{}, err
	}
	defer file.Close()

	var credentials Credentials
	if err := json.NewDecoder(file).Decode(&credentials); err != nil {
		return Credentials{}, err
	}

	return credentials, nil
}
