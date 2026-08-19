FROM python:3.11-slim AS builder

WORKDIR /app
COPY pyproject.toml uv.lock ./
RUN pip install --no-cache-dir uv && uv sync --frozen --no-dev

FROM python:3.11-slim
WORKDIR /app
COPY --from=builder /app/.venv /app/.venv
COPY src/ /app/src/
COPY README.md LICENSE ./
ENV PATH="/app/.venv/bin:$PATH"
ENV TRANSPORT=stdio
ENTRYPOINT ["python", "-m", "patreon_mcp_server.server"]
