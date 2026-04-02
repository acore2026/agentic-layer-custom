## ADDED Requirements

### Requirement: PCAP-Formatted Telemetry
The Universal Mock Tool SHALL emit `network_pcap` events for both requests and responses, including protocol info, source/destination, and payload details.

#### Scenario: Tool request PCAP emission
- **WHEN** a mock tool is called (e.g., `Subscription_tool`)
- **THEN** it SHALL emit a `network_pcap` event with `direction: request` and the input arguments

#### Scenario: Tool response PCAP emission
- **WHEN** a mock tool returns a result
- **THEN** it SHALL emit a `network_pcap` event with `direction: response` and the output payload
