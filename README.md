# mdp — Markdown Preview Server

A single long-running local server that previews any markdown file by path, with GitHub-flavored rendering and live reload.

Unlike `gh markdown-preview` which blocks the terminal and spawns a new server per file, `mdp` runs one background server that handles any file you throw at it.

## Install

```bash
go install github.com/wbingli/mdp@latest
```

Or build from source:

```bash
git clone https://github.com/wbingli/mdp.git
cd mdp
go build -o mdp .
```

## Usage

```bash
# Start server in background
mdp start

# Preview a file (auto-starts server if needed)
mdp open README.md

# Check server status
mdp status

# Restart server
mdp restart

# Stop server
mdp stop

# Run server in foreground (for debugging)
mdp serve
```

### Shell alias

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
md() { mdp open "$@"; }
```

Then just:

```bash
md README.md
```

## Features

- **GitHub-flavored markdown** — renders with goldmark GFM (tables, strikethrough, autolinks, task lists)
- **GitHub styling** — uses `github-markdown-css` with automatic dark/light mode
- **Live reload** — edit a file and the browser updates instantly via SSE
- **Relative images** — `![img](./images/foo.png)` just works
- **Recent files** — visit `http://localhost:6419/` to see recently previewed files
- **Single binary** — no runtime dependencies
- **Auto-start** — `mdp open` starts the server automatically if it's not running

## How it works

The URL scheme maps the file's absolute path directly to the URL path:

```
mdp open README.md
→ http://localhost:6419/Users/you/project/README.md
```

Non-markdown files (images, etc.) are served raw, so relative references in markdown resolve naturally.

File changes are detected via `fsnotify` at the directory level, debounced at 100ms, and pushed to the browser as re-rendered HTML over Server-Sent Events.

## Configuration

```bash
mdp serve --port 8080 --host 0.0.0.0
```

Server logs and PID file are stored in `~/.mdp/`.

## License

MIT
