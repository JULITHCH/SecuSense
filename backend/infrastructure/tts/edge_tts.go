package tts

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/secusense/backend/config"
)

// Voice represents an available TTS voice
type Voice struct {
	Name     string `json:"name"`
	Language string `json:"language"`
	Gender   string `json:"gender"`
}

// Client handles text-to-speech generation using Edge TTS
type Client struct {
	voice     string
	outputDir string
}

// NewClient creates a new Edge TTS client
func NewClient(cfg config.TTSConfig) *Client {
	voice := cfg.Voice
	if voice == "" {
		voice = "de-DE-KatjaNeural" // Default German female voice
	}

	outputDir := cfg.OutputDir
	if outputDir == "" {
		outputDir = "./generated/audio"
	}

	// Ensure output directory exists
	os.MkdirAll(outputDir, 0755)

	return &Client{
		voice:     voice,
		outputDir: outputDir,
	}
}

// GenerateAudio generates an audio file from text using Edge TTS
// Returns the relative URL path to the generated audio file
func (c *Client) GenerateAudio(ctx context.Context, text string, language string) (string, error) {
	// Select appropriate voice based on language
	voice := c.getVoiceForLanguage(language)

	// Generate unique filename
	filename := fmt.Sprintf("%s.mp3", uuid.New().String())
	outputPath := filepath.Join(c.outputDir, filename)

	// Build edge-tts command
	cmd := exec.CommandContext(ctx, "edge-tts",
		"--voice", voice,
		"--text", text,
		"--write-media", outputPath,
	)

	// Run command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("edge-tts failed: %w, output: %s", err, string(output))
	}

	// Return the relative URL path
	return fmt.Sprintf("/api/v1/audio/%s", filename), nil
}

// GenerateAudioBatch generates multiple audio files for a list of texts
func (c *Client) GenerateAudioBatch(ctx context.Context, texts []string, language string) ([]string, error) {
	urls := make([]string, len(texts))

	for i, text := range texts {
		url, err := c.GenerateAudio(ctx, text, language)
		if err != nil {
			return nil, fmt.Errorf("failed to generate audio for text %d: %w", i, err)
		}
		urls[i] = url
	}

	return urls, nil
}

// getVoiceForLanguage returns the appropriate voice for the given language
func (c *Client) getVoiceForLanguage(language string) string {
	voices := map[string]string{
		"de": "de-DE-KatjaNeural",
		"en": "en-US-JennyNeural",
		"fr": "fr-FR-DeniseNeural",
		"es": "es-ES-ElviraNeural",
		"it": "it-IT-ElsaNeural",
		"pt": "pt-BR-FranciscaNeural",
	}

	// Extract base language code (e.g., "de" from "de-DE")
	langCode := strings.Split(language, "-")[0]
	if voice, ok := voices[langCode]; ok {
		return voice
	}

	return c.voice // Fall back to configured default
}

// GetAvailableVoices returns a list of commonly used voices
func (c *Client) GetAvailableVoices() []Voice {
	return []Voice{
		{Name: "de-DE-KatjaNeural", Language: "German", Gender: "Female"},
		{Name: "de-DE-ConradNeural", Language: "German", Gender: "Male"},
		{Name: "en-US-JennyNeural", Language: "English (US)", Gender: "Female"},
		{Name: "en-US-GuyNeural", Language: "English (US)", Gender: "Male"},
		{Name: "en-GB-SoniaNeural", Language: "English (UK)", Gender: "Female"},
		{Name: "fr-FR-DeniseNeural", Language: "French", Gender: "Female"},
		{Name: "es-ES-ElviraNeural", Language: "Spanish", Gender: "Female"},
		{Name: "it-IT-ElsaNeural", Language: "Italian", Gender: "Female"},
		{Name: "pt-BR-FranciscaNeural", Language: "Portuguese (BR)", Gender: "Female"},
	}
}

// IsAvailable checks if edge-tts is installed and available
func (c *Client) IsAvailable() bool {
	_, err := exec.LookPath("edge-tts")
	return err == nil
}

// GetOutputDir returns the configured output directory
func (c *Client) GetOutputDir() string {
	return c.outputDir
}
