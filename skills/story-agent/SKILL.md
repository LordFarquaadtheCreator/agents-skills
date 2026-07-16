---
name: story-agent
description: Retrieve documents from the story-tools RAG vector store by semantic query
argument-hint: "[query]"
allowed-tools:
  - mcp_call_tool
---

Helps the client find documents by querying the story-tools RAG vector store.

The story-tools MCP exposes a `search_documents` tool that performs hybrid semantic search (content embedding × 0.7 + tags embedding × 0.3) over an ingested vector DB and returns matching chunks with scores. No LLM generation — retrieval only.

Call `search_documents` on the `story-tools` MCP server with the user's query. Present the returned chunks and scores to the user.

If no vector DB is loaded, tell the user to load one first (via `load_vector_db` or `list_vector_dbs`).
