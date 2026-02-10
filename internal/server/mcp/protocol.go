package mcp

import "encoding/json"

// ProtocolVersion is the MCP protocol version.
const ProtocolVersion = "2024-11-05"

// Server name and version.
const (
	ServerName    = "clinvoker"
	ServerVersion = "1.0.0"
)

// MCP method names.
const (
	MethodInitialize = "initialize"
	MethodPing       = "ping"
	MethodToolsList  = "tools/list"
	MethodToolsCall  = "tools/call"
)

// InitializeParams represents the params sent by the client during initialization.
type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      Implementation     `json:"clientInfo"`
}

// ClientCapabilities represents client-declared capabilities.
type ClientCapabilities struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

// RootsCapability declares whether the client supports listing roots.
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability declares whether the client supports sampling.
type SamplingCapability struct{}

// Implementation describes a client or server implementation.
type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult is the server's response to initialize.
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      Implementation     `json:"serverInfo"`
}

// ServerCapabilities declares what the server supports.
type ServerCapabilities struct {
	Tools *ToolsCapability `json:"tools,omitempty"`
}

// ToolsCapability declares tool support.
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Tool represents an MCP tool definition.
type Tool struct {
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	InputSchema  json.RawMessage `json:"inputSchema"`
	OutputSchema json.RawMessage `json:"outputSchema"`
}

// ToolsListResult is the response for tools/list.
type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

// ToolCallParams represents the params for tools/call.
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// ToolCallResult is the response for tools/call.
type ToolCallResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock represents a content block in a tool call result.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// TextContent creates a text content block.
func TextContent(text string) ContentBlock {
	return ContentBlock{Type: "text", Text: text}
}

// ErrorResult creates a ToolCallResult indicating an error.
func ErrorResult(message string) *ToolCallResult {
	return &ToolCallResult{
		Content: []ContentBlock{TextContent(message)},
		IsError: true,
	}
}
