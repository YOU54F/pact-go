package main

import (
	"testing"

	"github.com/cucumber/godog"
)


func RunFeatureTests(t *testing.T, path string) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{path},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func TestFeaturesV1(t *testing.T) {
	RunFeatureTests(t, "pact-compatibility-suite/features/V1")
}

func TestFeaturesV2(t *testing.T) {
	RunFeatureTests(t, "pact-compatibility-suite/features/V2")
}

func TestFeaturesV3(t *testing.T) {
	RunFeatureTests(t, "pact-compatibility-suite/features/V3")
}

func TestFeaturesV4(t *testing.T) {
	RunFeatureTests(t, "pact-compatibility-suite/features/V4")
}

// func iEat(arg1 int) error {
// 	return godog.ErrPending
// }
func InitializeScenario(ctx *godog.ScenarioContext) {
	// ctx.Step(`^I eat (\d+)$`, iEat)
}
