#!/usr/bin/env python3
import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed
from difflib import unified_diff
from typing import Any, Dict, List, Optional, Tuple


def load_env_file(path: str) -> Dict[str, str]:
    env: Dict[str, str] = {}
    if not os.path.exists(path):
        return env
    with open(path, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            key, value = line.split("=", 1)
            env[key.strip()] = value.strip()
    return env


def normalize(obj: Any) -> Any:
    if isinstance(obj, dict):
        return {k: normalize(obj[k]) for k in sorted(obj.keys())}
    if isinstance(obj, list):
        if not obj:
            return obj
        if all(isinstance(x, dict) for x in obj):
            keys = set.intersection(*(set(x.keys()) for x in obj)) if obj else set()
            if "id" in keys:
                return [normalize(x) for x in sorted(obj, key=lambda x: (x.get("id") is None, x.get("id")))]
            if "name" in keys:
                return [normalize(x) for x in sorted(obj, key=lambda x: (x.get("name") is None, x.get("name")))]
        if all(isinstance(x, (str, int, float, bool, type(None))) for x in obj):
            return sorted(obj, key=lambda x: (x is None, x))
        return [normalize(x) for x in obj]
    return obj


def fetch_json(url: str, headers: Dict[str, str], timeout: int) -> Tuple[int, str, Optional[Any]]:
    req = urllib.request.Request(url, headers=headers, method="GET")
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            status = resp.status
            raw = resp.read().decode("utf-8", errors="replace")
            try:
                return status, raw, json.loads(raw)
            except json.JSONDecodeError:
                return status, raw, None
    except urllib.error.HTTPError as e:
        raw = e.read().decode("utf-8", errors="replace") if e.fp else ""
        try:
            return e.code, raw, json.loads(raw) if raw else None
        except json.JSONDecodeError:
            return e.code, raw, None
    except Exception as e:
        return 0, f"ERROR: {e}", None


def diff_payload(a: Any, b: Any) -> str:
    a_text = json.dumps(a, ensure_ascii=True, sort_keys=True, indent=2)
    b_text = json.dumps(b, ensure_ascii=True, sort_keys=True, indent=2)
    diff = unified_diff(
        a_text.splitlines(),
        b_text.splitlines(),
        fromfile="local",
        tofile="remote",
        lineterm="",
    )
    return "\n".join(diff)


def build_endpoints(args: argparse.Namespace) -> List[Tuple[str, int]]:
    endpoints: List[Tuple[str, int]] = [
        ("/api/v1/healthz", 10),
        ("/api/v1/projects?page=1&per_page=20", 20),
        ("/api/v1/projects/dropdown?page=1&per_page=20", 20),
        ("/api/v1/customers?page=1&per_page=20", 20),
        ("/api/v1/campaigns?page=1&per_page=20", 20),
        ("/api/v1/fields?page=1&per_page=20", 20),
        ("/api/v1/crops?page=1&per_page=20", 20),
        ("/api/v1/workorders?page=1&page_size=20", 20),
        ("/api/v1/lots?page=1&page_size=20", 20),
        ("/api/v1/supplies?page=1&per_page=20", 20),
        ("/api/v1/investors?page=1&per_page=20", 20),
        ("/api/v1/managers?page=1&per_page=20", 20),
        ("/api/v1/lease-types?page=1&per_page=20", 20),
        ("/api/v1/categories?page=1&per_page=20", 20),
        ("/api/v1/types?page=1&per_page=20", 20),
        ("/api/v1/providers?page=1&per_page=20", 20),
    ]

    if args.project_id:
        endpoints.extend(
            [
                (f"/api/v1/dashboard?project_id={args.project_id}", 30),
                (f"/api/v1/projects/{args.project_id}", 20),
                (f"/api/v1/projects/{args.project_id}/fields?page=1&per_page=50", 30),
                (f"/api/v1/projects/{args.project_id}/dollar-values", 30),
                (f"/api/v1/workorders?project_id={args.project_id}&page=1&page_size=20", 30),
                (f"/api/v1/workorders/metrics?project_id={args.project_id}", 60),
                (f"/api/v1/lots?project_id={args.project_id}&page=1&page_size=20", 30),
                (f"/api/v1/lots/metrics?project_id={args.project_id}", 60),
                (f"/api/v1/supplies?project_id={args.project_id}&page=1&per_page=50", 30),
                (f"/api/v1/labors/metrics?project_id={args.project_id}", 60),
                (f"/api/v1/projects/{args.project_id}/labors?page=1&per_page=50", 60),
                (f"/api/v1/projects/{args.project_id}/commercializations", 60),
                (f"/api/v1/projects/{args.project_id}/supply-movements", 60),
                (f"/api/v1/projects/{args.project_id}/stocks/summary?cutoff_date=2024-01-01", 60),
                (f"/api/v1/projects/{args.project_id}/stocks/periods", 60),
                (f"/api/v1/reports/investor-contribution?project_id={args.project_id}", 120),
                (f"/api/v1/reports/summary-results?project_id={args.project_id}", 120),
                (f"/api/v1/data-integrity/costs-check?project_id={args.project_id}", 180),
            ]
        )

    if args.project_id and args.field_id:
        endpoints.extend(
            [
                (
                    f"/api/v1/reports/field-crop?project_id={args.project_id}&field_id={args.field_id}",
                    120,
                ),
                (
                    f"/api/v1/labors/group/{args.project_id}?fieldID={args.field_id}&page=1&page_size=50",
                    60,
                ),
            ]
        )

    if args.customer_id:
        endpoints.append((f"/api/v1/projects/customer/{args.customer_id}?page=1&per_page=20", 30))

    if args.project_name:
        encoded_name = urllib.parse.quote(args.project_name, safe="")
        endpoints.append((f"/api/v1/projects/search?name={encoded_name}&page=1&per_page=20", 30))

    if args.lot_id:
        endpoints.append((f"/api/v1/lots/{args.lot_id}", 20))
    if args.workorder_id:
        endpoints.append((f"/api/v1/workorders/{args.workorder_id}", 20))
        endpoints.append((f"/api/v1/labors/{args.workorder_id}", 30))
        endpoints.append((f"/api/v1/invoice/{args.workorder_id}", 30))
    if args.supply_id:
        endpoints.append((f"/api/v1/supplies/{args.supply_id}", 20))
    if args.field_id:
        endpoints.append((f"/api/v1/fields/{args.field_id}", 20))
    if args.crop_id:
        endpoints.append((f"/api/v1/crops/{args.crop_id}", 20))
    if args.customer_id:
        endpoints.append((f"/api/v1/customers/{args.customer_id}", 20))
    if args.investor_id:
        endpoints.append((f"/api/v1/investors/{args.investor_id}", 20))
    if args.manager_id:
        endpoints.append((f"/api/v1/managers/{args.manager_id}", 20))
    if args.lease_type_id:
        endpoints.append((f"/api/v1/lease-types/{args.lease_type_id}", 20))
    if args.category_id:
        endpoints.append((f"/api/v1/categories/{args.category_id}", 20))
    if args.type_id:
        endpoints.append((f"/api/v1/types/{args.type_id}", 20))

    return endpoints


def main() -> int:
    parser = argparse.ArgumentParser(description="Comparar endpoints local vs remoto")
    parser.add_argument("--local", dest="local_base", default=None, help="Base URL local")
    parser.add_argument("--remote", dest="remote_base", default=None, help="Base URL remoto")
    parser.add_argument("--project-id", dest="project_id")
    parser.add_argument("--field-id", dest="field_id")
    parser.add_argument("--campaign-id", dest="campaign_id")
    parser.add_argument("--customer-id", dest="customer_id")
    parser.add_argument("--workorder-id", dest="workorder_id")
    parser.add_argument("--lot-id", dest="lot_id")
    parser.add_argument("--supply-id", dest="supply_id")
    parser.add_argument("--investor-id", dest="investor_id")
    parser.add_argument("--manager-id", dest="manager_id")
    parser.add_argument("--lease-type-id", dest="lease_type_id")
    parser.add_argument("--category-id", dest="category_id")
    parser.add_argument("--type-id", dest="type_id")
    parser.add_argument("--crop-id", dest="crop_id")
    parser.add_argument("--project-name", dest="project_name")
    parser.add_argument("--workers", dest="workers", type=int, default=6)
    args = parser.parse_args()

    env = load_env_file(os.path.join(os.path.dirname(__file__), "..", "projects", "ponti-api", ".env"))
    local_base = args.local_base or os.environ.get("LOCAL_BASE") or "http://localhost:8080"
    remote_base = args.remote_base or os.environ.get("REMOTE_BASE")
    if not remote_base:
        print("ERROR: definir REMOTE_BASE o pasar --remote", file=sys.stderr)
        return 2

    headers = {
        "X-API-KEY": os.environ.get("X_API_KEY", env.get("X_API_KEY", "abc123secreta")),
        "X-USER-ID": os.environ.get("X_USER_ID", "123"),
        "Accept": "application/json",
    }

    endpoints = build_endpoints(args)
    if not endpoints:
        print("ERROR: no hay endpoints configurados", file=sys.stderr)
        return 2

    print(f"Local:  {local_base}")
    print(f"Remote: {remote_base}")
    print(f"Headers: X-API-KEY={headers['X-API-KEY']} X-USER-ID={headers['X-USER-ID']}")
    print(f"Total endpoints: {len(endpoints)}")
    print("")

    results = []

    def run_one(path: str, timeout: int):
        url_local = f"{local_base}{path}"
        url_remote = f"{remote_base}{path}"
        t0 = time.time()
        st_l, raw_l, json_l = fetch_json(url_local, headers, timeout)
        st_r, raw_r, json_r = fetch_json(url_remote, headers, timeout)
        elapsed = time.time() - t0
        return path, timeout, elapsed, st_l, raw_l, json_l, st_r, raw_r, json_r

    with ThreadPoolExecutor(max_workers=args.workers) as pool:
        future_map = {pool.submit(run_one, p, t): (p, t) for p, t in endpoints}
        for fut in as_completed(future_map):
            results.append(fut.result())

    mismatches = []
    for path, timeout, elapsed, st_l, raw_l, json_l, st_r, raw_r, json_r in sorted(results, key=lambda x: x[0]):
        status_match = st_l == st_r
        if json_l is not None and json_r is not None:
            norm_l = normalize(json_l)
            norm_r = normalize(json_r)
            body_match = norm_l == norm_r
            diff = "" if body_match else diff_payload(norm_l, norm_r)
        else:
            body_match = raw_l.strip() == raw_r.strip()
            diff = "" if body_match else "\n".join(
                unified_diff(raw_l.splitlines(), raw_r.splitlines(), fromfile="local", tofile="remote", lineterm="")
            )
        ok = status_match and body_match
        results_line = "OK" if ok else "DIFF"
        print(f"[{results_line}] {path} ({elapsed:.1f}s) status {st_l}/{st_r}")
        if not ok:
            mismatches.append((path, diff[:4000]))

    print("")
    print(f"OK: {len(endpoints) - len(mismatches)} / {len(endpoints)}")
    if mismatches:
        print("\n--- DIFFS (truncados) ---")
        for path, diff in mismatches:
            print(f"\n# {path}\n{diff}\n")
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
