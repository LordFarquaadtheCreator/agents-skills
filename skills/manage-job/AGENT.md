# Agent Instructions — manage-job

## Structure

```
manage-job/
├── main.go              # rootCmd + main(), flag registration
├── appscript/           # API client package (see appscript/AGENT.md)
│   ├── appscript.go     # AppScript struct: Get, Create, Patch, Delete
│   ├── utils.go         # config loading, repoRoot, sheetsConfig
│   └── appscript_test.go
├── cmd/                 # Cobra commands (see cmd/AGENT.md)
│   ├── get.go           # GetCmd
│   ├── track.go         # TrackCmd
│   ├── patch.go         # PatchCmd
│   └── delete.go        # DeleteCmd
├── go.mod
└── manage-job           # compiled binary (committed)
```

## Deployment ID → URL

The Apps Script deployment ID is stored in `config/sheets-deployment.yaml` at the repository root. The Go binary reads this file at runtime and constructs the web app URL:

```
https://script.google.com/macros/s/<deploymentId>/exec
```

The `deploymentId` key in the YAML file is the hash portion of the URL. When a new deployment is created in the Apps Script editor (Deploy → New deployment), a new ID is generated. Update `config/sheets-deployment.yaml` with the new ID — no code changes needed.

## Config file format

```yaml
deploymentId: AKfycbwRQ52XCi5htaaHLO1Laizu8-pyYFKI0GEWELSnJHsP1CBDc-9OxNlkWGhlG-8l8tDxIQ
```

File must exist at `config/sheets-deployment.yaml`. Directory `config/` is gitignored — this file is local only and must be created on each machine.

## Rebuilding

```bash
cd /Users/farquaad/agents-data/skills/manage-job && go build -o manage-job .
```

The binary is committed to the repo so agents can use it without building.

## Testing

```bash
cd /Users/farquaad/agents-data/skills/manage-job && go test ./appscript/ -v
```
