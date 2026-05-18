#!/usr/bin/env python3
import json
import sys
from pathlib import Path

MCP_CONFIG_PATH = Path.home() / ".codeium/windsurf/mcp_config.json"
GH_PATS_PATH = Path.home() / "agents-data/config/gh-pats.json"


def swap_github_token(mode: str) -> None:
    """
    Swap GitHub PAT in mcp_config.json based on mode.
    
    Args:
        mode: Either "work_mode" or "personal_mode"
    
    Raises:
        ValueError: If mode is invalid
        FileNotFoundError: If config files don't exist
        json.JSONDecodeError: If JSON parsing fails
        KeyError: If required keys are missing
        PermissionError: If file cannot be written
    """
    # Input validation
    if mode not in ("work_mode", "personal_mode"):
        raise ValueError(f"Invalid mode: {mode}. Must be 'work_mode' or 'personal_mode'")
    
    # Map mode to PAT key
    pat_key = "work_PAT" if mode == "work_mode" else "personal_PAT"
    
    # Read PATs file
    try:
        with open(GH_PATS_PATH, "r") as f:
            pats_data = json.load(f)
    except FileNotFoundError:
        raise FileNotFoundError(f"PATs file not found: {GH_PATS_PATH}")
    except json.JSONDecodeError as e:
        raise json.JSONDecodeError(f"Invalid JSON in PATs file: {e.msg}", e.doc, e.pos)
    
    # Get the new token
    try:
        new_token = pats_data[pat_key]
    except KeyError:
        raise KeyError(f"Key '{pat_key}' not found in {GH_PATS_PATH}")
    
    # Read mcp_config.json
    try:
        with open(MCP_CONFIG_PATH, "r") as f:
            config_data = json.load(f)
    except FileNotFoundError:
        raise FileNotFoundError(f"MCP config not found: {MCP_CONFIG_PATH}")
    except json.JSONDecodeError as e:
        raise json.JSONDecodeError(f"Invalid JSON in MCP config: {e.msg}", e.doc, e.pos)
    
    # Find and update github entry
    try:
        github_entry = config_data["mcpServers"]["github"]
        
        # Update token (replace everything after "Bearer ")
        github_entry["headers"]["Authorization"] = f"Bearer {new_token}"
        
    except KeyError as e:
        raise KeyError(f"Required key missing in github entry: {e}")
    
    # Write back to mcp_config.json
    try:
        with open(MCP_CONFIG_PATH, "w") as f:
            json.dump(config_data, f, indent=2)
    except PermissionError:
        raise PermissionError(f"Cannot write to {MCP_CONFIG_PATH}")
    
    print(f"Successfully updated GitHub token to {mode}")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python set-gh-mcp-token.py <work_mode|personal_mode>", file=sys.stderr)
        sys.exit(1)
    
    try:
        swap_github_token(sys.argv[1])
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
