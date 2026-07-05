#!/bin/bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKILLS_SRC="$REPO_ROOT/skills"

if [ $# -lt 1 ]; then
  echo "Usage: $0 <target-dir>"
  echo "Example: $0 ~/.devin/skills"
  exit 1
fi

TARGET="$1"
mkdir -p "$TARGET"

# symlink each skill
for skill_dir in "$SKILLS_SRC"/*/; do
  skill_name="$(basename "$skill_dir")"
  link_path="$TARGET/$skill_name"
  if [ -L "$link_path" ]; then
    rm "$link_path"
  elif [ -e "$link_path" ]; then
    echo "skip $skill_name: exists and is not a symlink"
    continue
  fi
  ln -s "${skill_dir%/}" "$link_path"
  echo "linked $skill_name"
done

# remove stale symlinks (point to nothing or not in skills/)
for link in "$TARGET"/*; do
  [ -L "$link" ] || continue
  skill_name="$(basename "$link")"
  if [ ! -e "$link" ] || [ ! -d "$SKILLS_SRC/$skill_name" ]; then
    rm "$link"
    echo "removed stale $skill_name"
  fi
done
