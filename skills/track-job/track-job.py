#!/usr/bin/env python3
import sys
import json
import re
from datetime import datetime
import urllib.request
import urllib.error

VALID_INDUSTRIES = {"Tech", "Health Care", "Retail", "Finance", "Gig", "Other"}
VALID_STATUSES = {
    "Not Started",
    "Applied Only",
    "Applied + Emailed",
    "Applied + Called",
    "Applied + Emailed + Called",
    "Interview!",
    "Done",
}
SCRIPT_URL = "https://script.google.com/macros/s/AKfycbzqNBqQGJgcWDlLHEWr5ppmYDYBchOeg05rT4_ptoM5CkKP0EUy9puAAGU96masWBSuIg/exec"


def validate_url(url):
    if not url or not re.match(r"^https?://", url):
        raise ValueError("Link must be a valid URL starting with http:// or https://")
    return url


def validate_email(email):
    if not email or "@" not in email or "." not in email:
        raise ValueError("Email must be a valid email address")
    return email


def validate_phone(phone):
    if phone:
        digits = re.sub(r"[^\d]", "", phone)
        if not (10 <= len(digits) <= 15):
            raise ValueError("Phone number must be 10-15 digits")
    return phone


def validate_industry(industry):
    if industry not in VALID_INDUSTRIES:
        raise ValueError(f"Industry must be one of: {', '.join(sorted(VALID_INDUSTRIES))}")
    return industry


def validate_status(status):
    if status not in VALID_STATUSES:
        raise ValueError(f"Status must be one of: {', '.join(sorted(VALID_STATUSES))}")
    return status


def validate_date(date_str):
    try:
        datetime.strptime(date_str, "%Y-%m-%d")
        return date_str
    except ValueError:
        raise ValueError("Date must be in ISO 8601 format (YYYY-MM-DD)")


def post_to_sheet(data):
    body = json.dumps(data).encode("utf-8")
    req = urllib.request.Request(
        SCRIPT_URL,
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )

    try:
        with urllib.request.urlopen(req) as response:
            result = response.read().decode("utf-8")
            print(f"Success: {result}")
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


def main():
    if len(sys.argv) < 5:
        print("Usage: track-job <link> <email> <industry> <status> [phone] [notes]", file=sys.stderr)
        print("\nRequired:", file=sys.stderr)
        print("  link     - Job posting URL", file=sys.stderr)
        print("  email    - Employer contact email", file=sys.stderr)
        print("  industry - Tech, Health Care, Retail, Finance, Gig, Other", file=sys.stderr)
        print("  status   - Not Started, Applied Only, Applied + Emailed, Applied + Called,", file=sys.stderr)
        print("             Applied + Emailed + Called, Interview!, Done", file=sys.stderr)
        print("\nOptional:", file=sys.stderr)
        print("  phone    - Contact phone number", file=sys.stderr)
        print("  notes    - Free-form notes", file=sys.stderr)
        return 1

    link = validate_url(sys.argv[1])
    email = validate_email(sys.argv[2])
    industry = sys.argv[3]
    status = sys.argv[4]
    phone = sys.argv[5] if len(sys.argv) > 5 else None
    notes = sys.argv[6] if len(sys.argv) > 6 else None

    if phone:
        phone = validate_phone(phone)

    industry = validate_industry(industry)
    status = validate_status(status)

    today = datetime.now().strftime("%Y-%m-%d")

    data = {
        "link": link,
        "dateApplied": today,
        "industry": industry,
        "phoneNumber": phone,
        "email": email,
        "status": status,
        "notes": notes,
    }

    return post_to_sheet(data)


if __name__ == "__main__":
    sys.exit(main())
