package observability

import (
	"context"
	"os"
	"strconv"
	"strings"

	langfuseplugin "github.com/achetronic/adk-utils-go/plugin/langfuse"
	"google.golang.org/adk/runner"
)

// Langfuse holds the shared plugin config and request-level defaults.
type Langfuse struct {
	PluginConfig runner.PluginConfig
	Shutdown     func(context.Context) error
	tags         []string
	environment  string
	release      string
}

// TraceOptions defines per-request Langfuse metadata.
type TraceOptions struct {
	TraceName   string
	UserID      string
	Tags        []string
	Metadata    map[string]string
	Environment string
	Release     string
}

// NewLangfuseFromEnv creates Langfuse plugin config from environment variables.
func NewLangfuseFromEnv(ctx context.Context, serviceName string) (*Langfuse, error) {
	enabled, err := envBool("LANGFUSE_ENABLED", true)
	if err != nil {
		return nil, err
	}

	lf := &Langfuse{
		Shutdown:    func(context.Context) error { return nil },
		tags:        splitCSV(os.Getenv("LANGFUSE_TAGS")),
		environment: os.Getenv("LANGFUSE_ENVIRONMENT"),
		release:     os.Getenv("LANGFUSE_RELEASE"),
	}
	if !enabled {
		return lf, nil
	}

	cfg := &langfuseplugin.Config{
		PublicKey:   os.Getenv("LANGFUSE_PUBLIC_KEY"),
		SecretKey:   os.Getenv("LANGFUSE_SECRET_KEY"),
		Host:        langfuseHost(),
		Environment: lf.environment,
		Release:     lf.release,
		ServiceName: serviceName,
	}
	if cfg.Host != "" {
		insecure, err := envBool("LANGFUSE_INSECURE", false)
		if err != nil {
			return nil, err
		}
		cfg.Insecure = insecure
	}

	if !cfg.IsEnabled() {
		return lf, nil
	}

	pluginCfg, shutdown, err := langfuseplugin.Setup(cfg)
	if err != nil {
		return nil, err
	}
	lf.PluginConfig = pluginCfg
	lf.Shutdown = shutdown
	return lf, nil
}

// DecorateContext applies request-scoped Langfuse metadata.
func (l *Langfuse) DecorateContext(ctx context.Context, opts TraceOptions) context.Context {
	if l == nil {
		return ctx
	}

	if userID := strings.TrimSpace(opts.UserID); userID != "" {
		ctx = langfuseplugin.WithUserID(ctx, userID)
	}

	tags := append(append([]string{}, l.tags...), opts.Tags...)
	if len(tags) > 0 {
		ctx = langfuseplugin.WithTags(ctx, dedupe(tags))
	}

	if metadata := cleanMetadata(opts.Metadata); len(metadata) > 0 {
		ctx = langfuseplugin.WithTraceMetadata(ctx, metadata)
	}

	environment := strings.TrimSpace(opts.Environment)
	if environment == "" {
		environment = l.environment
	}
	if environment != "" {
		ctx = langfuseplugin.WithEnvironment(ctx, environment)
	}

	release := strings.TrimSpace(opts.Release)
	if release == "" {
		release = l.release
	}
	if release != "" {
		ctx = langfuseplugin.WithRelease(ctx, release)
	}

	if traceName := strings.TrimSpace(opts.TraceName); traceName != "" {
		ctx = langfuseplugin.WithTraceName(ctx, traceName)
	}

	return ctx
}

func envBool(name string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}
	return strconv.ParseBool(raw)
}

func langfuseHost() string {
	if host := strings.TrimSpace(os.Getenv("LANGFUSE_HOST")); host != "" {
		return host
	}
	return strings.TrimSpace(os.Getenv("LANGFUSE_BASE_URL"))
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return dedupe(out)
}

func dedupe(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func cleanMetadata(metadata map[string]string) map[string]string {
	if len(metadata) == 0 {
		return nil
	}

	out := make(map[string]string, len(metadata))
	for key, value := range metadata {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
