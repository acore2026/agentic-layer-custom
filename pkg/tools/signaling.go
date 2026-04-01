package tools

import (
	"fmt"
	"time"

	"google.golang.org/adk/tool"
)

// IssueAccessTokenArgs defines the input for the Issue_Access_Token_tool.
type IssueAccessTokenArgs struct {
	UEID string `json:"ue_id" description:"The unique identifier of the User Equipment"`
}

// IssueAccessTokenResult defines the output for the Issue_Access_Token_tool.
type IssueAccessTokenResult struct {
	Token string `json:"token" description:"The issued access token"`
}

// IssueAccessTokenTool is a mock blocking tool for issuing access tokens.
func IssueAccessTokenTool(ctx tool.Context, args *IssueAccessTokenArgs) (*IssueAccessTokenResult, error) {
	fmt.Printf("[Signaling] Issuing access token for UE: %s...\n", args.UEID)
	// Simulate blocking signaling latency
	time.Sleep(2 * time.Second)
	token := fmt.Sprintf("TOKEN-%s-%d", args.UEID, time.Now().Unix())
	fmt.Printf("[Signaling] Token issued: %s\n", token)
	return &IssueAccessTokenResult{Token: token}, nil
}

// CreateSubnetPDUSessionArgs defines the input for the Create_Subnet_PDUSession_tool.
type CreateSubnetPDUSessionArgs struct {
	UEID        string `json:"ue_id" description:"The unique identifier of the User Equipment"`
	AccessToken string `json:"access_token" description:"The access token required for the operation"`
	SubnetID    string `json:"subnet_id" description:"The identifier of the subnet to create the session on"`
}

// CreateSubnetPDUSessionResult defines the output for the Create_Subnet_PDUSession_tool.
type CreateSubnetPDUSessionResult struct {
	Status string `json:"status" description:"The result of the PDU session creation"`
}

// CreateSubnetPDUSessionTool is a mock blocking tool for creating PDU sessions.
func CreateSubnetPDUSessionTool(ctx tool.Context, args *CreateSubnetPDUSessionArgs) (*CreateSubnetPDUSessionResult, error) {
	fmt.Printf("[Signaling] Creating PDU Session for UE: %s on Subnet: %s with Token: %s...\n", args.UEID, args.SubnetID, args.AccessToken)
	// Simulate blocking signaling latency
	time.Sleep(3 * time.Second)
	fmt.Printf("[Signaling] PDU Session created successfully for UE: %s\n", args.UEID)
	return &CreateSubnetPDUSessionResult{Status: "SUCCESS"}, nil
}

// UniversalMockTool is a generic mock tool that logs its call and returns a standard success response.
func UniversalMockTool(ctx tool.Context, toolName string, args map[string]any) (any, error) {
	fmt.Printf("[MOCK] Calling %s with args: %v\n", toolName, args)

	// Small delay to simulate signaling
	time.Sleep(500 * time.Millisecond)

	ueID, _ := args["ue_id"].(string)
	if ueID == "" {
		ueID = "UNKNOWN"
	}

	return map[string]any{
		"status":  "SUCCESS",
		"message": "Simulated response for " + toolName,
		"token":   "TOKEN-" + toolName + "-" + ueID,
		"ue_id":    ueID,
	}, nil
}
