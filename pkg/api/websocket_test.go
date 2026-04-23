package api

import (
	"context"
	"testing"
	"time"

	langfuseplugin "github.com/achetronic/adk-utils-go/plugin/langfuse"
	"google.golang.org/adk/runner"

	"agentic-layer-custom/pkg/observability"
)

func TestPluginConfigNilLangfuse(t *testing.T) {
	cfg := pluginConfig(nil)
	if len(cfg.Plugins) != 0 {
		t.Fatalf("expected empty plugin config, got %d plugins", len(cfg.Plugins))
	}
}

func TestPluginConfigFromLangfuse(t *testing.T) {
	expected := runner.PluginConfig{CloseTimeout: 5 * time.Second}
	lf := &observability.Langfuse{PluginConfig: expected}

	cfg := pluginConfig(lf)
	if cfg.CloseTimeout != expected.CloseTimeout {
		t.Fatalf("unexpected close timeout: got %s want %s", cfg.CloseTimeout, expected.CloseTimeout)
	}
}

func TestDecorateSignalingContext(t *testing.T) {
	lf := &observability.Langfuse{}
	ctx := decorateSignalingContext(context.Background(), lf, IntentRequest{
		Type: "execute_intent",
		Data: struct {
			Intent     string `json:"intent"`
			ScenarioID string `json:"scenarioId"`
		}{
			Intent:     "initial registration for UE-01",
			ScenarioID: "scenario-42",
		},
	}, "session-123")

	if got := langfuseplugin.UserIDFromContext(ctx); got != "web-user" {
		t.Fatalf("UserIDFromContext() = %q", got)
	}
	if got := langfuseplugin.TraceNameFromContext(ctx); got != "agent-gateway.execute_intent" {
		t.Fatalf("TraceNameFromContext() = %q", got)
	}

	metadata := langfuseplugin.TraceMetadataFromContext(ctx)
	if metadata["session_id"] != "session-123" {
		t.Fatalf("TraceMetadataFromContext()[session_id] = %q", metadata["session_id"])
	}
	if metadata["scenario_id"] != "scenario-42" {
		t.Fatalf("TraceMetadataFromContext()[scenario_id] = %q", metadata["scenario_id"])
	}
	if metadata["route"] != "signaling" {
		t.Fatalf("TraceMetadataFromContext()[route] = %q", metadata["route"])
	}

	tags := langfuseplugin.TagsFromContext(ctx)
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %v", tags)
	}
}
