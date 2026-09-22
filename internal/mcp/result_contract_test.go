package mcp

import (
	"context"
	"encoding/json"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestAllBusinessToolsUseTextOnlyContract(t *testing.T) {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "result-contract-test", Version: "test"}, nil)
	RegisterWorkspaceTools(server)
	RegisterQATools(server)
	RegisterProjectTools(server)
	RegisterPinTools(server)
	RegisterRepoWikiTools(server)
	RegisterPreviewTools(server)
	RegisterPreviewLpwTools(server)
	RegisterPagesTools(server)

	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()
	serverSession, err := server.Connect(t.Context(), serverTransport, nil)
	if err != nil {
		t.Fatalf("connect server: %v", err)
	}
	defer serverSession.Close()

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "result-contract-client", Version: "test"}, nil)
	clientSession, err := client.Connect(t.Context(), clientTransport, nil)
	if err != nil {
		t.Fatalf("connect client: %v", err)
	}
	defer clientSession.Close()

	result, err := clientSession.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(result.Tools) != 40 {
		t.Fatalf("tool count = %d, want 40", len(result.Tools))
	}
	for _, tool := range result.Tools {
		if tool.OutputSchema != nil {
			t.Errorf("tool %s must not declare OutputSchema", tool.Name)
		}
	}
}

func TestPreviewLpwSchemaReturnsSingleTextContent(t *testing.T) {
	arguments, err := json.Marshal(map[string]any{"type": "section"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := handlePreviewLpwSchema(t.Context(), &sdkmcp.CallToolRequest{
		Params: &sdkmcp.CallToolParamsRaw{Arguments: arguments},
	})
	if err != nil {
		t.Fatalf("handlePreviewLpwSchema: %v", err)
	}
	if result.IsError {
		t.Fatal("preview_lpw_schema returned an error")
	}
	if result.StructuredContent != nil {
		t.Fatalf("StructuredContent = %#v, want nil", result.StructuredContent)
	}
	if len(result.Content) != 1 {
		t.Fatalf("content block count = %d, want 1", len(result.Content))
	}
	if _, ok := result.Content[0].(*sdkmcp.TextContent); !ok {
		t.Fatalf("content type = %T, want *mcp.TextContent", result.Content[0])
	}
}

func TestLegacyToolFailuresSetIsError(t *testing.T) {
	emptyRequest := func() *sdkmcp.CallToolRequest {
		return &sdkmcp.CallToolRequest{Params: &sdkmcp.CallToolParamsRaw{Arguments: json.RawMessage(`{}`)}}
	}

	SetWorkspaceLogic(nil)
	SetQaLogic(nil)
	SetProjectLogic(nil)
	SetPinLogic(nil)
	SetRepoWikiLogic(nil)

	tests := []struct {
		name    string
		handler sdkmcp.ToolHandler
	}{
		{name: "workspace", handler: handleWorkspaceList},
		{name: "qa", handler: handleQaSessionList},
		{name: "project", handler: handleProjectList},
		{name: "pin", handler: handlePinList},
		{name: "repowiki", handler: handleRepoWikiList},
		{name: "stub", handler: stubToolHandler("missing_tool")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.handler(context.Background(), emptyRequest())
			if err != nil {
				t.Fatalf("handler error: %v", err)
			}
			if !result.IsError {
				t.Fatalf("IsError = false, want true; result = %#v", result)
			}
		})
	}
}
