## Context

The backend manages network signaling procedures as "skills" stored in the `skill/` directory. Each skill is a subdirectory containing a `SKILL.md` file. While the `ConnectionAgent` utilizes these for orchestration, the React frontend currently relies on hardcoded definitions, leading to potential desynchronization.

## Goals / Non-Goals

**Goals:**
- Provide a single source of truth for skill definitions via a REST API (`/v1/skills`).
- Automatically reflect additions or changes to `SKILL.md` files in the UI.
- Maintain compatibility with existing CORS and API routing patterns.

**Non-Goals:**
- Implementing skill editing via the API.
- Changing the existing `SKILL.md` file format.

## Decisions

### 1. Skill Discovery Mechanism
**Decision**: Implement a scanning function that traverses the `skill/` directory at runtime.
**Rationale**: This ensures that any new skills added to the filesystem are immediately available without a backend restart or manual registration.
**Alternative**: Static registration in code. Rejected because it duplicates the information already present on the filesystem.

### 2. API Routing
**Decision**: Register `/v1/skills` in the standard HTTP server setup (e.g., in `cmd/agent-gateway/main.go` or a sub-router).
**Rationale**: Consistency with existing API patterns (e.g., `/v1/intents/stream`) and avoiding collision with ADK's default `/api` prefix.

### 3. Data Mapping
**Decision**: Map the directory name to `id` and `name`, and read the `SKILL.md` content into the `definition` field.
**Rationale**: Simple and direct mapping that satisfies the frontend requirement.

## Risks / Trade-offs

- **[Risk] Path Sensitivity**: The backend might fail to find the `skill/` directory if run from a different working directory.
  - **Mitigation**: Use relative paths from the project root or allow configuration via environment variables.
- **[Trade-off] Runtime IO**: Reading files on every request might be slightly slower than in-memory caching.
  - **Mitigation**: The number of skills is small (~10-20), so disk IO is negligible. Caching can be added later if needed.
