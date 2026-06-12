package agent

import (
	"errors"
	"testing"
)

func TestNormalizeModelRefs(t *testing.T) {
	got := normalizeModelRefs("openai/a", []string{
		"openai/a",
		" openai/b ",
		"",
		"openai/c",
		"openai/b",
	})
	want := []string{"openai/a", "openai/b", "openai/c"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q (full=%v)", i, got[i], want[i], got)
		}
	}
}

func TestLLMFallbackEligible(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "free tier only", err: errors.New("API error 403: AllocationQuota.FreeTierOnly."), want: true},
		{name: "insufficient quota", err: errors.New("429 insufficient_quota"), want: true},
		{name: "too many requests", err: errors.New("API error 429: too many requests"), want: true},
		{name: "bad request", err: errors.New("API error 400: invalid model"), want: false},
		{name: "nil", err: nil, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := llmFallbackEligible(tc.err); got != tc.want {
				t.Fatalf("llmFallbackEligible(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
