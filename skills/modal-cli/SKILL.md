---
name: modal-cli
description: >
  Modal CLI reference — full command surface for running, deploying, managing,
  and observing Modal apps and resources from the shell. Use when the user wants
  to run, deploy, serve, shell into, inspect, or manage Modal apps, containers,
  volumes, secrets, queues, dicts, endpoints, environments, profiles, tokens,
  or billing via the `modal` command. Auto-triggers on `modal` CLI invocations.
---

# Modal CLI

Reference for the `modal` command-line interface, installed alongside the
[`modal`](https://pypi.org/project/modal/) Python package.

Source: https://modal.com/docs/cli/latest

## When to use

Use this skill whenever a task requires invoking the `modal` CLI: running or
deploying apps, serving with hot-reload, shelling into containers, managing
storage primitives (volumes, dicts, queues, secrets, images), provisioning LLM
endpoints, inspecting logs, managing environments/profiles/tokens, or pulling
billing/changelog reports.

For authoring Modal app *code* (Python SDK), prefer the
`modal-basic-skills` / `modal-gpu-dev` / `modal-gpu-experiment` skills. This
skill covers the CLI surface only.

## Conventions

- Most commands accept `-e, --env TEXT` to target an Environment. If omitted,
  defers to `MODAL_ENVIRONMENT` env var → active local profile → workspace
  default, in that order.
- Most `list`/`inspect` commands accept `--json` for machine-parseable output.
  Prefer `--json` + `jq` when scripting.
- Destructive commands accept `-y, --yes` to skip confirmation. Do **not** pass
  `-y` on first run when a human should review; the prompt is a safety check.
- `APP_REF` / `FUNC_REF` forms:
  - File: `my_app.py` or `my_app.py::func` or `my_app.py::app_var.func`
  - Module: `-m my_pkg.my_mod` (or `my_pkg.my_mod::func`)
  - App identifier by name: `my-app`; by ID: `ap-xxxxxxxx`
- Run `modal --help` and `modal <command> --help` to verify flags against the
  installed SDK version — online docs may reference newer features.
- `modal --version` shows the installed SDK version.
- `modal changelog --newer` reveals features released after the installed SDK.

## Command groups

| Group | Purpose |
| --- | --- |
| `modal run` | Run a function or local entrypoint. |
| `modal deploy` | Deploy an app (persistent). |
| `modal serve` | Hot-reload web functions during dev. |
| `modal shell` | Interactive shell / one-off command in a container. |
| `modal curl` | Authenticated curl to a web endpoint (experimental). |
| `modal app` | Manage deployed/running apps. |
| `modal container` | Manage and exec into running containers. |
| `modal endpoint` | Create/manage LLM inference endpoints. |
| `modal image` | Manage Images (named tags). |
| `modal volume` | Read/edit `modal.Volume` volumes. |
| `modal dict` | Manage `modal.Dict` objects. |
| `modal queue` | Manage `modal.Queue` objects. |
| `modal secret` | Manage secrets. |
| `modal setup` | Bootstrap config (auth). |
| `modal bootstrap` | Scaffold a sample app. |
| `modal workspace` | Workspace-level settings, proxy tokens, members. |
| `modal environment` | Create/interact with Environments (RBAC, billing). |
| `modal profile` | Switch between profiles. |
| `modal config` | Manage client config for active profile. |
| `modal token` | Manage tokens (API creds). |
| `modal skills` | Install/update Modal's agent skills. |
| `modal billing` | Workspace billing reports. |
| `modal changelog` | Fetch release notes. |
| `modal dashboard` | Open the Modal Dashboard in a browser. |

## Running & deploying

### `modal run` — run a function or local entrypoint

`FUNC_REF` format: `{file or module}::{function name}` or
`{file or module}::{app_var}.{function name}`.

```bash
modal run my_app.py::hello_world
modal run my_app.py                 # single app + single entrypoint/function
modal run -m my_project.my_app      # module path; remote funcs keep module names
```

Options:
- `-n, --name TEXT` — name for this run
- `-w, --write-result TEXT` — write return value (str/bytes) to local path
- `-q, --quiet` — hide progress indicators
- `-d, --detach` — keep app running if local process dies
- `-i, --interactive` — interactive mode
- `-e, --env TEXT` — environment
- `-m` — interpret arg as Python module path
- `--timestamps` — timestamps on log lines

### `modal deploy` — deploy a Modal application

```bash
modal deploy my_script.py
modal deploy -m my_package.my_mod
```

Options:
- `--name TEXT` — deployment name
- `-e, --env TEXT`
- `--stream-logs` — stream logs from the app post-deploy
- `--tag TEXT` — tag the deployment with a version
- `-m` — module path
- `--timestamps`
- `--strategy [rolling|recreate]` — deployment strategy

### `modal serve` — hot-reload web functions

URLs get a `-dev` suffix. Customize via `dev_suffix` in `.modal.toml` or
`MODAL_DEV_SUFFIX` env var (avoids collisions with concurrent `serve` users).

```bash
modal serve hello_world.py
```

Options:
- `-n, --name TEXT`
- `--timeout FLOAT`
- `-e, --env TEXT`
- `-m`
- `--timestamps`

### `modal shell` — shell or one-off command in a container

```bash
modal shell                                    # default Debian image
modal shell hello_world.py::my_function        # use a function's spec
modal shell hello_world.py::MyClass.my_method  # @modal.method
modal shell hello_world.py --cmd=python        # python shell
modal shell hello_world.py -c 'uv pip list' > env.txt
modal shell sb-abc123xyz                       # connect to running Sandbox by ID
```

Options (most apply only when *not* using a REF):
- `-c, --cmd TEXT` — command to run
- `-e, --env TEXT`
- `--image TEXT` — container image tag
- `--add-python TEXT` — add Python to image
- `--volume TEXT` — mount `modal.Volume` at `/mnt/{name}` (repeatable)
- `--add-local TEXT` — mount local file/dir at `/mnt/{basename}` (repeatable)
- `--secret TEXT` — mount `modal.Secret` (repeatable)
- `--cpu INTEGER`
- `--memory INTEGER` — MiB
- `--gpu TEXT` — `any`, `a10g`, `a100:4`, etc.
- `--cloud TEXT` — `aws`|`gcp`|`oci`|`auto`
- `--region TEXT` — single region or comma-separated list
- `--pty / --no-pty`
- `-m` — module path

### `modal curl` — authenticated request to a web endpoint (experimental)

Auth managed via local Modal API creds. Adds latency — for debugging only.
All args after the URL pass through to local `curl`.

```bash
modal curl https://user--my-app.us-west.modal.direct
modal curl -X GET https://user--my-app.us-west.modal.direct
```

## Apps & containers

### `modal app` — manage deployed/running apps

Subcommands:

- `dashboard APP_IDENTIFIER` — open app's dashboard page. `-e, --env`.
- `history APP_IDENTIFIER` — deployment history. `--json`, `-e`.
- `list` — running/deployed/recently stopped apps. `--json`, `-e`.
- `logs APP_IDENTIFIER` — fetch or stream logs. `-f` to follow; default last 100.
- `rollback APP_IDENTIFIER [VERSION]` — redeploy a previous version. App must
  be in "deployed" state. `--strategy [rolling|recreate]`, `-e`.
- `rollover APP_IDENTIFIER` — redeploy to get fresh containers, no code change.
  `--strategy [rolling|recreate]`, `-e`.
- `stop APP_IDENTIFIER` — permanently stop app + terminate containers.
  `-y, --yes`, `-e`.

`modal app logs` options:
- `-f, --follow` — stream until app stops
- `--since TEXT` — ISO 8601 datetime or relative (`1d`, `2h`, `30m`)
- `--until TEXT` — same format as `--since`
- `-n, --tail INTEGER` — last N entries
- `--search TEXT` — filter by text
- `--function TEXT` — Function ID (`fu-*`)
- `--function-call TEXT` — FunctionCall ID (`fc-*`)
- `--container TEXT` — Container ID (`ta-*`)
- `-s, --source TEXT` — `stdout`|`stderr`|`system`
- `--timestamps`
- `--show-function-id` / `--show-function-call-id` / `--show-container-id`
- `-e, --env TEXT`

Examples:
```bash
modal app logs my-app -f
modal app logs my-app --tail 1000
modal app logs my-app --since 2h
modal app logs my-app --since 2026-03-01T05:00:00 --until 2026-03-01T08:00:00
modal app logs my-app --source stderr --function fu-abc123
modal app rollback my-app v3
modal app rollover my-app --strategy recreate
```

### `modal container` — manage and connect to running containers

Subcommands:

- `exec CONTAINER_ID COMMAND...` — run a command in a container.
  `--pty / --no-pty`.
- `list` — running containers. `--app-id TEXT`, `--json`, `-e`.
- `logs CONTAINER_ID` — fetch/stream logs. Same time-range flags as
  `modal app logs` plus `--all`. `-f`, `--since`, `--until`, `-n, --tail`,
  `--search`, `-s, --source`, `--timestamps`.
- `stop CONTAINER_ID` — SIGINT; running inputs cancelled + rescheduled.
  `-y, --yes`.

Container IDs use `ta-*` prefix.

## LLM endpoints

### `modal endpoint` — create/manage LLM inference endpoints

Production-ready LLM inference servers. Docs: https://modal.com/docs/guide/endpoints

Subcommands:

- `create` — deploy a new endpoint.
- `list` — provisioning/running endpoints. `--json`, `-e`.
- `stop ENDPOINT_IDENTIFIER` — permanently stop. `-y`, `-e`.

`modal endpoint create` options:
- `-e, --env TEXT`
- `--name TEXT` — defaults to derived from model name
- `--model TEXT` — HF repo ID for base architecture (e.g.
  `Qwen/Qwen3.6-27B-FP8`). **required**
- `--routing-region TEXT` — defaults to `us-west`
- `--colocate-compute` — run all containers in routing region (price multiplier)
- `--unauthenticated` — allow unauthenticated HTTP
- `--custom-hf-repo TEXT` — HF repo for fine-tuned weights
- `--custom-hf-revision TEXT` — git revision for `--custom-hf-repo`
- `--custom-hf-token TEXT` — HF token for private repo
- `--custom-volume-name TEXT` — Modal Volume with custom weights
- `--custom-volume-path TEXT` — path within Volume

Examples:
```bash
modal endpoint create --model Qwen/Qwen3.6-27B-FP8
modal endpoint create --name qwen-chat --model Qwen/Qwen3.6-27B-FP8
modal endpoint create --name my-ft --model Qwen/Qwen3.6-27B-FP8 \
  --custom-hf-repo acme/qwen-ft --custom-hf-token $HF_TOKEN
modal endpoint create --name my-ft --model Qwen/Qwen3.6-27B-FP8 \
  --custom-volume-name qwen-ft --custom-volume-path /models/qwen
```

## Storage

### `modal image` — manage Images

- `names` — manage named Image tags.
  - `list` — list named Images. `--prefix TEXT`, `--json`, `-e`.

### `modal volume` — read/edit `modal.Volume` volumes

`modal.NetworkFileSystem` users: use `modal nfs` instead.

Subcommands:
- `cp VOLUME_NAME PATHS...` — copy within a volume. `-r, --recursive`, `-e`.
- `create NAME` — create named persistent volume. `--version INTEGER`
  (experimental), `-e`.
- `dashboard VOLUME_NAME` — open volume's dashboard page. `-e`.
- `delete NAME` — delete volume + data. `--allow-missing`, `-y`, `-e`.
- `get VOLUME_NAME REMOTE_PATH [LOCAL_DESTINATION]` — download files. Folders
  download recursively. `-` as destination writes to stdout. `--force`, `-e`.
- `list` — all volumes in environment. `--json`, `-e`.
- `ls VOLUME_NAME [PATH]` — list files/dirs in a volume. `--json`, `-e`.
- `put VOLUME_NAME LOCAL_PATH [REMOTE_PATH]` — upload file/dir. Trailing `/`
  on REMOTE_PATH treats it as a directory. `-f, --force`, `-e`.
- `rename OLD_NAME NEW_NAME` — rename a volume. `-y`, `-e`.
- `rm VOLUME_NAME REMOTE_PATH` — delete file/dir. `-r, --recursive`, `-e`.

Examples:
```bash
modal volume get my-vol logs/april-12-1.txt
modal volume get my-vol / volume_data_dump
modal volume get my-vol file.txt -        # stream to stdout
```

### `modal dict` — manage `modal.Dict` objects

Subcommands:
- `clear NAME` — delete all data. `-y`, `-e`.
- `create NAME` — no-op if exists. `-e`.
- `delete NAME` — delete dict + data. `--allow-missing`, `-y`, `-e`.
- `get NAME KEY` — print value (keys always str via CLI). `-e`.
- `items NAME [N]` — print contents. Truncates by default; use `N` or `--all`.
  `-a, --all`, `-r, --repr`, `--json`, `-e`.
- `list` — all named dicts. `--json`, `-e`.

### `modal queue` — manage `modal.Queue` objects

Subcommands:
- `clear NAME` — remove all data. `-p, --partition TEXT`, `-a, --all`, `-y`, `-e`.
- `create NAME` — no-op if exists. `-e`.
- `delete NAME` — delete queue + data. `--allow-missing`, `-y`, `-e`.
- `len NAME` — length of queue or partition. `-p, --partition TEXT`,
  `-t, --total` (sum across partitions), `-e`.
- `list` — all named queues. `--json`, `-e`.
- `peek NAME [N]` — next N items without removing. `-p, --partition TEXT`, `-e`.

### `modal secret` — manage secrets

Subcommands:
- `create SECRET_NAME [KEYVALUES]...` — create a new secret. KEYVALUES are
  `KEY=VALUE` pairs. `-e`, `--from-dotenv PATH`, `--from-json PATH`,
  `--force` (overwrite if exists).
- `delete NAME` — delete secret. `--allow-missing`, `-y`, `-e`.
- `list` — published secrets. `--json`, `-e`.

## Onboarding

### `modal setup` — bootstrap config

```bash
modal setup [--profile TEXT]
```

### `modal bootstrap` — scaffold a sample app

```bash
modal bootstrap [NAME]
  -o, --output TEXT   # template location
  --force             # overwrite existing output dir
```

## Configuration

### `modal workspace` — current workspace

Top-level account owning resources. Subcommands:
- `members list` — workspace members. `--json`.
- `proxy-tokens` — manage proxy tokens (auth to HTTP interfaces / web
  functions; passed as `Modal-Key` and `Modal-Secret` headers). Prefixes
  `wk-` / `ws-` — not interchangeable with API tokens (`ak-`/`as-`).
  - `create` — `--json`.
  - `list` — `-e, --environment TEXT` (filter), `--json`.
  - `delete TOKEN_ID` — `-y`.
  - `allow TOKEN_ID ENVIRONMENT_NAME` — RBAC: allow token to auth to env.
  - `revoke TOKEN_ID ENVIRONMENT_NAME` — RBAC: revoke env access.

### `modal environment` — environments

Sub-divisions of workspaces; same app in different namespaces. Each env has
its own Secrets; lookups default to same-env entities. Typical: dev + prod.

Subcommands:
- `billing report [ENVIRONMENT_NAME]` — billing report for env. See
  `modal billing report` for shared flags. Frontend for
  `Environment.billing.report` API. Start inclusive, end exclusive.
- `create NAME` — `--restricted` (enable RBAC).
- `delete NAME` — deletes all apps in env + the env irrevocably. `-y`.
- `list` — `--json`.
- `members` — RBAC member management (restricted envs only).
  - `list ENVIRONMENT` — `--json`.
  - `remove ENVIRONMENT MEMBER` — `--service-user`.
  - `update ENVIRONMENT MEMBER` — `--role [contributor|viewer]` (required),
    `--service-user`.
- `update CURRENT_NAME` — `--set-name TEXT`, `--set-web-suffix TEXT`
  (empty string = no suffix).

### `modal profile` — switch profiles

- `activate PROFILE` — change active profile.
- `current` — print active profile.
- `list` — all profiles, active highlighted. `--json`.

### `modal config` — client config for active profile

- `set-environment ENVIRONMENT_NAME` — default env when `--env` omitted. If
  unset and multiple envs exist, commands requiring an env will error.
- `show` — current config (debugging). `--redact / --no-redact` (redacts
  `token_secret`).

### `modal token` — manage tokens

- `info` — info about token currently in use.
- `new` — create token via authenticated web session. `--profile TEXT`,
  `--activate / --no-activate`, `--verify / --no-verify`.
- `set` — set credentials manually (prompts if omitted). `--token-id TEXT`,
  `--token-secret TEXT`, `--profile TEXT`, `--activate / --no-activate`,
  `--verify / --no-verify`.

### `modal skills` — install/update Modal's agent skills

- `install` — `-y, --yes`, `--no-docs`, `-g, --global` (user home),
  `--claude` (install to `.claude/` instead of `.agents/`).
- `show` — print Modal skill content to terminal.
- `update` — same flags as `install`.

## Observability

### `modal billing` — workspace billing

- `report` — billing report for the workspace. Frontend for
  `Workspace.billing.report` API. Start inclusive, end exclusive; full
  intervals only.

Options:
- `--start TEXT` — ISO date (`2025-01-01`) or relative (`yesterday`,
  `3 days ago`). UTC by default.
- `--end TEXT` — same; defaults to now.
- `--for TEXT` — convenience range: `today`, `yesterday`, `this week`,
  `last week`, `this month`, `last month`.
- `-r, --resolution TEXT` — `d` (daily) or `h` (hourly).
- `--tz TEXT` — `local`, offset (`5`, `-4`, `+05:30`), or IANA name. Requires
  hourly resolution.
- `-t, --tag-names TEXT` — comma-separated tag names to include.
- `--show-resources` — break down by resource type (CPU, Memory, GPU types).
- `--json` / `--csv` — output format.

Examples:
```bash
modal billing report --start 2025-12-01 --end 2026-01-01
modal billing report --for "last month" --tag-names team,project
modal billing report --for today --resolution h
modal billing report --for "this month" --show-resources
modal billing report --for yesterday -r h --tz local
modal billing report --for "last month" --csv > report.csv
modal billing report --start 2025-12-01 --json > report.json
```

### `modal changelog` — release notes

Prints changelog as markdown. Useful for including recent updates in agent
context. Default: most recent updates in current release series.

Options:
- `--last INTEGER` — N most recent entries before installed version.
- `--since TEXT` — entries after version (`X.Y.Z`) or date (`YYYY-MM-DD`),
  exclusive.
- `--for TEXT` — entries for version (`X.Y.Z`) or series (`X.Y`).
- `--newer` — entries newer than installed version.
- `--all` — all entries.
- `--json`.

Note: `--since` and `--last` only show changes up to the installed version.

Examples:
```bash
modal changelog --since 1.2.0
modal changelog --since 2026-01-01
modal changelog --newer
modal changelog --last 3
modal changelog --for 1.3.1
```

### `modal dashboard` — open Modal Dashboard in browser

```bash
modal dashboard [OBJECT_ID]
```

## Common workflows

### First-time setup
```bash
modal setup                       # authenticate
modal bootstrap my-app            # scaffold sample app
modal run my-app/hello.py         # test it
```

### Dev loop with hot-reload
```bash
modal serve my_app.py             # web functions, -dev URL suffix
```

### Deploy to production
```bash
modal deploy my_app.py --env prod --tag v1.2.3
modal app logs my-app -f          # stream logs
modal app history my-app          # check deployments
```

### Inspect running resources
```bash
modal app list --json | jq '.[] | .name'
modal container list --app-id ap-123
modal volume ls my-vol /data
modal secret list
```

### Debug a running container
```bash
modal container exec ta-123456 bash --pty
modal container logs ta-123456 -f --since 30m
modal container stop ta-123456
```

### Manage secrets
```bash
modal secret create my-secret API_KEY=xxx DB_URL=yyy
modal secret create dotenv-secret --from-dotenv .env
modal secret delete my-secret
```

### Volume data ops
```bash
modal volume create my-vol
modal volume put my-vol ./local_dir /remote_dir
modal volume get my-vol /remote_dir ./local_copy
modal volume ls my-vol /
modal volume rm my-vol /remote_dir -r
```

### Rollback a bad deploy
```bash
modal app history my-app          # find target version
modal app rollback my-app v3
```

### Provision an LLM endpoint
```bash
modal endpoint create --name qwen --model Qwen/Qwen3.6-27B-FP8
modal endpoint list
modal endpoint stop qwen
```

### Stay current
```bash
modal --version
modal changelog --newer           # features released after installed SDK
```
