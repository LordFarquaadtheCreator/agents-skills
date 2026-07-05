#!/bin/bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKILLS_DIR="$REPO_ROOT/skills"

usage() {
  echo "Usage: $0 <skill-name> [description] [--openai]"
  echo ""
  echo "Positional arguments:"
  echo "  skill-name   Required. Name of the skill (used as directory name)."
  echo "  description  Optional. One-line description for the skill frontmatter."
  echo ""
  echo "Options:"
  echo "  --openai     Also create agents/openai.yaml (OpenAI/Agents SDK format)."
  exit 1
}

# --- Parse args ---
OPENAI=false
args=()

for arg in "$@"; do
  case "$arg" in
    --openai)
      OPENAI=true
      ;;
    --help|-h)
      usage
      ;;
    *)
      args+=("$arg")
      ;;
  esac
done

SKILL_NAME="${args[0]:-}"
DESCRIPTION="${args[1]:-}"

if [ -z "$SKILL_NAME" ]; then
  echo "Error: skill-name is required."
  echo ""
  usage
fi

# Validate skill name: only lowercase letters, digits, and hyphens
if ! [[ "$SKILL_NAME" =~ ^[a-z0-9-]+$ ]]; then
  echo "Error: skill-name must contain only lowercase letters, digits, and hyphens."
  exit 1
fi

SKILL_DIR="$SKILLS_DIR/$SKILL_NAME"

if [ -d "$SKILL_DIR" ]; then
  echo "Error: skill '$SKILL_NAME' already exists at $SKILL_DIR"
  exit 1
fi

# --- Create the skill ---
mkdir -p "$SKILL_DIR"

# SKILL.md
if [ -n "$DESCRIPTION" ]; then
  cat > "$SKILL_DIR/SKILL.md" <<EOF
---
name: $SKILL_NAME
description: $DESCRIPTION
---

# $SKILL_NAME


EOF
else
  cat > "$SKILL_DIR/SKILL.md" <<EOF
---
name: $SKILL_NAME
description:
---

# $SKILL_NAME


EOF
fi

# agents/openai.yaml (optional)
if $OPENAI; then
  mkdir -p "$SKILL_DIR/agents"

  # Derive a display name: convert hyphens to spaces and title-case each word
  display_name=$(echo "$SKILL_NAME" | sed 's/-/ /g' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) substr($i,2)}1')

  short_desc="$DESCRIPTION"
  if [ -z "$short_desc" ]; then
    short_desc="$display_name skill"
  fi

  cat > "$SKILL_DIR/agents/openai.yaml" <<EOF
interface:
  display_name: "$display_name"
  short_description: "$short_desc"
  default_prompt: "Use the $SKILL_NAME skill."
EOF
fi

# --- Summary ---
echo "Created skill '$SKILL_NAME' at $SKILL_DIR"
echo "  SKILL.md"
if $OPENAI; then
  echo "  agents/openai.yaml"
fi
