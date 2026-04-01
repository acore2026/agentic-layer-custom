---
name: init-registration
description: UE initial registration procedure.
---
# UE Initial Registration

## Overview

This Skill directs the workflow of a tool chain for UE initial registration in the core network. Follow the defined process to output the specific **Tool Name** and **Fill in Parameters**. Ensure all arguments are instantiated based on the context provided.

## Tool Inventory


## Workflow

Follow the pseudo-code logic below and fill in parameters for each tool to execute tasks.

    # Step 1: Obtain User Intent
    CALL "User_intent"
    
    # Step 2: Perform Authentication
    CALL "Auth_tool"

    # Step 3: Obtain Subscription Information
    CALL "Subscription_tool"
    
    # Step 4: Obtain AM Policies
    CALL "Am_Policy_tool"
        
    # Step 5: Send Registration Acceptance
    CALL "UE_control_tool"

    # Step 6: Obtain and Send UE Policy
    CALL "UE_Policy_tool"

    OUTPUT "DONE"


## Critical Rules

- Do not skip any step.
- All parameters marked as required must be provided. 
- When filling parameters, ensure values for identical keys remain consistent across all steps.
- For tools in the sequence (e.g., if Tool A is followed by Tool B), ensure that **all identical keys** shared between them maintain the exact same values.
- If any tool returns **false** or fails to execute, you must output "ABORT" and exit the workflow.


## Output Format

User_intent(ue_id="SUCI_12345", plmn="PLMN_001", tac="TAC_101", userLoc="Location_001", ue_security_capability="NEA0", ngksi="0x12")
...