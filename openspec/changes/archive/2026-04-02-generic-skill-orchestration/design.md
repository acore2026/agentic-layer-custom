## Context

The current `ConnectionAgent` has hardcoded instructions and tools. To support a wider range of core network procedures (Registration, PDU Session, ACN), we are moving to a dynamic, skill-driven model. The agent's behavior will be defined by external `SKILL.md` files, and all required tools will be automatically registered with a generic mock implementation for sequence verification.

## Goals / Non-Goals

**Goals:**
- Implement dynamic loading of `SKILL.md` files into the Connection Agent's prompt at runtime.
- Automate tool discovery and registration based on `CALL` patterns in skill files.
- Implement a single `UniversalMockTool` to handle all discovered tool calls with logging and generic success responses.
- Rely solely on LLM-driven reasoning for state passing between sequential tool calls.

**Non-Goals:**
- Implementing real logic for any signaling tool (this is for flow verification only).
- Creating new `SKILL.md` files (utilizing the existing three).
- Modifying the `SystemAgent` routing logic or its own instructions.

## Decisions

- **Dynamic Instruction Bundling**: The `ConnectionAgent` will read all `SKILL.md` files from the `skill/` directory during initialization. The workflows will be concatenated and appended to its system prompt. This ensures the agent is always up-to-date with any new skills without code changes.
- **Regex-Based Tool Discovery**: Unique tool names will be extracted from skills using the regex pattern `CALL "([^"]+)"`. This is chosen for simplicity and robustness given the current deterministic structure of the `SKILL.md` files.
- **Universal Mock Tool**: A single Go function will be registered for every discovered tool. It will use `tool.Context.ToolName()` to identify the call and return a flexible `map[string]any` containing common fields (status, token, id). This satisfies the requirement for a generic, non-hardcoded toolset.

## Risks / Trade-offs

- **[Risk] Prompt Bloat** → Concatenating many skills could hit context limits or degrade LLM performance. *Mitigation*: The current three skills are compact. We will monitor the prompt size during the PoC.
- **[Risk] Ambiguous State Extraction** → Because the mock tool returns generic fields, the LLM might struggle to identify which "token" to use for which tool. *Mitigation*: Add explicit instructions to the prompt telling the LLM to "Reason about which specific field from the previous tool output maps to the next tool's argument."
