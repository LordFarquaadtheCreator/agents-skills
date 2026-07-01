# Agent Instructions — set-gh-token

## Structure

```
github/
├── main.go              # rootCmd + main()
├── cmd/                 # Cobra commands
│   ├── cli.go           # CliCmd — swaps gh CLI token
│   └── mcp.go           # McpCmd — swaps MCP token in mcp_config.json
├── pats/                # Shared package
│   └── pats.go          # LoadToken(mode) — reads gh-pats.yaml
├── go.mod
├── go.sum
└── set-gh-token         # compiled binary (gitignored)
```

## Config

PATs file at `~/agents-data/config/gh-pats.yaml`:
```yaml
work_PAT: "..."
personal_PAT: "..."
```

MCP config at `~/.codeium/windsurf/mcp_config.json`.

## Rebuilding

```bash
cd /Users/farquaad/agents-data/skills/github && go build -o set-gh-token .
```

## Usage

```bash
./set-gh-token mcp <work_mode|personal_mode>
./set-gh-token cli <work_mode|personal_mode>
```
