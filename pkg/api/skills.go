package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Skill represents a network signaling procedure definition.
type Skill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Definition  string `json:"definition"`
}

// SkillsResponse is the top-level JSON structure for the skills API.
type SkillsResponse struct {
	Skills []Skill `json:"skills"`
}

// HandleSkills serves the list of skills discovered from the filesystem.
func HandleSkills(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	skillDir := "skill" // Base directory for skills
	skills, err := loadSkillsFromDir(skillDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load skills: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(SkillsResponse{Skills: skills}); err != nil {
		fmt.Printf("[API] Failed to encode skills response: %v\n", err)
	}
}

func loadSkillsFromDir(skillDir string) ([]Skill, error) {
	var skills []Skill
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			skillID := entry.Name()
			skillFile := filepath.Join(skillDir, skillID, "SKILL.md")
			if _, err := os.Stat(skillFile); err == nil {
				content, err := os.ReadFile(skillFile)
				if err != nil {
					continue
				}

				skill := parseSkillMarkdown(skillID, string(content))
				skills = append(skills, skill)
			}
		}
	}
	return skills, nil
}

func parseSkillMarkdown(id, content string) Skill {
	skill := Skill{
		ID:         id,
		Definition: content,
	}

	// Simple YAML frontmatter parser
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			frontmatter := parts[1]
			lines := strings.Split(frontmatter, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "name:") {
					skill.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				} else if strings.HasPrefix(line, "description:") {
					skill.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				}
			}
		}
	}

	if skill.Name == "" {
		skill.Name = id
	}

	return skill
}
