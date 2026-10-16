#!/bin/bash
# Push all posts in blog-content/*.html to the site via POST /api/blog.
#
# Usage:
#   VATS_API_KEY=vats_xxx ./scripts/push-blog-posts.sh            # push as drafts (default)
#   VATS_API_KEY=vats_xxx STATUS=published ./scripts/push-blog-posts.sh
#   VATS_API_KEY=vats_xxx ./scripts/push-blog-posts.sh --dry-run  # show payloads, don't POST
#
# File format: "TITLE:/SLUG:/EXCERPT:/IMAGE:" header lines, "---", then HTML body.
set -euo pipefail
cd "$(dirname "$0")/../blog-content"

if [[ -z "${VATS_API_KEY:-}" ]]; then
  echo "ERROR: set VATS_API_KEY env var (never hardcode the key in this repo)" >&2
  exit 1
fi

API_BASE="${API_BASE:-https://vatsapartment.com}"
STATUS="${STATUS:-draft}"
DRY_RUN="${1:-}"

VATS_API_KEY="$VATS_API_KEY" API_BASE="$API_BASE" STATUS="$STATUS" DRY_RUN="$DRY_RUN" \
python3 <<'PY'
import json, os, glob, sys, urllib.request, urllib.error

api = os.environ["API_BASE"]
key = os.environ["VATS_API_KEY"]
status = os.environ["STATUS"]
dry = os.environ.get("DRY_RUN") == "--dry-run"

def parse(path):
    text = open(path, encoding="utf-8").read()
    header, body = text.split("---", 1)
    meta = {}
    for line in header.strip().splitlines():
        if ":" in line:
            k, v = line.split(":", 1)
            meta[k.strip()] = v.strip()
    return {
        "title": meta["TITLE"],
        "slug": meta["SLUG"],
        "excerpt": meta.get("EXCERPT", ""),
        "content": body.strip(),
        "image_url": meta.get("IMAGE", ""),
        "status": status,
    }

results = []
for path in sorted(glob.glob("*.html")):
    payload = parse(path)
    slug = payload["slug"]
    if dry:
        print(f"DRY {slug:45s} title={payload['title'][:50]!r} bytes={len(payload['content'])}")
        results.append((slug, "DRY"))
        continue
    req = urllib.request.Request(
        f"{api}/api/blog",
        data=json.dumps(payload).encode("utf-8"),
        headers={
            "Content-Type": "application/json",
            "X-API-Key": key,
            # Cloudflare browser-integrity check (error 1010) rejects
            # default python-urllib user agents.
            "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            body = json.loads(resp.read().decode())
            results.append((slug, f"OK {resp.status} id={body.get('id')}"))
    except urllib.error.HTTPError as e:
        detail = e.read().decode()[:120]
        results.append((slug, f"FAIL {e.code} {detail}"))
    except Exception as e:
        results.append((slug, f"FAIL {e}"))

print()
w = max(len(s) for s, _ in results)
for slug, r in results:
    print(f"{slug:<{w}}  {r}")
fails = sum(1 for _, r in results if r.startswith("FAIL"))
print(f"\n{len(results)} posts, {fails} failures, status={status}")
sys.exit(1 if fails else 0)
PY
