package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractUserDescription(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"I'm a freelance developer. I need help with proposals.",
			"I'm a freelance developer. I need help with proposals.",
		},
		{
			"\n\n  I'm a designer  \nMore text here",
			"I'm a designer",
		},
		{
			"",
			"a user seeking AI assistance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractUserDescription(tt.input)
			if got != tt.expected {
				t.Errorf("extractUserDescription(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractResponsibilities(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"I need help with:\n- Drafting proposals\n- Research\n- Scheduling",
			"- Drafting proposals\n- Research\n- Scheduling",
		},
		{
			"* Task 1\n* Task 2",
			"- Task 1\n- Task 2",
		},
		{
			"No list items here",
			"- Provide general assistance as needed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractResponsibilities(tt.input)
			if got != tt.expected {
				t.Errorf("extractResponsibilities(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGenerator(t *testing.T) {
	tmpDir := t.TempDir()

	// templatesDir should point to the templates directory (not personal-assistant subdirectory)
	templatesDir := filepath.Join(tmpDir, "templates")
	personalAssistantDir := filepath.Join(templatesDir, "personal-assistant")
	os.MkdirAll(personalAssistantDir, 0755)

	agentsTemplate := `# AGENTS.md
Name: {{.AssistantName}}
Desc: {{.UserDescription}}
Resp: {{.ResponsibilitiesList}}`
	os.WriteFile(filepath.Join(personalAssistantDir, "AGENTS.md.tmpl"), []byte(agentsTemplate), 0644)

	userTemplate := `# USER.md
{{.CustomInstructions}}`
	os.WriteFile(filepath.Join(personalAssistantDir, "USER.md.tmpl"), []byte(userTemplate), 0644)

	soulTemplate := `# SOUL.md
You are {{.AssistantName}}`
	os.WriteFile(filepath.Join(personalAssistantDir, "SOUL.md.tmpl"), []byte(soulTemplate), 0644)

	customersDir := filepath.Join(tmpDir, "customers")
	gen := NewWithBaseDir(templatesDir, customersDir)
	err := gen.Generate("test-customer", "Alex", "I'm a developer. I need help with:\n- Coding\n- Testing")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// The workspace files should be in customersDir/test-customer/.openclaw/workspace
	workspaceDir := filepath.Join(customersDir, "test-customer", ".openclaw", "workspace")

	files := []string{"AGENTS.md", "USER.md", "SOUL.md"}
	for _, file := range files {
		path := filepath.Join(workspaceDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", path)
		}
	}
}
