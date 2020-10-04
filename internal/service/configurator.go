package service

import (
	"encoding/json"
	"strconv"

	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/internet"
)

type Configurator struct{}

func (Configurator) Configuration() (map[string]string, error) {
	return map[string]string{
		// refer to internet.go
		"internet.CanFail":    strconv.FormatBool(internet.CanFail),
		"internet.MinLatency": strconv.Itoa(internet.MinLatency),
		"internet.MaxLatency": strconv.Itoa(internet.MaxLatency),
	}, nil
}

func (Configurator) SetConfiguration(config map[string]string) error {
	for _, err := range []error{
		// shit code, would not recommend. It's only an ok-ish idea here because
		// unmarshalConfig() returns ErrInvalidConfigAtField.
		unmarshalConfig(config, "internet.CanFail", &internet.CanFail),
		unmarshalConfig(config, "internet.MinLatency", &internet.MinLatency),
		unmarshalConfig(config, "internet.MaxLatency", &internet.MaxLatency),
	} {
		if err != nil {
			return err
		}
	}
	return nil
}

func unmarshalConfig(config map[string]string, key string, value interface{}) error {
	if err := json.Unmarshal([]byte(config[key]), value); err != nil {
		return &cchat.ErrInvalidConfigAtField{
			Key: key,
			Err: err,
		}
	}
	return nil
}
