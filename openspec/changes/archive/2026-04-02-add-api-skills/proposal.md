## Why

Currently, the "Skill Library" definitions (Markdown/YAML) are hardcoded in the frontend. This creates a synchronization issue where the UI may not match the logic used by the AI agents. Providing a `GET /api/skills` endpoint creates a Single Source of Truth, ensuring the UI always reflects the current signaling procedures defined in the backend.

## What Changes

- **New API Endpoint**: `GET /api/skills` returning JSON list of skills with their raw Markdown definitions.
- **Backend Logic**: Implementation to read skill definitions from the `skill/` directory or internal storage.
- **CORS Support**: Ensuring the new endpoint is accessible from the frontend origin.

## Capabilities

### New Capabilities
- `skills-api`: Provides a standardized interface for clients to discover and retrieve network signaling procedure definitions (skills).

### Modified Capabilities
- None.

## Impact

- **Backend**: New route in `pkg/api` (or similar), file reading logic for `skill/*.md`.
- **Frontend**: Transition from `src/lib/scenarios.ts` hardcoded data to dynamic fetching from `/api/skills`.
- **Maintenance**: Adding new skills only requires adding a Markdown file to the backend repository.
