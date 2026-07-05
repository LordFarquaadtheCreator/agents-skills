---
name: clipboard
description: Copy text output to the user's macOS clipboard using pbcopy. Use this whenever the user asks you to copy something, put something in their clipboard, or wants an output pasted for them to paste elsewhere.
---

# clipboard

This skill gives you the ability to copy text to the user's macOS clipboard using `pbcopy`. Use it whenever the user asks you to copy something for them.

## Usage

Pipe any text output to `pbcopy`:

```bash
echo "text to copy" | pbcopy
```

For multi-line content, use `printf` to avoid issues with shell interpretation of special characters:

```bash
printf '%s' "line 1
line 2
line 3" | pbcopy
```

For content that contains single quotes or other shell-sensitive characters, use a heredoc:

```bash
cat << 'EOF' | pbcopy
content goes here
EOF
```

## When to use this skill

- The user asks "copy that to my clipboard"
- The user says "paste that for me" or "put it in my clipboard"
- You generate a code snippet, summary, or any text the user explicitly wants to paste elsewhere
- The user asks you to copy a command they can run

## Notes

- `pbcopy` only works on macOS
- Binary content (images, files) is **not** supported — this is for text only
- The copy happens silently — no confirmation needed unless something fails
- If `pbcopy` fails, let the user know something went wrong
