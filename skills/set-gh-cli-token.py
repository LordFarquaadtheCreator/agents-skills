#!/usr/bin/env python3
import json
import sys
import subprocess
from pathlib import Path

GH_PATS_PATH = Path.home() / "agents-data/config/gh-pats.json"


def swap_gh_cli_token(mode: str) -> None:
    """
    Swap GitHub CLI token using gh auth login --with-token.
    
    Args:
        mode: Either "work_mode" or "personal_mode"
    
    Raises:
        ValueError: If mode is invalid
        FileNotFoundError: If config files don't exist
        json.JSONDecodeError: If JSON parsing fails
        KeyError: If required keys are missing
        subprocess.CalledProcessError: If gh auth login fails
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
    
    # Use gh auth login --with-token to set the token
    try:
        subprocess.run(
            ["gh", "auth", "login", "--with-token"],
            input=new_token,
            text=True,
            check=True,
        )
    except subprocess.CalledProcessError as e:
        raise subprocess.CalledProcessError(
            e.returncode, e.cmd, f"gh auth login failed: {e.stderr if e.stderr else str(e)}"
        )
    
    print(f"Successfully updated gh CLI token to {mode}")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python set-gh-cli-token.py <work_mode|personal_mode>", file=sys.stderr)
        sys.exit(1)
    
    try:
        swap_gh_cli_token(sys.argv[1])
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
