#!/usr/bin/env python3
"""Validate every stdio MCP server in Hermes' config against Hermes' own
mcp_types model (the one that rejected pawchive's tools/list).

Run with the Hermes venv python:
  /Users/farquaad/.hermes/hermes-agent/venv/bin/python check_hermes_mcps.py [server ...]
"""

import json
import os
import select
import subprocess
import sys
import time

import yaml
from mcp_types._v2025_11_25 import ListToolsResult

CONFIG = "/Users/farquaad/.hermes/config.yaml"

INIT = {
    "jsonrpc": "2.0", "id": 1, "method": "initialize",
    "params": {"protocolVersion": "2025-11-25", "capabilities": {},
               "clientInfo": {"name": "hermes-model-check", "version": "0"}},
}


def read_response(proc, req_id, timeout):
    deadline = time.monotonic() + timeout
    while True:
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            raise TimeoutError(f"no response for id {req_id}")
        ready, _, _ = select.select([proc.stdout], [], [], remaining)
        if not ready:
            continue
        line = proc.stdout.readline()
        if not line:
            raise RuntimeError("server closed stdout")
        line = line.strip()
        if not line:
            continue
        try:
            msg = json.loads(line)
        except json.JSONDecodeError:
            POLLUTION.append(line[:120])
            continue
        if msg.get("id") == req_id:
            return msg


POLLUTION = []


def summarize(exc):
    if hasattr(exc, "errors"):
        parts = []
        for err in exc.errors()[:20]:
            loc = ".".join(str(p) for p in err["loc"])
            parts.append(f"{loc}: {err['msg']} (input={err.get('input')!r})")
        return "; ".join(parts)
    return str(exc)


def check(name, spec, timeout=30):
    POLLUTION.clear()
    cmd = [spec["command"]] + [str(a) for a in spec.get("args") or []]
    env = os.environ.copy()
    env.update({str(k): str(v) for k, v in (spec.get("env") or {}).items()})
    try:
        proc = subprocess.Popen(cmd, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                stderr=subprocess.DEVNULL, text=True, env=env)
    except OSError as exc:
        return f"ERR  {name}: {exc}"
    try:
        proc.stdin.write(json.dumps(INIT) + "\n")
        proc.stdin.flush()
        read_response(proc, 1, timeout)
        proc.stdin.write(json.dumps({"jsonrpc": "2.0", "method": "notifications/initialized"}) + "\n")
        proc.stdin.write(json.dumps({"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}}) + "\n")
        proc.stdin.flush()
        result = read_response(proc, 2, timeout)["result"]
        try:
            model = ListToolsResult.model_validate(result)
            note = f" [stdout pollution: {POLLUTION[0]!r}]" if POLLUTION else ""
            return f"OK   {name}: {len(model.tools)} tools{note}"
        except Exception as exc:  # noqa: BLE001 - report pydantic error verbatim
            return f"FAIL {name}: {summarize(exc)}"
    except Exception as exc:  # noqa: BLE001 - spawn/handshake errors
        return f"ERR  {name}: {exc}"
    finally:
        try:
            proc.stdin.close()
        except Exception:
            pass
        try:
            proc.wait(timeout=10)
        except subprocess.TimeoutExpired:
            proc.kill()


def main():
    if len(sys.argv) >= 3 and sys.argv[1] == "--specs":
        for spec in json.loads(sys.argv[2]):
            print(check(spec["name"], spec), flush=True)
        return
    wanted = set(sys.argv[1:])
    with open(CONFIG) as fh:
        cfg = yaml.safe_load(fh)
    servers = cfg.get("mcp_servers") or {}
    for name in sorted(servers):
        spec = servers[name]
        if wanted and name not in wanted:
            continue
        if not isinstance(spec, dict) or "command" not in spec:
            print(f"SKIP {name}: not stdio ({spec})")
            continue
        label = "" if spec.get("enabled", True) else " [disabled]"
        print(check(name, spec) + label, flush=True)


if __name__ == "__main__":
    main()
