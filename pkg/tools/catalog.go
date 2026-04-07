package tools

import (
	"sort"
)

// ToolCatalogParam represents a parameter for a tool.
type ToolCatalogParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

// NormalizedTool represents a tool in the catalog with normalized metadata.
type NormalizedTool struct {
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	RequiredParams []string           `json:"required_params"`
	AllParams      []string           `json:"all_params"`
	Parameters     []ToolCatalogParam `json:"parameters"`
}

// NormalizedToolCatalog represents the full catalog of available tools.
type NormalizedToolCatalog struct {
	Tools     []NormalizedTool          `json:"tools"`
	ToolNames []string                  `json:"tool_names"`
	ByName    map[string]NormalizedTool `json:"by_name"`
}

// GetNormalizedToolCatalog returns the catalog of available signaling tools.
func GetNormalizedToolCatalog() NormalizedToolCatalog {
	// For now, we manually define the catalog based on pkg/tools/signaling.go.
	// In the future, this could be dynamically extracted via reflection or a config file.
	rawTools := []NormalizedTool{
		{
			Name:        "Issue_Access_Token_tool",
			Description: "Issues an access token for a specific UE identifier.",
			RequiredParams: []string{"ue_id"},
			AllParams:      []string{"ue_id"},
			Parameters: []ToolCatalogParam{
				{Name: "ue_id", Type: "string", Required: true, Description: "The unique identifier of the User Equipment"},
			},
		},
		{
			Name:        "Create_Subnet_PDUSession_tool",
			Description: "Creates a PDU session for a UE on a specific subnet using an access token.",
			RequiredParams: []string{"ue_id", "access_token", "subnet_id"},
			AllParams:      []string{"ue_id", "access_token", "subnet_id"},
			Parameters: []ToolCatalogParam{
				{Name: "ue_id", Type: "string", Required: true, Description: "The unique identifier of the User Equipment"},
				{Name: "access_token", Type: "string", Required: true, Description: "The access token required for the operation"},
				{Name: "subnet_id", Type: "string", Required: true, Description: "The identifier of the subnet to create the session on"},
			},
		},
		{
			Name:        "Subscription_tool",
			Description: "Checks the subscription status and eligibility for a UE.",
			RequiredParams: []string{"ue_id"},
			AllParams:      []string{"ue_id", "service_type"},
			Parameters: []ToolCatalogParam{
				{Name: "ue_id", Type: "string", Required: true},
				{Name: "service_type", Type: "string", Required: false},
			},
		},
		{
			Name:        "Create_Or_Update_Subnet_Context_tool",
			Description: "Initializes or updates the network context for a specific subnet.",
			RequiredParams: []string{"subnet_id"},
			AllParams:      []string{"subnet_id", "context_data"},
			Parameters: []ToolCatalogParam{
				{Name: "subnet_id", Type: "string", Required: true},
				{Name: "context_data", Type: "object", Required: false},
			},
		},
		{
			Name:        "Validate_Access_Token_tool",
			Description: "Verifies the validity and permissions of an issued access token.",
			RequiredParams: []string{"token"},
			AllParams:      []string{"token"},
			Parameters: []ToolCatalogParam{
				{Name: "token", Type: "string", Required: true},
			},
		},
		{
			Name:        "Auth_tool",
			Description: "Performs UE authentication and key agreement.",
			RequiredParams: []string{"ue_id"},
			AllParams:      []string{"ue_id"},
			Parameters: []ToolCatalogParam{
				{Name: "ue_id", Type: "string", Required: true},
			},
		},
		{
			Name:        "Am_Policy_tool",
			Description: "Retrieves and manages Access and Mobility policies.",
			RequiredParams: []string{"ue_id"},
			AllParams:      []string{"ue_id"},
			Parameters: []ToolCatalogParam{
				{Name: "ue_id", Type: "string", Required: true},
			},
		},
		{
			Name:        "UE_control_tool",
			Description: "Sends control messages and lifecycle commands to the UE.",
			RequiredParams: []string{"ue_id", "command"},
			AllParams:      []string{"ue_id", "command"},
			Parameters: []ToolCatalogParam{
				{Name: "ue_id", Type: "string", Required: true},
				{Name: "command", Type: "string", Required: true},
			},
		},
		{
			Name:        "UE_Policy_tool",
			Description: "Retrieves and delivers UE-specific configuration policies.",
			RequiredParams: []string{"ue_id"},
			AllParams:      []string{"ue_id"},
			Parameters: []ToolCatalogParam{
				{Name: "ue_id", Type: "string", Required: true},
			},
		},
		{
			Name:        "Policy_tool",
			Description: "Generic policy retrieval and evaluation tool.",
			RequiredParams: []string{"context"},
			AllParams:      []string{"context"},
			Parameters: []ToolCatalogParam{
				{Name: "context", Type: "string", Required: true},
			},
		},
		{
			Name:        "UP_Selection_tool",
			Description: "Selects the optimal User Plane Function (UPF) for a session.",
			RequiredParams: []string{"ue_location", "dnn"},
			AllParams:      []string{"ue_location", "dnn"},
			Parameters: []ToolCatalogParam{
				{Name: "ue_location", Type: "string", Required: true},
				{Name: "dnn", Type: "string", Required: true},
			},
		},
		{
			Name:        "UP_Control_Create_tool",
			Description: "Establishes or modifies a user plane session on the UPF.",
			RequiredParams: []string{"upf_id", "session_details"},
			AllParams:      []string{"upf_id", "session_details"},
			Parameters: []ToolCatalogParam{
				{Name: "upf_id", Type: "string", Required: true},
				{Name: "session_details", Type: "object", Required: true},
			},
		},
		{
			Name:        "RAN_Control_tool",
			Description: "Communicates with the Radio Access Network for resource allocation.",
			RequiredParams: []string{"gnb_id", "context"},
			AllParams:      []string{"gnb_id", "context"},
			Parameters: []ToolCatalogParam{
				{Name: "gnb_id", Type: "string", Required: true},
				{Name: "context", Type: "object", Required: true},
			},
		},
		{
			Name:        "User_intent",
			Description: "Captures and normalizes the high-level user request into network parameters.",
			RequiredParams: []string{"ue_id"},
			AllParams:      []string{"ue_id", "plmn", "tac", "userLoc"},
			Parameters: []ToolCatalogParam{
				{Name: "ue_id", Type: "string", Required: true},
				{Name: "plmn", Type: "string", Required: false},
				{Name: "tac", Type: "string", Required: false},
				{Name: "userLoc", Type: "string", Required: false},
			},
		},
	}

	catalog := NormalizedToolCatalog{
		Tools:     rawTools,
		ToolNames: make([]string, 0, len(rawTools)),
		ByName:    make(map[string]NormalizedTool, len(rawTools)),
	}

	for _, t := range rawTools {
		catalog.ToolNames = append(catalog.ToolNames, t.Name)
		catalog.ByName[t.Name] = t
	}

	sort.Strings(catalog.ToolNames)
	return catalog
}
