package mcp

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/WagnerJust/gyroscope/internal/config"
)

var sample = map[string]config.MCPServer{
	"linear": {Command: "npx", Args: []string{"-y", "linear-mcp"}, Env: map[string]string{"KEY": "${LINEAR_KEY}"}},
	"docs":   {URL: "https://x/mcp", Headers: map[string]string{"Authorization": "${DOCS_TOKEN}"}},
}

func TestClaudeFile_ShapeAndEnvPassthrough(t *testing.T) {
	b, err := ClaudeFile(sample)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, b)
	}
	servers, ok := m["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("missing mcpServers wrapper:\n%s", b)
	}
	linear := servers["linear"].(map[string]any)
	if linear["command"] != "npx" {
		t.Errorf("stdio command wrong: %v", linear["command"])
	}
	docs := servers["docs"].(map[string]any)
	if docs["type"] != "http" || docs["url"] != "https://x/mcp" {
		t.Errorf("remote server wrong: %v", docs)
	}
	// ${VAR} passes through verbatim (never expanded, HTML-escaping off).
	if !strings.Contains(string(b), "${LINEAR_KEY}") {
		t.Errorf("env var ref not preserved verbatim:\n%s", b)
	}
}

func TestOpenCodeServers_LocalRemoteShapes(t *testing.T) {
	m := OpenCodeServers(sample)
	linear := m["linear"].(map[string]any)
	if linear["type"] != "local" {
		t.Errorf("stdio should map to local: %v", linear)
	}
	cmd, ok := linear["command"].([]string)
	if !ok || len(cmd) != 3 || cmd[0] != "npx" {
		t.Errorf("command should be a merged array [npx -y linear-mcp]: %v", linear["command"])
	}
	if _, hasEnv := linear["environment"]; !hasEnv {
		t.Errorf("env should map to environment: %v", linear)
	}
	docs := m["docs"].(map[string]any)
	if docs["type"] != "remote" || docs["url"] != "https://x/mcp" {
		t.Errorf("remote shape wrong: %v", docs)
	}
}

func TestClaudeFile_Deterministic(t *testing.T) {
	a, _ := ClaudeFile(sample)
	b, _ := ClaudeFile(sample)
	if string(a) != string(b) {
		t.Error("ClaudeFile must be deterministic (sorted keys) for byte-equal checks")
	}
}

func TestPIConfig_StandardMCPServersShape(t *testing.T) {
	b, err := PIConfig(sample)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, b)
	}
	servers, ok := m["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("PI config must use the mcpServers wrapper:\n%s", b)
	}
	linear := servers["linear"].(map[string]any)
	if linear["command"] != "npx" {
		t.Errorf("stdio command wrong: %v", linear)
	}
	if !strings.Contains(string(b), "${LINEAR_KEY}") {
		t.Errorf("env var ref not preserved:\n%s", b)
	}
}

// The generated extension carries the official SDK (pinned) and the registerTool
// bridge — the shape validated live against pi 0.84.3.
func TestPIExtensionArtifacts(t *testing.T) {
	if !strings.Contains(PIExtensionPackageJSON, "@modelcontextprotocol/sdk") {
		t.Error("package.json must declare the official MCP SDK")
	}
	if strings.Contains(PIExtensionPackageJSON, "pi-mcp-extension") || strings.Contains(PIExtensionIndex, "pi-mcp-extension") {
		t.Error("must not depend on the community wrapper")
	}
	for _, want := range []string{"pi.registerTool(", "StdioClientTransport", "mcp_", "listTools()", "session_shutdown"} {
		if !strings.Contains(PIExtensionIndex, want) {
			t.Errorf("generated extension missing %q", want)
		}
	}
	// No backticks — it must survive as a Go raw-string constant.
	if strings.Contains(PIExtensionIndex, "`") {
		t.Error("extension source must avoid backticks (Go raw-string constant)")
	}
}
