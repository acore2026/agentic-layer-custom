package observability

import (
	"context"
	"testing"

	langfuseplugin "github.com/achetronic/adk-utils-go/plugin/langfuse"
)

func TestNewLangfuseFromEnvDisabledWithoutCredentials(t *testing.T) {
	t.Setenv("LANGFUSE_ENABLED", "")
	t.Setenv("LANGFUSE_PUBLIC_KEY", "")
	t.Setenv("LANGFUSE_SECRET_KEY", "")
	t.Setenv("LANGFUSE_TAGS", "one, two,one")
	t.Setenv("LANGFUSE_ENVIRONMENT", "staging")
	t.Setenv("LANGFUSE_RELEASE", "v1.2.3")

	lf, err := NewLangfuseFromEnv(context.Background(), "agent-gateway")
	if err != nil {
		t.Fatalf("NewLangfuseFromEnv() error = %v", err)
	}
	if lf == nil {
		t.Fatal("NewLangfuseFromEnv() returned nil")
	}
	if got := len(lf.PluginConfig.Plugins); got != 0 {
		t.Fatalf("expected no plugins when credentials are missing, got %d", got)
	}
	if err := lf.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

func TestNewLangfuseFromEnvEnabled(t *testing.T) {
	t.Setenv("LANGFUSE_ENABLED", "true")
	t.Setenv("LANGFUSE_PUBLIC_KEY", "pk-test")
	t.Setenv("LANGFUSE_SECRET_KEY", "sk-test")
	t.Setenv("LANGFUSE_HOST", "http://127.0.0.1:18080")
	t.Setenv("LANGFUSE_INSECURE", "true")

	lf, err := NewLangfuseFromEnv(context.Background(), "agent-gateway")
	if err != nil {
		t.Fatalf("NewLangfuseFromEnv() error = %v", err)
	}
	if got := len(lf.PluginConfig.Plugins); got != 1 {
		t.Fatalf("expected one plugin, got %d", got)
	}
	if err := lf.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

func TestNewLangfuseFromEnvUsesBaseURLAlias(t *testing.T) {
	t.Setenv("LANGFUSE_ENABLED", "true")
	t.Setenv("LANGFUSE_PUBLIC_KEY", "pk-test")
	t.Setenv("LANGFUSE_SECRET_KEY", "sk-test")
	t.Setenv("LANGFUSE_HOST", "")
	t.Setenv("LANGFUSE_BASE_URL", "http://127.0.0.1:18080")
	t.Setenv("LANGFUSE_INSECURE", "true")

	lf, err := NewLangfuseFromEnv(context.Background(), "agent-gateway")
	if err != nil {
		t.Fatalf("NewLangfuseFromEnv() error = %v", err)
	}
	if got := len(lf.PluginConfig.Plugins); got != 1 {
		t.Fatalf("expected one plugin, got %d", got)
	}
	if err := lf.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

func TestLangfuseDecorateContext(t *testing.T) {
	lf := &Langfuse{
		tags:        []string{"repo-tag", "agent-gateway"},
		environment: "staging",
		release:     "v1.2.3",
	}

	ctx := lf.DecorateContext(context.Background(), TraceOptions{
		TraceName: "agent-gateway.execute_intent",
		UserID:    "web-user",
		Tags:      []string{"signaling", "agent-gateway"},
		Metadata: map[string]string{
			"route":       "signaling",
			"scenario_id": "demo-1",
			"empty":       "   ",
		},
	})

	if got := langfuseplugin.UserIDFromContext(ctx); got != "web-user" {
		t.Fatalf("UserIDFromContext() = %q", got)
	}
	if got := langfuseplugin.TraceNameFromContext(ctx); got != "agent-gateway.execute_intent" {
		t.Fatalf("TraceNameFromContext() = %q", got)
	}

	tags := langfuseplugin.TagsFromContext(ctx)
	if len(tags) != 3 {
		t.Fatalf("expected 3 deduplicated tags, got %v", tags)
	}

	metadata := langfuseplugin.TraceMetadataFromContext(ctx)
	if metadata["route"] != "signaling" || metadata["scenario_id"] != "demo-1" {
		t.Fatalf("TraceMetadataFromContext() = %#v", metadata)
	}
	if _, ok := metadata["empty"]; ok {
		t.Fatalf("expected empty metadata value to be dropped, got %#v", metadata)
	}

	if got := langfuseplugin.EnvironmentFromContext(ctx); got != "staging" {
		t.Fatalf("EnvironmentFromContext() = %q", got)
	}
	if got := langfuseplugin.ReleaseFromContext(ctx); got != "v1.2.3" {
		t.Fatalf("ReleaseFromContext() = %q", got)
	}
}

func TestEnvBoolRejectsInvalidValue(t *testing.T) {
	t.Setenv("LANGFUSE_ENABLED", "definitely-not-bool")
	_, err := NewLangfuseFromEnv(context.Background(), "agent-gateway")
	if err == nil {
		t.Fatal("expected error for invalid LANGFUSE_ENABLED")
	}
}
