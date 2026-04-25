// say-response reads a Claude Code Stop hook JSON from stdin and speaks
// the last assistant message via a local Kokoro TTS server (localhost:8880).
// Exits silently when CLAUDE_VOICE is not "1".
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	ttsEndpoint = "http://localhost:8880/v1/audio/speech"
	ttsModel    = "kokoro"
	ttsVoice    = "af_sky"
	maxRunes    = 800
)

type stopInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
}

type transcriptLine struct {
	Message struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func main() {
	if os.Getenv("CLAUDE_VOICE") != "1" {
		return
	}

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		return
	}

	var input stopInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return
	}

	path := resolveTranscriptPath(input)
	if path == "" {
		return
	}

	text := lastAssistantText(path)
	if text == "" {
		return
	}

	if err := speakText(truncate(text, maxRunes)); err != nil {
		fmt.Fprintf(os.Stderr, "say-response: %v\n", err)
	}
}

// resolveTranscriptPath returns the JSONL transcript path.
// Prefers the path from the hook JSON; falls back to constructing it from
// session_id and PWD using Claude Code's naming convention.
func resolveTranscriptPath(input stopInput) string {
	if input.TranscriptPath != "" {
		return input.TranscriptPath
	}
	if input.SessionID == "" {
		return ""
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	// Claude Code replaces / and . with - to build the project slug
	slug := strings.NewReplacer("/", "-", ".", "-").Replace(cwd)
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects", slug, input.SessionID+".jsonl")
}

func lastAssistantText(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var last string
	for _, line := range bytes.Split(data, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var entry transcriptLine
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		if entry.Message.Role != "assistant" {
			continue
		}
		if t := extractText(entry.Message.Content); t != "" {
			last = t
		}
	}
	return last
}

// extractText handles both plain-string content and content-block arrays.
func extractText(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	var blocks []contentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return ""
	}
	var parts []string
	for _, b := range blocks {
		if b.Type == "text" && b.Text != "" {
			parts = append(parts, b.Text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max]) + "..."
}

func speakText(text string) error {
	body, err := json.Marshal(map[string]string{
		"model": ttsModel,
		"input": text,
		"voice": ttsVoice,
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(ttsEndpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("TTS server unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("TTS server returned %d", resp.StatusCode)
	}

	audio, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return playAudio(audio)
}

func playAudio(audio []byte) error {
	// macOS: write to temp file and use afplay
	if _, err := exec.LookPath("afplay"); err == nil {
		tmp, err := os.CreateTemp("", "tts-*.wav")
		if err != nil {
			return err
		}
		defer os.Remove(tmp.Name())
		if _, err := tmp.Write(audio); err != nil {
			return err
		}
		tmp.Close()
		return exec.Command("afplay", tmp.Name()).Run()
	}
	// Linux: pipe directly to aplay
	cmd := exec.Command("aplay", "-q", "-")
	cmd.Stdin = bytes.NewReader(audio)
	return cmd.Run()
}
