## Why

The current 6G AI Core Agent Gateway relies on a CLI-based interface, which makes it difficult to visualize complex multi-agent interactions, tool calls, and the internal reasoning process (Chain of Thought). Implementing the ADK Web UI will provide a powerful, browser-based dashboard for developers to monitor the running status, debug agent behavior, and trace signaling sequences in real-time.

## What Changes

- **ADK Web Integration**: Integrate the `google/adk-web` developer UI into the project.
- **Web UI Launcher**: Add a new entry point or update the existing gateway to launch the web-based dashboard alongside or instead of the CLI.
- **Agent Tracing**: Enable detailed event tracing to capture tool calls and "thinking" logs for display in the web interface.
- **Configurable Mode**: Allow users to toggle between CLI and Web UI modes via environment variables or command-line flags.

## Capabilities

### New Capabilities
- `web-dashboard`: A browser-based interface for interacting with and monitoring agents.
- `agent-tracing`: System-wide event tracing for visualizing Chain of Thought and tool executions.

### Modified Capabilities
- `connection-agent`: Enhance telemetry to support rich visualization in the web UI.
- `system-agent`: Enhance telemetry to support routing visualization in the web UI.

## Impact

- `cmd/agent-gateway/main.go`: Updated to support launching the web UI using the ADK launcher.
- `pkg/agents/`: Potential minor updates to ensure all agent events are correctly bubbled up for tracing.
- Dependencies: Addition of `google.golang.org/adk/launcher` and related web UI packages.
