package workshop

import (
	"context"
	"testing"

	langfuseplugin "github.com/achetronic/adk-utils-go/plugin/langfuse"

	"agentic-layer-custom/pkg/observability"
)

func TestServiceAgentPluginConfigNilLangfuse(t *testing.T) {
	svc := NewServiceAgent(nil)
	if got := len(svc.pluginConfig().Plugins); got != 0 {
		t.Fatalf("expected empty plugin config, got %d plugins", got)
	}
}

func TestServiceAgentDecorateContext(t *testing.T) {
	svc := NewServiceAgent(&observability.Langfuse{})
	ctx := svc.decorateContext(context.Background(), StartRunRequest{
		RunID:                "run-123",
		ReasoningEnabled:     true,
		CurrentSkillMarkdown: "# existing",
	}, "run-123", "service-run-123", "service-user")

	if got := langfuseplugin.UserIDFromContext(ctx); got != "service-user" {
		t.Fatalf("UserIDFromContext() = %q", got)
	}
	if got := langfuseplugin.TraceNameFromContext(ctx); got != "skill-workshop.start_run" {
		t.Fatalf("TraceNameFromContext() = %q", got)
	}

	tags := langfuseplugin.TagsFromContext(ctx)
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %v", tags)
	}

	metadata := langfuseplugin.TraceMetadataFromContext(ctx)
	if metadata["route"] != "skill_workshop" {
		t.Fatalf("route metadata = %q", metadata["route"])
	}
	if metadata["run_id"] != "run-123" {
		t.Fatalf("run_id metadata = %q", metadata["run_id"])
	}
	if metadata["session_id"] != "service-run-123" {
		t.Fatalf("session_id metadata = %q", metadata["session_id"])
	}
	if metadata["reasoning_enabled"] != "true" {
		t.Fatalf("reasoning_enabled metadata = %q", metadata["reasoning_enabled"])
	}
	if metadata["has_current_skill"] != "true" {
		t.Fatalf("has_current_skill metadata = %q", metadata["has_current_skill"])
	}
}
