package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type PluginSettings struct {
	ClientID string                `json:"client_id"`
	BaseURL  string                `json:"base_url"`
	Secrets  *SecretPluginSettings `json:"-"`
}

type SecretPluginSettings struct {
	ClientSecret string `json:"client_secret"`
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	settings := PluginSettings{}
	if err := json.Unmarshal(source.JSONData, &settings); err != nil {
		return nil, fmt.Errorf("could not unmarshal PluginSettings json: %w", err)
	}
	settings.Secrets = loadSecretPluginSettings(source.DecryptedSecureJSONData)
	return &settings, nil
}

func loadSecretPluginSettings(source map[string]string) *SecretPluginSettings {
	return &SecretPluginSettings{
		ClientSecret: source["client_secret"],
	}
}
