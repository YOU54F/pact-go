//go:build provider && plugin
// +build provider,plugin

package grpc

import (
	"fmt"
	"log"

	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/logutils"

	pb "github.com/pact-foundation/pact-go/v2/examples/grpc/routeguide"
	"github.com/pact-foundation/pact-go/v2/examples/grpc/routeguide/server"
	l "github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestGrpcProvider(t *testing.T) {

	go startProvider()
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "TRACE"
	}
	l.SetLogLevel(logutils.LogLevel(logLevel))
	verifier := provider.NewVerifier()

	err := verifier.VerifyProvider(t, provider.VerifyRequest{
		ProviderBaseURL: "http://localhost:8222",
		Transports: []provider.Transport{
			{
				Protocol: "grpc",
				Port:     8222,
			},
		},
		Provider: "grpcprovider",
		PactFiles: []string{
			filepath.ToSlash(fmt.Sprintf("%s/../pacts/grpcconsumer-grpcprovider.json", dir)),
		},
	})

	assert.NoError(t, err)
}

func startProvider() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8222))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRouteGuideServer(grpcServer, server.NewServer())
	grpcServer.Serve(lis)
}
