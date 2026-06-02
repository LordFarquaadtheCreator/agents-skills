#!/usr/bin/env python3
import json
import sys
import time
import urllib.error
import urllib.request

SCRIPT_URL = "https://script.google.com/macros/s/AKfycbx5PFyxIoe4KutrXMAlzQKYEa-gabA4EWslhxHaUN-_M0Aag4NA3i8NNz5fZ_tjydvbeg/exec"
REQUEST_TIMEOUT = 30
MAX_RETRIES = 3


def _strip_html_wrapping(raw: str) -> str:
    """GAS /exec endpoints sometimes wrap JSON in HTML (CORS workaround)."""
    stripped = raw.strip()
    if stripped.startswith("{"):
        return stripped
    start = stripped.find("{")
    if start == -1:
        return stripped
    end = stripped.rfind("}")
    if end == -1 or end < start:
        return stripped
    candidate = stripped[start : end + 1]
    try:
        json.loads(candidate)
        return candidate
    except json.JSONDecodeError:
        return stripped


def _fetch() -> str:
    """Make the HTTP request with timeout and redirect handling."""
    req = urllib.request.Request(SCRIPT_URL, method="GET")
    with urllib.request.urlopen(req, timeout=REQUEST_TIMEOUT) as response:
        body = response.read().decode("utf-8")

    if not body.strip():
        raise RuntimeError("Empty response from server")

    return body


def _check_gas_error(body: str) -> None:
    """
    Check if a GAS JSON response signals an application-level error.
    GAS often returns HTTP 200 with: {"success": false, "error": "..."}
    or                                   {"success": "error", ...}
    """
    try:
        data = json.loads(body)
    except json.JSONDecodeError:
        return  # not JSON, let caller handle

    success_val = data.get("success")
    if success_val is False or success_val == "error":
        error_msg = data.get("error") or data.get("message") or "Unknown GAS error"
        raise RuntimeError(f"GAS error: {error_msg}")


def get_jobs():
    last_error = None

    for attempt in range(1, MAX_RETRIES + 1):
        try:
            body = _fetch()
            body = _strip_html_wrapping(body)
            _check_gas_error(body)
            print(body)
            return 0

        except urllib.error.HTTPError as e:
            status_code = e.code
            reason = e.reason
            resp_body = e.read().decode("utf-8", errors="replace")
            print(f"Error: HTTP {status_code} - {reason}", file=sys.stderr)
            if resp_body:
                print(f"Response: {resp_body}", file=sys.stderr)
            # 4xx errors are not worth retrying
            if 400 <= status_code < 500:
                return 1
            last_error = f"HTTP {status_code}"

        except urllib.error.URLError as e:
            last_error = str(e.reason)
            print(f"Error: {e.reason}", file=sys.stderr)

        except (json.JSONDecodeError, RuntimeError) as e:
            print(f"Error: {e}", file=sys.stderr)
            return 1  # not worth retrying structural errors

        except Exception as e:
            last_error = str(e)
            print(f"Error: {e}", file=sys.stderr)

        if attempt < MAX_RETRIES:
            wait = 2**attempt
            print(
                f"Retrying in {wait}s (attempt {attempt}/{MAX_RETRIES})...",
                file=sys.stderr,
            )
            time.sleep(wait)

    print(f"Failed after {MAX_RETRIES} attempts: {last_error}", file=sys.stderr)
    return 1


if __name__ == "__main__":
    sys.exit(get_jobs())
