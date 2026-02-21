package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"github.com/SkalaSkalolaz/llmclient"
)

const (
	defaultTimeout     = 5 * time.Minute
	audioFileMode      = 0644
	usageTemplate      = "Usage: %s {provider} {model} {api_key} {output_file} {text|@file|-}\nRun '%s -h' for more information.\n"
	envPollinationsKey = "POLLINATIONS_API_KEY"
)

type Config struct {
	Provider   string
	Model      string
	APIKey     string
	OutputPath string
	Text       string
}

func main() {
	ctx, cancel := signalAwareContext(context.Background())
	defer cancel()

	if err := run(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func signalAwareContext(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	return ctx, cancel
}

func run(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf(usageTemplate, args[0], args[0])
	}

	provider := args[1]

	if provider == "-h" || provider == "--help" {
		return showHelp(ctx, args[0])
	}

	if len(args) >= 3 && args[2] == "models" {
		return showProviderModels(ctx, args[0], provider)
	}

	cfg, err := parseArgs(args)
	if err != nil {
		return err
	}

	return generateAudio(ctx, cfg)
}

func parseArgs(args []string) (*Config, error) {
	if len(args) < 6 {
		return nil, fmt.Errorf(usageTemplate, args[0], args[0])
	}

	text := args[5]
	if text == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}
		text = string(data)
	} else if len(text) > 0 && text[0] == '@' {
		filePath := text[1:]
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %q: %w", filePath, err)
		}
		text = string(content)
	}

	if text == "" {
		return nil, errors.New("text cannot be empty")
	}

	apiKey := args[3]
	if apiKey == "" {
		apiKey = os.Getenv(envPollinationsKey)
	}

	cfg := &Config{
		Provider:   args[1],
		Model:      args[2],
		APIKey:     apiKey,
		OutputPath: args[4],
		Text:       text,
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Provider == "" {
		return errors.New("provider cannot be empty")
	}
	if cfg.Model == "" {
		return errors.New("model cannot be empty")
	}
	if cfg.OutputPath == "" {
		return errors.New("output file path cannot be empty")
	}
	if cfg.Text == "" {
		return errors.New("text cannot be empty")
	}
	return nil
}

func generateAudio(ctx context.Context, cfg *Config) error {
	client := llmclient.NewClient(llmclient.WithTimeout(defaultTimeout))

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	audioData, err := client.GenerateAudio(ctx, &llmclient.AudioRequest{
		Provider: cfg.Provider,
		Model:    cfg.Model,
		APIKey:   cfg.APIKey,
		Prompt:   cfg.Text,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return errors.New("operation cancelled by user")
		}
		return fmt.Errorf("failed to generate audio: %s", parseAPIError(err.Error()))
	}

	if len(audioData.Data) == 0 {
		return errors.New("received empty audio data")
	}

	if err := saveAudioFile(cfg.OutputPath, audioData.Data); err != nil {
		return fmt.Errorf("failed to save audio file: %w", err)
	}

	fmt.Printf("Audio saved successfully: %s (%d bytes)\n", cfg.OutputPath, len(audioData.Data))
	return nil
}

type apiErrorDetail struct {
	Message string `json:"message"`
}

type apiErrorInner struct {
	Message string          `json:"message"`
	Detail  *apiErrorDetail `json:"detail"`
}

type apiErrorResponse struct {
	Error      *apiErrorInner `json:"error"`
	Message    string         `json:"message"`
	Success    bool           `json:"success"`
	StatusCode int            `json:"status"`
}

func parseAPIError(errMsg string) string {
	jsonPattern := regexp.MustCompile(`api error \d+: (.+)`)
	matches := jsonPattern.FindStringSubmatch(errMsg)
	if len(matches) < 2 {
		return errMsg
	}

	jsonStr := matches[1]

	var outerResp apiErrorResponse
	if err := json.Unmarshal([]byte(jsonStr), &outerResp); err == nil {
		if outerResp.Error != nil && outerResp.Error.Message != "" {
			return outerResp.Error.Message
		}
		if outerResp.Message != "" {
			return outerResp.Message
		}
	}

	var innerResp struct {
		Error struct {
			Message string `json:"message"`
			Detail  struct {
				Message string `json:"message"`
			} `json:"detail"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &innerResp); err == nil {
		if innerResp.Error.Detail.Message != "" {
			return innerResp.Error.Detail.Message
		}
		if innerResp.Error.Message != "" {
			return innerResp.Error.Message
		}
	}

	return jsonStr
}

func saveAudioFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, audioFileMode)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

func showHelp(ctx context.Context, programName string) error {
	fmt.Printf("Text-to-Speech Generator using llmclient\n\n")
	fmt.Printf("Usage:\n")
	fmt.Printf("  %s {provider} {model} {api_key} {output_file} {text|@file|-}\n", programName)
	fmt.Printf("  %s {provider} models\n\n", programName)
	fmt.Println("Available audio models for text-to-speech:")
	fmt.Println("==========================================")

	client := llmclient.NewClient(llmclient.WithTimeout(30 * time.Second))

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	models, err := client.ListAudioModels(ctx, &llmclient.AudioModelsRequest{
		Provider: "pollinations",
		APIKey:   "",
	})
	if err != nil {
		fmt.Printf("Warning: Could not fetch models from API: %v\n", err)
		fmt.Println("\nCommon audio models:")
		printModelInfo("elevenlabs", "ElevenLabs TTS", []string{"Default voice"})
		printModelInfo("openai-audio", "OpenAI TTS", []string{"alloy", "echo", "fable", "onyx", "nova", "shimmer"})
		return nil
	}

	ttsModels := llmclient.FilterTextToSpeechModels(models.Models)
	if len(ttsModels) == 0 {
		fmt.Println("No text-to-speech models found")
		return nil
	}

	for _, model := range ttsModels {
		desc := model.Description
		if desc == "" {
			desc = "No description available"
		}

		printModelInfo(model.Name, desc, model.Voices)
	}

	fmt.Printf("\nProvider: pollinations (default and recommended for audio)\n")
	fmt.Printf("\nNote: Audio generation requires an API key.\n")
	fmt.Printf("      Set %s environment variable or pass it as argument.\n", envPollinationsKey)
	fmt.Printf("\nText input options:\n")
	fmt.Printf("  - Direct text: \"Hello, world!\"\n")
	fmt.Printf("  - From file:   @input.txt\n")
	fmt.Printf("  - From stdin:  -\n")
	fmt.Printf("\nExamples:\n")
	fmt.Printf("  %s pollinations elevenlabs YOUR_KEY output.wav \"Hello, world!\"\n", programName)
	fmt.Printf("  %s pollinations elevenlabs \"\" output.wav @speech.txt\n", programName)
	fmt.Printf("  echo \"Hello from pipe\" | %s pollinations elevenlabs \"\" output.wav -\n", programName)
	fmt.Printf("\n  export %s=YOUR_KEY\n", envPollinationsKey)
	fmt.Printf("  %s pollinations elevenlabs \"\" output.wav \"Using env var\"\n", programName)

	return nil
}

func printModelInfo(name, description string, voices []string) {
	fmt.Printf("\n  Model: %s\n", name)
	fmt.Printf("  Description: %s\n", description)

	if len(voices) > 0 {
		fmt.Printf("  Voices: %v\n", voices)
	} else {
		fmt.Printf("  Voices: default\n")
	}
}

func showProviderModels(ctx context.Context, programName, provider string) error {
	client := llmclient.NewClient(llmclient.WithTimeout(30 * time.Second))

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	models, err := client.ListAudioModels(ctx, &llmclient.AudioModelsRequest{
		Provider: provider,
		APIKey:   "",
	})
	if err != nil {
		return fmt.Errorf("failed to fetch models for provider %q: %w", provider, err)
	}

	ttsModels := llmclient.FilterTextToSpeechModels(models.Models)
	sttModels := llmclient.FilterSpeechToTextModels(models.Models)

	fmt.Printf("Audio models for provider: %s\n", provider)
	fmt.Println("==========================================")

	if len(ttsModels) > 0 {
		fmt.Println("\nText-to-Speech Models:")
		for _, model := range ttsModels {
			desc := model.Description
			if desc == "" {
				desc = "No description available"
			}
			printModelInfo(model.Name, desc, model.Voices)
		}
	}

	if len(sttModels) > 0 {
		fmt.Println("\nSpeech-to-Text Models:")
		for _, model := range sttModels {
			desc := model.Description
			if desc == "" {
				desc = "No description available"
			}
			printModelInfo(model.Name, desc, model.Voices)
		}
	}

	if len(ttsModels) == 0 && len(sttModels) == 0 {
		fmt.Println("No audio models found for this provider")
	}

	fmt.Printf("\nExample usage:\n")
	fmt.Printf("  %s %s elevenlabs YOUR_KEY output.wav \"Hello, world!\"\n", programName, provider)

	return nil
}
