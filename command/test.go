package command

import (
	"errors"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/pact-foundation/pact-go/v2/installer"
)

// libPathEnvVar names the directory the purego and goffi backends load the
// pact_ffi library from at runtime.
const libPathEnvVar = "PACT_LD_LIBRARY_PATH"

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run go test with the Pact library path configured",
	Long: "Wrapper for 'go test' that sets " + libPathEnvVar + " to the directory the Pact libraries " +
		"were installed to, for builds that load them at runtime (CGO_ENABLED=0, or the pact_purego and pact_goffi build tags)",
	DisableFlagParsing: true,
	Run: func(_ *cobra.Command, args []string) {
		// An explicit environment setting wins over the recorded directory.
		libDir := os.Getenv(libPathEnvVar)
		if libDir == "" {
			libDir = installer.GetConfigLibDir()
		}
		if libDir == "" {
			log.Println("[ERROR] no library directory is configured: run `pact-go install`, or set", libPathEnvVar)
			os.Exit(1)
		}

		log.Printf("[INFO] running go test with %s=%s", libPathEnvVar, libDir)

		goCmd := exec.Command("go", append([]string{"test"}, args...)...) // #nosec G204 -- runs go test with the caller's own arguments
		goCmd.Env = append(os.Environ(), libPathEnvVar+"="+libDir)
		goCmd.Stdout = os.Stdout
		goCmd.Stderr = os.Stderr
		goCmd.Stdin = os.Stdin

		err := goCmd.Run()
		if err == nil {
			return
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}

		log.Printf("[ERROR] failed to run go test: %v", err)
		os.Exit(1)
	},
}

func init() {
	RootCmd.AddCommand(testCmd)
}
