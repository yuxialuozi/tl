# tl

`tl` is a small terminal tool for learning English vocabulary and translating between English and Chinese. It uses any OpenAI-compatible Chat Completions API and keeps the terminal experience intentionally simple.

## Features

- Look up a single English word and get its part of speech, a simple English definition, a Chinese meaning, and two natural examples.
- Translate English sentences or paragraphs into Chinese, and Chinese text into English.
- Automatically choose vocabulary mode for one English word and translation mode for everything else.
- Read TOML configuration with environment variable overrides.
- Render clean output with no TUI, history, favorites, or interactive mode.

## Installation

Build from source:

```bash
git clone https://github.com/yuxialuozi/tl.git
cd tl
make build
./bin/tl --help
```

To install with Go after the repository is published:

```bash
go install github.com/yuxialuozi/tl/cmd/tl@latest
```

## Configuration

The default configuration path is:

```text
~/.config/tl/config.toml
```

Copy the example and edit it:

```bash
mkdir -p ~/.config/tl
cp configs/config.example.toml ~/.config/tl/config.toml
```

```toml
[provider]
base_url = "https://api.example.com/v1"
api_key = ""
model = "YOUR_MODEL"

[app]
timeout = 30
enable_thinking = false
```

The endpoint must expose the OpenAI-compatible `POST /chat/completions` API. `base_url` normally includes the version prefix, for example `https://api.openai.com/v1`.

Set `enable_thinking = false` to send `enable_thinking: false` with each request. This is useful for Qwen models that support a non-thinking mode. Leave it unset to use the provider default.

API key environment variables are preferred:

```bash
export TL_API_KEY="your-key"
export TL_BASE_URL="https://api.example.com/v1"
export TL_MODEL="your-model"
```

Configuration priority is:

```text
environment variables > config file > defaults
```

`tl config` shows the configuration path, file status, endpoint, model, timeout, and whether an API key is configured. It never prints the API key.

## Usage

Look up one English word:

```bash
tl keyring
```

Translate an English sentence:

```bash
tl "The daemon runs in the background."
```

Translate Chinese text:

```bash
tl "守护进程在后台运行。"
```

Show configuration status:

```bash
tl config
```

Show the version:

```bash
tl version
```

Text containing spaces must be quoted. `tl` accepts exactly one input string, and `config` and `version` are reserved command names.

## Development

Requirements:

- Go 1.23 or newer
- make

Common commands:

```bash
make build
make test
make lint
make clean
```

`make test` runs unit tests with mock HTTP servers and never calls a real API. `make lint` runs `go vet ./...`.

For a release build, pass a version explicitly:

```bash
make build VERSION=0.1.0
```

The version is embedded with linker flags, so future GitHub Release builds can use the same Makefile without changing the source.

## Firefox Extension

A companion Firefox add-on is available in `firefox-extension/`. It shows a small **译** button after selecting English or Chinese text and translates the selection in place.

![tl Firefox extension translating a selected word](docs/images/firefox-extension-demo.png)

See `firefox-extension/README.md` for installation and configuration details. The extension defaults to the same endpoint, model, timeout, and non-thinking mode as the CLI, but the API key must be entered in the extension settings and is never bundled into the source.

## Project Layout

```text
cmd/tl/               CLI entry point
internal/cli/         Cobra command wiring
internal/config/      TOML loading, environment overrides, validation
internal/llm/         OpenAI-compatible HTTP client
internal/translator/  mode detection, prompts, JSON parsing
internal/renderer/    Lipgloss and plain-text output
configs/              example configuration
```

## License

MIT
