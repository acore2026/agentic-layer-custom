## 1. Skill Discovery Implementation

- [x] 1.1 Create `pkg/api/skills.go` to handle the skills retrieval logic.
- [x] 1.2 Implement directory scanning logic to find `SKILL.md` files within the `skill/` subdirectories.
- [x] 1.3 Implement a helper to parse metadata (id, name, description) and raw Markdown content from each `SKILL.md`.

## 2. API Route Registration

- [x] 2.1 Integrate the new `Skills` handler into the existing API server in `cmd/agent-gateway/main.go`.
- [x] 2.2 Verify that CORS middleware correctly applies to the new `/api/skills` endpoint.

## 3. Verification

- [x] 3.1 Test the `/v1/skills` endpoint using `curl` or a browser to ensure valid JSON response.
- [x] 3.2 Verify that adding a new directory in `skill/` with a `SKILL.md` is automatically reflected in the API output.
