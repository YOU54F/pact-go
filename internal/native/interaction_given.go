package native

// interactionGiven adds a provider state to the interaction.
func interactionGiven(handle uint32, state string) {
	pactffi_given(handle, state)
}

// interactionGivenWithParams adds a provider state to the interaction,
// once for each of its parameters.
func interactionGivenWithParams(handle uint32, state string, params map[string]any) {
	for k, v := range params {
		pactffi_given_with_param(handle, state, k, stringFromInterface(v))
	}
}
