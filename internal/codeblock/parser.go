// Package codeblock provides markdown code block parsing and extraction.
package codeblock

import (
	"path/filepath"
	"regexp"
	"strings"
)

// CodeBlock represents a single code block extracted from markdown.
type CodeBlock struct {
	Language   string `json:"language"`              // From fence (python, bash, etc.)
	Content    string `json:"content"`               // The code itself
	StartLine  int    `json:"start_line"`            // Line number in source
	EndLine    int    `json:"end_line"`              // Ending line number
	FilePath   string `json:"file_path,omitempty"`   // Detected file path (if any)
	IsNew      bool   `json:"is_new,omitempty"`      // Appears to be new file vs modification
	SourcePane string `json:"source_pane,omitempty"` // Pane ID where block was found
}

// Extraction contains all code blocks from a source.
type Extraction struct {
	Source     string      `json:"source"`      // Session/pane identifier
	Blocks     []CodeBlock `json:"blocks"`      // Extracted code blocks
	TotalLines int         `json:"total_lines"` // Total lines in source
}

// fencePattern matches markdown code fences
// Matches: ```language\ncode\n``` or ```\ncode\n```
// Note: Language capture group allows for characters like +, #, -, etc.
var fencePattern = regexp.MustCompile("(?m)^```([^\\r\\n]*)\\r?\\n([\\s\\S]*?)^```")

// Parser extracts code blocks from text.
type Parser struct {
	// Options for parsing
	LanguageFilter []string // Only extract blocks with these languages (empty = all)
}

// NewParser creates a new code block parser.
func NewParser() *Parser {
	return &Parser{}
}

// WithLanguageFilter sets languages to filter for.
func (p *Parser) WithLanguageFilter(langs []string) *Parser {
	p.LanguageFilter = langs
	return p
}

// Parse extracts code blocks from the given text.
func (p *Parser) Parse(text string) []CodeBlock {
	var blocks []CodeBlock

	// Find all matches
	matches := fencePattern.FindAllStringSubmatchIndex(text, -1)

	for _, match := range matches {
		// match[0:2] = full match
		// match[2:4] = language group
		// match[4:6] = content group

		lang := ""
		if match[2] >= 0 && match[3] >= 0 {
			lang = text[match[2]:match[3]]
		}

		// Check language filter
		if len(p.LanguageFilter) > 0 && !p.matchesLanguage(lang) {
			continue
		}

		content := ""
		if match[4] >= 0 && match[5] >= 0 {
			content = text[match[4]:match[5]]
		}

		// Calculate line numbers
		startLine := countLines(text[:match[0]]) + 1
		endLine := startLine + countLines(text[match[0]:match[1]]) - 1

		// Try to detect file path
		filePath, isNew := detectFilePath(content, lang)

		block := CodeBlock{
			Language:  normalizeLanguage(lang),
			Content:   strings.TrimRight(content, "\n\r"),
			StartLine: startLine,
			EndLine:   endLine,
			FilePath:  filePath,
			IsNew:     isNew,
		}

		blocks = append(blocks, block)
	}

	return blocks
}

// languageMap maps aliases to canonical language names
var languageMap = map[string]string{
	"js": "javascript", "jsx": "javascript", "javascript": "javascript",
	"ts": "typescript", "tsx": "typescript", "typescript": "typescript",
	"py": "python", "python": "python",
	"rb": "ruby", "ruby": "ruby",
	"sh": "bash", "shell": "bash", "zsh": "bash", "bash": "bash",
	"yml": "yaml", "yaml": "yaml",
	"dockerfile": "docker", "docker": "docker",
	"md": "markdown", "markdown": "markdown",
	"c++": "cpp", "cpp": "cpp", "cc": "cpp", "cxx": "cpp", "hpp": "cpp",
	"c": "c", "h": "c",
	"cs": "csharp", "csharp": "csharp",
	"rs": "rust", "rust": "rust",
	"kt": "kotlin", "kotlin": "kotlin",
	"go": "go",
	"java": "java",
	"php": "php",
	"html": "html",
	"css": "css",
	"scss": "scss", "sass": "scss",
	"less": "less",
	"sql": "sql",
	"json": "json",
	"xml": "xml",
	"toml": "toml",
	"lua": "lua",
	"pl": "perl", "perl": "perl",
	"r": "r",
	"swift": "swift",
	"scala": "scala",
	"el": "elisp", "lisp": "lisp",
	"clj": "clojure",
	"tf": "terraform",
	"vim": "vim",
}

// matchesLanguage checks if a language matches the filter.
func (p *Parser) matchesLanguage(lang string) bool {
	lang = normalizeLanguage(lang)
	for _, filter := range p.LanguageFilter {
		filter = normalizeLanguage(filter)
		if lang == filter {
			return true
		}
	}
	return false
}

// normalizeLanguage normalizes language names.
func normalizeLanguage(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		return "text"
	}
	if canonical, ok := languageMap[lang]; ok {
		return canonical
	}
	return lang
}

// detectFilePath attempts to detect the file path from code block content.
func detectFilePath(content, lang string) (path string, isNew bool) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return "", false
	}

	firstLine := strings.TrimSpace(lines[0])

	// Check first line for path comment patterns
	// Python/Shell: # path/to/file.py
	if strings.HasPrefix(firstLine, "# ") {
		candidate := strings.TrimPrefix(firstLine, "# ")
		if looksLikeFilePath(candidate) {
			return candidate, false
		}
	}

	// JavaScript/Go/C: // path/to/file.js
	if strings.HasPrefix(firstLine, "// ") {
		candidate := strings.TrimPrefix(firstLine, "// ")
		if looksLikeFilePath(candidate) {
			return candidate, false
		}
	}

	// HTML/XML: <!-- path/to/file.html -->
	if strings.HasPrefix(firstLine, "<!-- ") && strings.HasSuffix(firstLine, " -->") {
		candidate := strings.TrimPrefix(firstLine, "<!-- ")
		candidate = strings.TrimSuffix(candidate, " -->")
		if looksLikeFilePath(candidate) {
			return candidate, false
		}
	}

	// Try to infer from content
	switch normalizeLanguage(lang) {
	case "go":
		// Look for package declaration
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "package ") {
				pkg := strings.TrimPrefix(line, "package ")
				if pkg != "main" {
					return pkg + "/" + pkg + ".go", true
				}
				return "main.go", true
			}
		}
	case "python":
		// Python code without explicit path comment - we can't reliably detect the path
		// Don't set isNew=true without a path as it's misleading
		return "", false
	}

	return "", false
}

// looksLikeFilePath checks if a string looks like a file path.
func looksLikeFilePath(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	// Must have an extension or be a directory path
	ext := filepath.Ext(s)
	if ext != "" {
		return true
	}

	// Check for path separators
	return strings.Contains(s, "/") || strings.Contains(s, "\\")
}

// countLines counts the number of newlines in a string.
func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n")
}

// ExtractFromText is a convenience function to extract code blocks from text.
func ExtractFromText(text string) []CodeBlock {
	return NewParser().Parse(text)
}

// ExtractWithFilter extracts code blocks filtered by language.
func ExtractWithFilter(text string, languages []string) []CodeBlock {
	return NewParser().WithLanguageFilter(languages).Parse(text)
}
