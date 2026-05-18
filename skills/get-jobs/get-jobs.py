#!/usr/bin/env python3
import sys
import urllib.request
import urllib.error
import json

SCRIPT_URL = "https://script.google.com/macros/s/AKfycbzqNBqQGJgcWDlLHEWr5ppmYDYBchOeg05rT4_ptoM5CkKP0EUy9puAAGU96masWBSuIg/exec"


def get_jobs():
    req = urllib.request.Request(
        SCRIPT_URL,
        method="GET",
    )

    try:
        with urllib.request.urlopen(req) as response:
            result = response.read().decode("utf-8")
            print(result)
            return 0
    except urllib.error.HTTPError as e:
        print(f"Error: HTTP {e.code} - {e.reason}", file=sys.stderr)
        print(f"Response: {e.read().decode('utf-8')}", file=sys.stderr)
        return 1
    except urllib.error.URLError as e:
        print(f"Error: {e.reason}", file=sys.stderr)
        return 1
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(get_jobs())
