## 1. Skill Loading Utilities

- [x] 1.1 Implement a `LoadSkills` function in `pkg/agents/connection.go` to read and concatenate all `SKILL.md` content from the `skill/` directory.
- [x] 1.2 Implement a `DiscoverTools` function in `pkg/agents/connection.go` using regex `CALL "([^"]+)"` to extract unique tool names from the loaded skill content.

## 2. Universal Mock Tool

- [x] 2.1 Implement `UniversalMockTool` in `pkg/tools/signaling.go` that takes a generic map of arguments and returns a map with `status`, `token`, and `ue_id` fields.
- [x] 2.2 Ensure `UniversalMockTool` logs the tool name and arguments to `stdout` using `ctx.ToolName()`.

## 3. Dynamic Agent Registration

- [x] 3.1 Update `NewConnectionAgent` in `pkg/agents/connection.go` to call the new discovery and loading utilities during initialization.
- [x] 3.2 Modify the tool registration loop to create a `functiontool` for each discovered name, all pointing to the `UniversalMockTool`.
- [x] 3.3 Append the concatenated skill instructions and explicit state management guidance to the agent's `Instruction` config.

## 4. Verification

- [x] 4.1 Run the `agent-gateway` and provide the intent "initial registration for UE-01" to verify the full `init-registration` sequence is logged.
- [x] 4.2 Provide the intent "connect UE-02 to the network" and verify the `acn` sequence is logged.
- [x] 4.3 Confirm that the LLM correctly extracts the `token` from `Issue_Access_Token_tool` and passes it as an argument to `Create_Subnet_PDUSession_tool`.
