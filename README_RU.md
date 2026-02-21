# Text-to-Audio Generator

![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)

CLI-утилита для генерации аудио из текста с использованием различных TTS-провайдеров.

## Возможности

- Генерация аудио через несколько провайдеров (Pollinations, ElevenLabs, OpenAI)
- Загрузка текста из файла, stdin или прямым вводом
- Просмотр доступных моделей для каждого провайдера
- Поддержка API-ключей через аргументы или переменные окружения
- Graceful shutdown (SIGINT/SIGTERM)

## Установка

```bash
go install tts-generator@latest
```

Или сборка из исходников:

```bash
git clone https://github.com/username/text2audio.git
cd text2audio
go build -o text2audio .
```

## Использование

```bash
text2audio {provider} {model} {api_key} {output_file} {text|@file|-}
text2audio {provider} models
text2audio -h
```

### Параметры

| Параметр | Описание |
|----------|----------|
| `provider` | Провайдер TTS (pollinations) |
| `model` | Мель для генерации аудио |
| `api_key` | API-ключ (или пустая строка для использования env) |
| `output_file` | Путь к выходному аудиофайлу |
| `text` | Текст: прямой ввод, `@file` или `-` для stdin |

### Примеры

Прямой ввод текста:
```bash
text2audio pollinations elevenlabs YOUR_KEY output.wav "Hello, world!"
```

Из файла:
```bash
text2audio pollinations elevenlabs "" output.wav @speech.txt
```

Из stdin:
```bash
echo "Hello from pipe" | text2audio pollinations elevenlabs "" output.wav -
```

Использование переменной окружения:
```bash
export POLLINATIONS_API_KEY=YOUR_KEY
text2audio pollinations elevenlabs "" output.wav "Using env var"
```

Список моделей провайдера:
```bash
text2audio pollinations models
```

## Переменные окружения

| Переменная | Описание |
|------------|----------|
| `POLLINATIONS_API_KEY` | API-ключ для Pollinations (используется если аргумент пустой) |

## Зависимости

- [github.com/SkalaSkalolaz/llmclient](https://github.com/SkalaSkalolaz/llmclient) v1.1.1

## Лицензия

[BSD 3-Clause License](LICENSE)
