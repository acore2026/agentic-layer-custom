## 1. Dependency Management

- [x] 1.1 Add `google.golang.org/adk/cmd/launcher/universal` and `google.golang.org/adk/cmd/launcher/web/webui` to `go.mod`: `go get google.golang.org/adk/cmd/launcher/universal google.golang.org/adk/cmd/launcher/web/webui`

## 2. Main Entry Point Update

- [x] 2.1 Update `cmd/agent-gateway/main.go` to support the `USE_WEB_UI` environment variable.
- [x] 2.2 Refactor `main.go` to use `launcher.Run` when `USE_WEB_UI` is true.
- [x] 2.3 Register both `SystemAgent` and `ConnectionAgent` as workers in the launcher configuration.

## 3. Telemetry and Tracing

- [x] 3.1 Verify that `pkg/model/kimi/kimi.go` correctly propagates reasoning content using `genai.Part.Thought` for UI visualization.
- [x] 3.2 Ensure `UniversalMockTool` calls produce events that are correctly captured by the launcher's tracer.

## 4. Unified Gateway Orchestrator

- [x] 4.1 Implement `NewGatewayAgent` in `pkg/agents/gateway.go` that encapsulates the logic from `RouteIntent`.
- [x] 4.2 Update `cmd/agent-gateway/main.go` to use the `GatewayAgent` as the primary worker in the Web UI.

## 5. Verification

- [x] 4.1 Launch the gateway with `USE_WEB_UI=true` and verify the dashboard is accessible at `http://localhost:8080`.
- [x] 4.2 Select the `ConnectionAgent` in the UI and trigger an "initial registration" intent.
- [x] 4.3 Verify that the trace log shows each signaling step and the dimmed "thinking" process correctly.
- [x] 4.4 Verify that the CLI mode still works as expected when `USE_WEB_UI=false`.
