package handlers

import "testing"

func FuzzValidatePromptRequest(f *testing.F) {
	f.Add("claude", "hello")
	f.Add("", "")

	f.Fuzz(func(t *testing.T, backend, prompt string) {
		_ = ValidatePromptRequest(PromptRequest{Backend: backend, Prompt: prompt})
	})
}

func FuzzValidateChainRequest(f *testing.F) {
	f.Add("hello", false, false)
	f.Add("{{session}}", false, false)

	f.Fuzz(func(t *testing.T, prompt string, passSessionID bool, persistSessions bool) {
		_ = ValidateChainRequest(ChainRequest{
			Steps:           []ChainStep{{Backend: "claude", Prompt: prompt}},
			PassSessionID:   passSessionID,
			PersistSessions: persistSessions,
		})
	})
}
