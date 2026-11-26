package native

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"unsafe"
)

// Verifier is a Go representation of the pact_ffi VerifierHandle, used to
// verify a provider against one or more consumer pacts.
type Verifier struct {
	handle unsafe.Pointer
}

// NewVerifier creates a new provider verifier, identifying the calling
// application as name and version (this module calls it with "pact-go"
// and this module's own version). The caller is responsible for calling
// Shutdown when done with it, to free the underlying resources.
func NewVerifier(name string, version string) *Verifier {
	h := pactffi_verifier_new_for_application(name, version)

	return &Verifier{
		handle: h,
	}
}

// Version returns the current semver FFI interface version.
func (v *Verifier) Version() string {
	return Version()
}

var (
	// ErrVerifierPanic indicates a panic occurred when invoking the verifier.
	ErrVerifierPanic = errors.New("a general panic occurred when starting/invoking verifier (this indicates a defect in the framework)")

	// ErrInvalidVerifierConfig indicates an issue configuring the verifier.
	ErrInvalidVerifierConfig = errors.New("configuration for the verifier was invalid and an unknown error occurred (this is most likely a defect in the framework)")

	// ErrVerifierFailed and ErrVerifierFailedToRun are mutually exclusive: a
	// single Verifier call returns one or the other, never both.
	//
	// ErrVerifierFailed is the standard error if a verification failed (e.g. beacause the pact verification was not successful).
	ErrVerifierFailed = errors.New("the verifier failed to successfully verify the pacts, this indicates an issue with the provider API")
	// ErrVerifierFailedToRun indicates the verification process was unable to run.
	ErrVerifierFailedToRun = errors.New("the verifier failed to execute (this is most likely a defect in the framework)")
)

// Shutdown releases the resources held by this verifier.
func (v *Verifier) Shutdown() {
	pactffi_verifier_shutdown(v.handle)
}

// SetProviderInfo sets the name and network location of the provider
// under verification.
func (v *Verifier) SetProviderInfo(name string, scheme string, host string, port uint16, path string) {
	pactffi_verifier_set_provider_info(v.handle, name, scheme, host, port, path)
}

// AddTransport registers an additional transport (e.g. a message queue
// alongside HTTP) that the provider exposes for verification.
func (v *Verifier) AddTransport(protocol string, port uint16, path string, scheme string) {
	log.Println("[DEBUG] Adding transport with protocol:", protocol, "port:", port, "path:", path, "scheme:", scheme)
	pactffi_verifier_add_provider_transport(v.handle, protocol, port, path, scheme)
}

// SetFilterInfo restricts verification to interactions matching
// description and/or state; noState additionally restricts to
// interactions with no provider state.
func (v *Verifier) SetFilterInfo(description string, state string, noState bool) {
	pactffi_verifier_set_filter_info(v.handle, description, state, boolToCUchar(noState))
}

// SetProviderState sets the URL the verifier calls to set up (and, if
// teardown is set, tear down) provider states before each interaction.
// If body is set, state parameters are sent as a JSON request body rather
// than query parameters.
func (v *Verifier) SetProviderState(url string, teardown bool, body bool) {
	pactffi_verifier_set_provider_state(v.handle, url, boolToCUchar(teardown), boolToCUchar(body))
}

// SetVerificationOptions sets the options used by the verifier when
// calling the provider.
func (v *Verifier) SetVerificationOptions(disableSSLVerification bool, requestTimeout int64) {
	// TODO: this returns an int and therefore can error. We should have all of these functions return values??
	pactffi_verifier_set_verification_options(v.handle, boolToCUchar(disableSSLVerification), uint64(requestTimeout))
}

// SetConsumerFilters is intended to restrict verification to pacts from
// the given consumers.
func (v *Verifier) SetConsumerFilters(consumers []string) {
	// TODO: check if this actually works! It is untested and its effect
	// on the underlying pact_ffi call has not been confirmed.
	cConsumers, freeConsumers := cStringArray(consumers)
	defer freeConsumers()
	pactffi_verifier_set_consumer_filters(v.handle, cConsumers, uint16(len(consumers)))
}

// AddCustomHeader adds a header to be sent with every request made to the
// provider during verification.
func (v *Verifier) AddCustomHeader(name string, value string) {
	pactffi_verifier_add_custom_header(v.handle, name, value)
}

// AddFileSource adds a single Pact file as a source to verify.
func (v *Verifier) AddFileSource(file string) {
	pactffi_verifier_add_file_source(v.handle, file)
}

// AddDirectorySource adds a directory as a source to verify: every pact
// file in it that matches the provider name is verified.
func (v *Verifier) AddDirectorySource(directory string) {
	pactffi_verifier_add_directory_source(v.handle, directory)
}

// AddURLSource adds a URL as a source to verify: the pact file is fetched
// from url, using basic auth if username and password are set, or bearer
// token auth if token is set.
func (v *Verifier) AddURLSource(url string, username string, password string, token string) {
	pactffi_verifier_url_source(v.handle, url, username, password, token)
}

// BrokerSourceWithSelectors adds a Pact Broker as a source to verify,
// fetching every pact matching the provider name and the given consumer
// version selectors (see
// https://docs.pact.io/pact_broker/advanced_topics/consumer_version_selectors/).
// Authentication follows the same rules as AddURLSource.
func (v *Verifier) BrokerSourceWithSelectors(url string, username string, password string, token string, enablePending bool, includeWipPactsSince string, providerTags []string, providerBranch string, selectors []string, consumerVersionTags []string) {
	cProviderTags, freeProviderTags := cStringArray(providerTags)
	defer freeProviderTags()
	cSelectors, freeSelectors := cStringArray(selectors)
	defer freeSelectors()
	cConsumerVersionTags, freeConsumerVersionTags := cStringArray(consumerVersionTags)
	defer freeConsumerVersionTags()

	pactffi_verifier_broker_source_with_selectors(v.handle, url, username, password, token, boolToCUchar(enablePending), includeWipPactsSince, cProviderTags, uint16(len(providerTags)), providerBranch, cSelectors, uint16(len(selectors)), cConsumerVersionTags, uint16(len(consumerVersionTags)))
}

// SetPublishOptions sets the values needed to publish verification
// results back to the Pact Broker: providerVersion is required, the rest
// are optional and pass as "" or nil to omit.
func (v *Verifier) SetPublishOptions(providerVersion string, buildURL string, providerTags []string, providerBranch string) {
	cProviderTags, freeProviderTags := cStringArray(providerTags)
	defer freeProviderTags()

	pactffi_verifier_set_publish_options(v.handle, providerVersion, buildURL, cProviderTags, uint16(len(providerTags)), providerBranch)
}

// Execute runs the verification against every source and option
// configured on v, returning an error if verification failed or could
// not run.
func (v *Verifier) Execute() error {
	// TODO: Validate
	result := pactffi_verifier_execute(v.handle)

	/// | Error | Description |
	/// |-------|-------------|
	/// | 1     | The verification process failed, see output for errors |
	switch int(result) {
	case 0:
		return nil
	case 1:
		// The FFI also returns 1 when a pact could not be loaded or an
		// interaction could not be evaluated, so the result decides.
		return classifyVerifierFailure(v.resultJSON())
	case 2:
		return ErrVerifierFailedToRun
	default:
		return fmt.Errorf("an unknown error (%d) occurred when verifying the provider (this indicates a defect in the framework)", int(result))
	}
}

// SetNoPactsIsError controls whether finding no pacts to verify is
// treated as a verification error.
func (v *Verifier) SetNoPactsIsError(isError bool) {
	pactffi_verifier_set_no_pacts_is_error(v.handle, boolToCUchar(isError))
}

// SetColoredOutput enables or disables ANSI colour codes in the verifier
// output; colour is enabled by default.
func (v *Verifier) SetColoredOutput(isColoredOutput bool) {
	pactffi_verifier_set_coloured_output(v.handle, boolToCUchar(isColoredOutput))
}

// resultJSON returns the verification result of the last Execute as JSON,
// or "" if the FFI could not produce it.
func (v *Verifier) resultJSON() string {
	cJSON := pactffi_verifier_json(v.handle)
	if cJSON == nil {
		return ""
	}
	defer libRustFree(cJSON)

	return goString(cJSON)
}

// verificationResult is the part of the pactffi_verifier_json document
// needed to classify a failed verification.
type verificationResult struct {
	Errors []struct {
		Interaction string `json:"interaction"`
		Mismatch    struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"mismatch"`
	} `json:"errors"`
}

// classifyVerifierFailure decides what a failed verification means, given
// its pactffi_verifier_json result. It returns ErrVerifierFailed only when
// every error is a mismatch, i.e. the provider answered and the answer did
// not satisfy the pact. Any other error (a pact that failed to load, an
// interaction that could not be evaluated) means the verifier produced no
// verdict for it, so it returns ErrVerifierFailedToRun, as it does when
// the result cannot be read.
func classifyVerifierFailure(resultJSON string) error {
	var result verificationResult
	err := json.Unmarshal([]byte(resultJSON), &result)
	if err != nil {
		return fmt.Errorf("%w: could not read the verification result: %w", ErrVerifierFailedToRun, err)
	}

	var notRun []string
	for _, e := range result.Errors {
		if e.Mismatch.Type != "mismatches" {
			notRun = append(notRun, fmt.Sprintf("%s - %s", e.Interaction, e.Mismatch.Message))
		}
	}
	if len(notRun) > 0 {
		return fmt.Errorf("%w: %s", ErrVerifierFailedToRun, strings.Join(notRun, "; "))
	}

	return ErrVerifierFailed
}

func boolToCUchar(val bool) uint8 {
	if val {
		return 1
	}
	return 0
}
