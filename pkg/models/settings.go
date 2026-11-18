package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type PluginSettings struct {
	URL                    string                `json:"url"`
	Database               string                `json:"database"`
	Username               string                `json:"username"`
	EnableSecureSocksProxy bool                  `json:"enableSecureSocksProxy"`
	Secrets                *SecretPluginSettings `json:"-"`
}

type SecretPluginSettings struct {
	Password string `json:"password"`
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	settings := PluginSettings{}
	if err := json.Unmarshal(source.JSONData, &settings); err != nil {
		return nil, fmt.Errorf("could not unmarshal PluginSettings json: %w", err)
	}
	settings.Secrets = &SecretPluginSettings{
		Password: source.DecryptedSecureJSONData["password"],
	}
	return &settings, nil
}
