# Text-to-Audio Generator

![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)

[README на русском](README_RU.md)

A CLI utility for generating audio from text using various TTS providers.

## Features

- Audio generation via multiple providers (Pollinations, ElevenLabs, OpenAI)
- Load text from file, stdin, or direct input
- View available models for each provider
- API key support via arguments or environment variables
- Graceful shutdown (SIGINT/SIGTERM)

## Installation

```bash
go install tts-generator@latest
```

Or build from source:

```bash
git clone https://github.com/username/text2audio.git
cd text2audio
go build -o text2audio .
```

## Usage

```bash
text2audio {provider} {model} {api_key} {output_file} {text|@file|-}
text2audio {provider} models
text2audio -h
```

### Arguments

| Argument | Description |
|----------|-------------|
| `provider` | TTS provider (pollinations) |
| `model` | Model for audio generation |
| `api_key` | API key (or empty string to use env) |
| `output_file` | Path to output audio file |
| `text` | Text: direct input, `@file`, or `-` for stdin |

### Examples

Direct text input:
```bash
text2audio pollinations elevenlabs YOUR_KEY output.wav "Hello, world!"
```

From file:
```bash
text2audio pollinations elevenlabs "" output.wav @speech.txt
```

From stdin:
```bash
echo "Hello from pipe" | text2audio pollinations elevenlabs "" output.wav -
```

Using environment variable:
```bash
export POLLINATIONS_API_KEY=YOUR_KEY
text2audio pollinations elevenlabs "" output.wav "Using env var"
```

List provider models:
```bash
text2audio pollinations models
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `POLLINATIONS_API_KEY` | API key for Pollinations (used if argument is empty) |

## Dependencies

- [github.com/SkalaSkalolaz/llmclient](https://github.com/SkalaSkalolaz/llmclient) v1.1.1

## License

[BSD 3-Clause License](LICENSE)
