package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Generator struct {
	templatesDir string
	baseDir      string
}

type TemplateData struct {
	AssistantName        string
	UserDescription      string
	CustomInstructions   string
	ResponsibilitiesList string
}

func New(templatesDir string) *Generator {
	return &Generator{
		templatesDir: templatesDir,
		baseDir:      "./tmp/customers",
	}
}

func NewWithBaseDir(templatesDir, baseDir string) *Generator {
	return &Generator{
		templatesDir: templatesDir,
		baseDir:      baseDir,
	}
}

func (g *Generator) Generate(customerID, assistantName, customInstructions string) error {
	data := &TemplateData{
		AssistantName:        assistantName,
		UserDescription:      extractUserDescription(customInstructions),
		CustomInstructions:   customInstructions,
		ResponsibilitiesList: extractResponsibilities(customInstructions),
	}

	workspaceDir := filepath.Join(g.baseDir, customerID, ".openclaw", "workspace")
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return fmt.Errorf("create workspace directory: %w", err)
	}

	templates := []struct {
		name     string
		filename string
	}{
		{"personal-assistant/AGENTS.md.tmpl", "AGENTS.md"},
		{"personal-assistant/USER.md.tmpl", "USER.md"},
		{"personal-assistant/SOUL.md.tmpl", "SOUL.md"},
	}

	for _, tmpl := range templates {
		templatePath := filepath.Join(g.templatesDir, tmpl.name)
		outputPath := filepath.Join(workspaceDir, tmpl.filename)

		if err := g.generateFile(templatePath, outputPath, data); err != nil {
			return fmt.Errorf("generate %s: %w", tmpl.filename, err)
		}
	}

	return nil
}

func (g *Generator) generateFile(templatePath, outputPath string, data *TemplateData) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}

func extractUserDescription(instructions string) string {
	lines := strings.Split(instructions, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed
		}
	}
	return "a user seeking AI assistance"
}

func extractResponsibilities(instructions string) string {
	var responsibilities []string
	lines := strings.Split(instructions, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			responsibility := strings.TrimPrefix(line, "-")
			responsibility = strings.TrimPrefix(responsibility, "*")
			responsibility = strings.TrimSpace(responsibility)
			if responsibility != "" {
				responsibilities = append(responsibilities, "- "+responsibility)
			}
		}
	}

	if len(responsibilities) == 0 {
		return "- Provide general assistance as needed"
	}

	return strings.Join(responsibilities, "\n")
}
