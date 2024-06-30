package native

/*
#if defined(__APPLE__) || defined(__linux__)
// https://github.com/wailsapp/wails/pull/2152/files#diff-d4a0fa73df7b0ab971e550f95249e358b634836e925ace96f7400480916ac09e
#include <errno.h>
#include <signal.h>
#include <stdio.h>
#include <string.h>

static void fix_signal(int signum)
{
    struct sigaction st;

    if (sigaction(signum, NULL, &st) < 0) {
        goto fix_signal_error;
    }
    st.sa_flags |= SA_ONSTACK;
    if (sigaction(signum, &st,  NULL) < 0) {
        goto fix_signal_error;
    }
    return;
fix_signal_error:
        fprintf(stderr, "error fixing handler for signal %d, please "
                "report this issue to "
                "https://github.com/pact-foundation/pact-go: %s\n",
                signum, strerror(errno));
}

static void install_signal_handlers()
{
#if defined(SIGCHLD)
    fix_signal(SIGCHLD);
#endif
#if defined(SIGHUP)
    fix_signal(SIGHUP);
#endif
#if defined(SIGINT)
    fix_signal(SIGINT);
#endif
#if defined(SIGQUIT)
    fix_signal(SIGQUIT);
#endif
#if defined(SIGABRT)
    fix_signal(SIGABRT);
#endif
#if defined(SIGFPE)
    fix_signal(SIGFPE);
#endif
#if defined(SIGTERM)
    fix_signal(SIGTERM);
#endif
#if defined(SIGBUS)
    fix_signal(SIGBUS);
#endif
#if defined(SIGSEGV)
    fix_signal(SIGSEGV);
#endif
#if defined(SIGXCPU)
    fix_signal(SIGXCPU);
#endif
#if defined(SIGXFSZ)
    fix_signal(SIGXFSZ);
#endif
}
#else
	static void install_signal_handlers()
	{
	}
#endif
*/
// import "C"
import (
	"fmt"
	"log"

	// "runtime"
	"strings"
)

type Verifier struct {
	handle uintptr
}

func (v *Verifier) Verify(args []string) error {
	log.Println("[DEBUG] executing verifier FFI with args", args)
	// if runtime.GOOS != "windows" {
	// 	C.install_signal_handlers()
	// }
	result := pactffi_verify(strings.Join(args, "\n"))
	/// | Error | Description |
	/// |-------|-------------|
	/// | 1 | The verification process failed, see output for errors |
	/// | 2 | A null pointer was received |
	/// | 3 | The method panicked |
	switch int(result) {
	case 0:
		return nil
	case 1:
		return ErrVerifierFailed
	case 2:
		return ErrInvalidVerifierConfig
	case 3:
		return ErrVerifierPanic
	default:
		return fmt.Errorf("an unknown error (%d) ocurred when verifying the provider (this indicates a defect in the framework)", int(result))
	}
}

// // Version returns the current semver FFI interface version
// func (v *Verifier) Version() string {
// 	return Version()
// }

var (
	// ErrVerifierPanic indicates a panic ocurred when invoking the verifier.
	ErrVerifierPanic = fmt.Errorf("a general panic occured when starting/invoking verifier (this indicates a defect in the framework)")

	// ErrInvalidVerifierConfig indicates an issue configuring the verifier
	ErrInvalidVerifierConfig = fmt.Errorf("configuration for the verifier was invalid and an unknown error occurred (this is most likely a defect in the framework)")

	//ErrVerifierFailed is the standard error if a verification failed (e.g. beacause the pact verification was not successful)
	ErrVerifierFailed = fmt.Errorf("the verifier failed to successfully verify the pacts, this indicates an issue with the provider API")
	//ErrVerifierFailedToRun indicates the verification process was unable to run
	ErrVerifierFailedToRun = fmt.Errorf("the verifier failed to execute (this is most likely a defect in the framework)")
)

func NewVerifier(name string, version string) *Verifier {
	h := pactffi_verifier_new_for_application(name, version)

	return &Verifier{
		handle: h,
	}
}

func (v *Verifier) Shutdown() {
	pactffi_verifier_shutdown(v.handle)
}

func (v *Verifier) SetProviderInfo(name string, scheme string, host string, port uint16, path string) {
	pactffi_verifier_set_provider_info(v.handle, name, scheme, host, port, path)
}

func (v *Verifier) AddTransport(protocol string, port uint16, path string, scheme string) {
	log.Println("[DEBUG] Adding transport with protocol:", protocol, "port:", port, "path:", path, "scheme:", scheme)

	pactffi_verifier_add_provider_transport(v.handle, protocol, port, path, scheme)
}

func (v *Verifier) SetFilterInfo(description string, state string, noState bool) {
	pactffi_verifier_set_filter_info(v.handle, description, state, boolToCInt(noState))
}

func (v *Verifier) SetProviderState(url string, teardown bool, body bool) {
	pactffi_verifier_set_provider_state(v.handle, url, boolToCInt(teardown), boolToCInt(body))
}

func (v *Verifier) SetVerificationOptions(disableSSLVerification bool, requestTimeout int64) {
	// TODO: this returns an int and therefore can error. We should have all of these functions return values??
	pactffi_verifier_set_verification_options(v.handle, boolToCInt(disableSSLVerification), uint64(requestTimeout))
}

func (v *Verifier) SetConsumerFilters(consumers []string) {
	pactffi_verifier_set_consumer_filters(v.handle, stringArrayToCByteArray(consumers), uint16(len(consumers)))
}

func (v *Verifier) AddCustomHeader(name string, value string) {
	pactffi_verifier_add_custom_header(v.handle, name, value)
}

func (v *Verifier) AddFileSource(file string) {
	pactffi_verifier_add_file_source(v.handle, file)
}

func (v *Verifier) AddDirectorySource(directory string) {
	pactffi_verifier_add_directory_source(v.handle, directory)
}

func (v *Verifier) AddURLSource(url string, username string, password string, token string) {
	pactffi_verifier_url_source(v.handle, url, username, password, token)
}

func (v *Verifier) BrokerSourceWithSelectors(url string, username string, password string, token string, enablePending bool, includeWipPactsSince string, providerTags []string, providerBranch string, selectors []string, consumerVersionTags []string) {
	pactffi_verifier_broker_source_with_selectors(v.handle, url, username, password, token, boolToCInt(enablePending), includeWipPactsSince, stringArrayToCByteArray(providerTags), uint16(len(providerTags)), providerBranch, stringArrayToCByteArray(selectors), uint16(len(selectors)), stringArrayToCByteArray(consumerVersionTags), uint16(len(consumerVersionTags)))
}

func (v *Verifier) SetPublishOptions(providerVersion string, buildUrl string, providerTags []string, providerBranch string) {
	pactffi_verifier_set_publish_options(v.handle, providerVersion, buildUrl, stringArrayToCByteArray(providerTags), uint16(len(providerTags)), providerBranch)
}

func (v *Verifier) Execute() error {
	// TODO: Validate
	// if runtime.GOOS != "windows" {
	// 	C.install_signal_handlers()
	// }
	result := pactffi_verifier_execute(v.handle)
	/// | Error | Description |
	/// |-------|-------------|
	/// | 1     | The verification process failed, see output for errors |
	switch int(result) {
	case 0:
		return nil
	case 1:
		return ErrVerifierFailed
	case 2:
		return ErrVerifierFailedToRun
	default:
		return fmt.Errorf("an unknown error (%d) ocurred when verifying the provider (this indicates a defect in the framework)", int(result))
	}
}

func (v *Verifier) SetNoPactsIsError(isError bool) {
	pactffi_verifier_set_no_pacts_is_error(v.handle, boolToCInt(isError))
}

func (v *Verifier) SetColoredOutput(isColoredOutput bool) {
	pactffi_verifier_set_coloured_output(v.handle, boolToCInt(isColoredOutput))
}

func stringArrayToCByteArray(inputs []string) []*byte {
	if len(inputs) == 0 {
		return nil
	}

	output := make([]*byte, len(inputs))

	for i, consumer := range inputs {
		output[i] = CString(consumer)
	}

	return ([]*byte)(output)
}

func boolToCInt(val bool) uint8 {
	if val {
		return uint8(1)
	}
	return uint8(0)
}
