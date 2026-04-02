## Context

The 6G AI Core Agent Gateway currently uses a custom CLI loop and manual agent runners. While functional, it lacks visualization for multi-step signaling and internal LLM reasoning. Google's Agent Development Kit (ADK) provides a built-in `webui` package that can be integrated via the `launcher` to provide a developer dashboard.

## Goals / Non-Goals

**Goals:**
- Integrate `google.golang.org/adk/launcher` and `google.golang.org/adk/launcher/webui`.
- Provide a browser-based dashboard at `http://localhost:8080`.
- Enable real-time tracing of tool calls and Chain of Thought.
- Support both CLI and Web UI modes.

**Non-Goals:**
- Production-ready web UI (this is for development and debugging).
- Modifying core signaling logic or skill definitions.
- Implementing a custom web server (using ADK's built-in one).

## Decisions

- **Launcher Integration**: We will migrate from manual `runner.Run` calls in `main.go` to the official `launcher.Run` pattern. This is required because the `webui` extension hooks into the launcher's lifecycle to capture events.
- **Worker Registration**: Both `SystemAgent` and `ConnectionAgent` will be registered with the launcher as available workers. This allows the user to interact with them individually in the dashboard.
- **Mode Toggle**: A new environment variable `USE_WEB_UI` (default: `false`) will be introduced to control whether to launch the CLI or the Web UI.
- **Trace Capture**: We will ensure that the `kimi` provider and other LLM integrations correctly produce `genai.Part` objects with `Thought: true` to leverage ADK Web's reasoning visualization.

## Risks / Trade-offs

- **[Risk] Launcher Compatibility** → The current `RouteIntent` logic involves a specific routing flow from System to Connection agent. The launcher might represent these as separate workers rather than a single flow.
- **[Mitigation]** → We will register them as separate workers in the dashboard for now, allowing developers to test each one independently.
- **[Risk] Port Conflicts** → The default port `8080` might be in use.
- **[Mitigation]** → We will allow port configuration via environment variables if necessary.
