package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/fastclaw-ai/fastclaw/internal/config"
	"github.com/fastclaw-ai/fastclaw/internal/provider"
)

type llmAttempt struct {
	model    string
	provider provider.Provider
}

func normalizeModelRefs(primary string, fallbacks []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, 1+len(fallbacks))
	add := func(ref string) {
		ref = strings.TrimSpace(ref)
		if ref == "" || seen[ref] {
			return
		}
		seen[ref] = true
		out = append(out, ref)
	}
	add(primary)
	for _, ref := range fallbacks {
		add(ref)
	}
	return out
}

func llmFallbackEligible(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	switch {
	case strings.Contains(s, "allocationquota.freetieronly"),
		strings.Contains(s, "insufficient_quota"),
		strings.Contains(s, "quota exceeded"),
		strings.Contains(s, "exceeded your current quota"),
		strings.Contains(s, "api error 429"),
		strings.Contains(s, "too many requests"):
		return true
	default:
		return false
	}
}

func cloneProviderConfigs(in map[string]config.ProviderConfig) map[string]config.ProviderConfig {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]config.ProviderConfig, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func (a *Agent) providerForModel(modelRef string) provider.Provider {
	modelRef = strings.TrimSpace(modelRef)
	if modelRef == "" {
		return nil
	}
	providerKey, _ := provider.SplitProviderModel(modelRef)
	if providerKey == "" {
		return a.provider
	}
	if cfg, ok := a.providerConfigs[providerKey]; ok && strings.TrimSpace(cfg.APIKey) != "" {
		return provider.NewProvider(cfg.APIKey, cfg.APIBase, cfg.APIType)
	}
	currentKey, _ := provider.SplitProviderModel(a.model)
	if providerKey == currentKey {
		return a.provider
	}
	return nil
}

func (a *Agent) llmAttempts() []llmAttempt {
	models := normalizeModelRefs(a.model, a.modelFallbacks)
	out := make([]llmAttempt, 0, len(models))
	for _, modelRef := range models {
		if prov := a.providerForModel(modelRef); prov != nil {
			out = append(out, llmAttempt{model: modelRef, provider: prov})
		}
	}
	return out
}

func (a *Agent) chatWithFallback(ctx context.Context, messages []provider.Message, tools []provider.Tool) (*provider.Response, string, error) {
	attempts := a.llmAttempts()
	if len(attempts) == 0 {
		return nil, "", errors.New("no usable LLM provider for configured model chain")
	}
	var errs []error
	for i, attempt := range attempts {
		resp, err := attempt.provider.Chat(ctx, messages, tools, attempt.model, a.maxTokens, a.temperature)
		if err == nil {
			if i > 0 {
				slog.Warn("llm fallback succeeded",
					"agent", a.name,
					"selected_model", a.model,
					"fallback_model", attempt.model,
					"attempt", i+1,
				)
			}
			return resp, attempt.model, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", attempt.model, err))
		if i < len(attempts)-1 && llmFallbackEligible(err) {
			slog.Warn("llm fallback triggered",
				"agent", a.name,
				"failed_model", attempt.model,
				"next_model", attempts[i+1].model,
				"error", err,
			)
			continue
		}
		return nil, "", errors.Join(errs...)
	}
	return nil, "", errors.Join(errs...)
}

func (a *Agent) chatStreamWithFallback(ctx context.Context, messages []provider.Message, tools []provider.Tool) (*provider.StreamReader, string, error) {
	attempts := a.llmAttempts()
	if len(attempts) == 0 {
		return nil, "", errors.New("no usable LLM provider for configured model chain")
	}
	var errs []error
	for i, attempt := range attempts {
		sr, err := attempt.provider.ChatStream(ctx, messages, tools, attempt.model, a.maxTokens, a.temperature)
		if err == nil {
			if i > 0 {
				slog.Warn("llm stream fallback succeeded",
					"agent", a.name,
					"selected_model", a.model,
					"fallback_model", attempt.model,
					"attempt", i+1,
				)
			}
			return sr, attempt.model, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", attempt.model, err))
		if i < len(attempts)-1 && llmFallbackEligible(err) {
			slog.Warn("llm stream fallback triggered",
				"agent", a.name,
				"failed_model", attempt.model,
				"next_model", attempts[i+1].model,
				"error", err,
			)
			continue
		}
		return nil, "", errors.Join(errs...)
	}
	return nil, "", errors.Join(errs...)
}
