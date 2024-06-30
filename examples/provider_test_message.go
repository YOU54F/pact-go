//go:build provider
// +build provider

// Package main contains a runnable Provider Pact test example.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	// "github.com/hashicorp/logutils"
	// "github.com/pact-foundation/pact-go/v2/log"

	"github.com/pact-foundation/pact-go/v2/message"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/provider"
)

func TestV3MessageProvider(t *testing.T) {
	var dir, _ = os.Getwd()
	var pactDir = fmt.Sprintf("%s/pacts", dir)

	type User struct {
		ID       int    `json:"id" pact:"example=27"`
		Name     string `json:"name" pact:"example=billy"`
		LastName string `json:"lastName" pact:"example=Sampson"`
		Date     string `json:"datetime" pact:"example=2020-01-01'T'08:00:45,format=yyyy-MM-dd'T'HH:mm:ss,generator=datetime"`
	}

	// logLevel := os.Getenv("LOG_LEVEL")
	// if logLevel == "" {
	// 	logLevel = "TRACE"
	// }
	// log.SetLogLevel(logutils.LogLevel(logLevel))
	var user *User

	verifier := provider.NewVerifier()

	// Map test descriptions to message producer (handlers)
	functionMappings := message.Handlers{
		"a user event": func([]models.ProviderState) (message.Body, message.Metadata, error) {
			if user != nil {
				return user, message.Metadata{
					"Content-Type": "application/json",
				}, nil
			} else {
				return models.ProviderStateResponse{
					"message": "not found",
				}, nil, nil
			}
		},
	}

	// Setup any required states for the handlers
	stateMappings := models.StateHandlers{
		"User with id 127 exists": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
			if setup {
				user = &User{
					ID:       127,
					Name:     "Billy",
					Date:     "2020-01-01",
					LastName: "Sampson",
				}
			}

			return models.ProviderStateResponse{"id": user.ID}, nil
		},
	}

	// Verify the Provider with local Pact Files

	if os.Getenv("SKIP_PUBLISH") != "true" {
		verifier.VerifyProvider(t, provider.VerifyRequest{
			StateHandlers:   stateMappings,
			Provider:        "V3MessageProvider",
			ProviderVersion: os.Getenv("APP_SHA"),
			BrokerURL:       os.Getenv("PACT_BROKER_BASE_URL"),
			MessageHandlers: functionMappings,
		})
	} else {
		verifier.VerifyProvider(t, provider.VerifyRequest{
			PactFiles:       []string{filepath.ToSlash(fmt.Sprintf("%s/PactGoV3MessageConsumer-V3MessageProvider.json", pactDir))},
			StateHandlers:   stateMappings,
			Provider:        "V3MessageProvider",
			MessageHandlers: functionMappings,
		})
	}

}
