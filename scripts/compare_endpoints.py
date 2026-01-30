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
                extra_keys = [
                    k
                    for k in (
                        "project_name",
                        "field_name",
                        "lot_name",
                        "labor_name",
                        "labor_category_name",
                        "type_name",
                        "supply_name",
                        "category_name",
                        "unit_price",
                        "cost_per_ha",
                        "dose",
                        "total_cost",
                    )
                    if k in keys
                ]

                def norm_val(value: Any) -> str:
                    if value is None:
                        return ""
                    return str(value)

                def sort_key(item: Dict[str, Any]) -> Tuple[str, ...]:
                    parts = [norm_val(item.get("id"))]
                    for k in extra_keys:
                        parts.append(norm_val(item.get(k)))
                    return tuple(parts)

                return [normalize(x) for x in sorted(obj, key=sort_key)]
            if "name" in keys:
                return [normalize(x) for x in sorted(obj, key=lambda x: (x.get("name") is None, x.get("name")))]
        if all(isinstance(x, (str, int, float, bool, type(None))) for x in obj):
            return sorted(obj, key=lambda x: (x is None, x))
        return [normalize(x) for x in obj]
    return obj


def fetch_json(url: str, headers: Dict[str, str], timeout: int) -> Tuple[int, str, Optional[Any], str, int]:
    req = urllib.request.Request(url, headers=headers, method="GET")
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            status = resp.status
            content_type = resp.headers.get("content-type", "")
            raw_bytes = resp.read()
            raw = raw_bytes.decode("utf-8", errors="replace")
            try:
                return status, raw, json.loads(raw), content_type, len(raw_bytes)
            except json.JSONDecodeError:
                return status, raw, None, content_type, len(raw_bytes)
    except urllib.error.HTTPError as e:
        raw = e.read().decode("utf-8", errors="replace") if e.fp else ""
        try:
            return e.code, raw, json.loads(raw) if raw else None, "", len(raw.encode("utf-8"))
        except json.JSONDecodeError:
            return e.code, raw, None, "", len(raw.encode("utf-8"))
    except Exception as e:
        return 0, f"ERROR: {e}", None, "", 0


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


def extract_list(payload: Any) -> List[Dict[str, Any]]:
    if isinstance(payload, list):
        return payload
    if isinstance(payload, dict):
        for key in ("data", "items"):
            if isinstance(payload.get(key), list):
                return payload[key]
    return []


def first_item(payload: Any) -> Optional[Dict[str, Any]]:
    items = extract_list(payload)
    return items[0] if items else None


def build_endpoints(args: argparse.Namespace, local_base: str, headers: Dict[str, str]) -> Tuple[List[Tuple[str, int]], Dict[str, Any]]:
    ids: Dict[str, Any] = {
        "project_id": args.project_id,
        "field_id": args.field_id,
        "campaign_id": args.campaign_id,
        "customer_id": args.customer_id,
        "workorder_id": args.workorder_id,
        "lot_id": args.lot_id,
        "supply_id": args.supply_id,
        "investor_id": args.investor_id,
        "manager_id": args.manager_id,
        "lease_type_id": args.lease_type_id,
        "category_id": args.category_id,
        "type_id": args.type_id,
        "crop_id": args.crop_id,
        "project_name": args.project_name,
        "labor_type_id": args.labor_type_id,
        "parameter_key": None,
        "parameter_category": None,
    }

    def pick_first_id(path: str, key: str = "id") -> Optional[Any]:
        st, raw, payload, _, _ = fetch_json(f"{local_base}{path}", headers, 30)
        if st != 200 or payload is None:
            return None
        item = first_item(payload)
        return item.get(key) if isinstance(item, dict) else None

    if not ids["project_id"]:
        st, raw, payload, _, _ = fetch_json(f"{local_base}/api/v1/projects?page=1&per_page=20", headers, 30)
        if st == 200 and payload is not None:
            item = first_item(payload)
            if item:
                ids["project_id"] = item.get("id") or item.get("project_id")
                ids["project_name"] = ids["project_name"] or item.get("name")

    if ids["project_id"] and not ids["project_name"]:
        st, raw, payload, _, _ = fetch_json(f"{local_base}/api/v1/projects/{ids['project_id']}", headers, 30)
        if st == 200 and isinstance(payload, dict):
            ids["project_name"] = payload.get("name")
            ids["customer_id"] = ids["customer_id"] or payload.get("customer_id")
            ids["campaign_id"] = ids["campaign_id"] or payload.get("campaign_id")

    if ids["project_id"] and not ids["field_id"]:
        st, raw, payload, _, _ = fetch_json(
            f"{local_base}/api/v1/projects/{ids['project_id']}/fields",
            headers,
            30,
        )
        if st == 200 and payload is not None:
            item = first_item(payload)
            if item:
                ids["field_id"] = item.get("id") or item.get("field_id")

    if ids["project_id"] and not ids["lot_id"]:
        st, raw, payload, _, _ = fetch_json(
            f"{local_base}/api/v1/lots?project_id={ids['project_id']}&page=1&page_size=20",
            headers,
            30,
        )
        if st == 200 and payload is not None:
            item = first_item(payload)
            if item:
                ids["lot_id"] = item.get("id") or item.get("lot_id")

    if ids["project_id"] and not ids["workorder_id"]:
        st, raw, payload, _, _ = fetch_json(
            f"{local_base}/api/v1/workorders?project_id={ids['project_id']}&page=1&page_size=20",
            headers,
            30,
        )
        if st == 200 and payload is not None:
            items = extract_list(payload)
            if items:
                ids["workorder_id"] = items[0].get("id") or items[0].get("work_order_id")

    if not ids["customer_id"]:
        ids["customer_id"] = pick_first_id("/api/v1/customers?page=1&per_page=20")
    if not ids["campaign_id"]:
        ids["campaign_id"] = pick_first_id("/api/v1/campaigns?page=1&per_page=20")
    if not ids["supply_id"]:
        ids["supply_id"] = pick_first_id(
            f"/api/v1/supplies?project_id={ids['project_id']}&page=1&per_page=50"
            if ids["project_id"]
            else "/api/v1/supplies?page=1&per_page=50"
        )
    if not ids["investor_id"]:
        ids["investor_id"] = pick_first_id("/api/v1/investors?page=1&per_page=20")
    if not ids["manager_id"]:
        ids["manager_id"] = pick_first_id("/api/v1/managers?page=1&per_page=20")
    if not ids["lease_type_id"]:
        ids["lease_type_id"] = pick_first_id("/api/v1/lease-types?page=1&per_page=20")
    if not ids["category_id"]:
        ids["category_id"] = pick_first_id("/api/v1/categories?page=1&per_page=20")
    if not ids["type_id"]:
        ids["type_id"] = pick_first_id("/api/v1/types?page=1&per_page=20")
    if not ids["crop_id"]:
        ids["crop_id"] = pick_first_id("/api/v1/crops")
    if not ids["field_id"]:
        ids["field_id"] = pick_first_id("/api/v1/fields")

    if ids["parameter_key"] is None or ids["parameter_category"] is None:
        st, raw, payload, _, _ = fetch_json(f"{local_base}/api/v1/business-parameters", headers, 30)
        if st == 200 and payload is not None:
            item = first_item(payload)
            if item:
                ids["parameter_key"] = item.get("parameter_key") or item.get("key")
                ids["parameter_category"] = item.get("category")

    endpoints: List[Tuple[str, int]] = [
        ("/healthz", 10),
        ("/ping", 10),
        ("/api/v1/projects?page=1&per_page=20", 20),
        ("/api/v1/projects/dropdown?page=1&per_page=20", 20),
        ("/api/v1/customers?page=1&per_page=20", 20),
        ("/api/v1/campaigns?page=1&per_page=20", 20),
        ("/api/v1/fields", 20),
        ("/api/v1/crops", 20),
        ("/api/v1/work-orders?page=1&page_size=20", 20),
        ("/api/v1/lots?page=1&page_size=20", 60),
        ("/api/v1/supplies?page=1&per_page=20", 20),
        ("/api/v1/investors?page=1&per_page=20", 20),
        ("/api/v1/managers?page=1&per_page=20", 20),
        ("/api/v1/lease-types?page=1&per_page=20", 20),
        ("/api/v1/categories?page=1&per_page=20", 20),
        ("/api/v1/types?page=1&per_page=20", 20),
        ("/api/v1/providers", 20),
        ("/api/v1/business-parameters", 20),
    ]

    if ids["project_id"]:
        endpoints.extend(
            [
                (f"/api/v1/dashboard?project_id={ids['project_id']}", 30),
                (f"/api/v1/projects/{ids['project_id']}", 20),
                (f"/api/v1/projects/{ids['project_id']}/fields", 30),
                (f"/api/v1/projects/{ids['project_id']}/dollar-values", 30),
                (f"/api/v1/work-orders?project_id={ids['project_id']}&page=1&page_size=20", 30),
                (f"/api/v1/work-orders/metrics?project_id={ids['project_id']}", 60),
                (f"/api/v1/work-orders/export?project_id={ids['project_id']}", 120),
                (f"/api/v1/lots?project_id={ids['project_id']}&page=1&page_size=20", 30),
                (f"/api/v1/lots/metrics?project_id={ids['project_id']}", 60),
                (f"/api/v1/lots/export?project_id={ids['project_id']}&page=1&page_size=20", 120),
                (f"/api/v1/supplies?project_id={ids['project_id']}&page=1&per_page=50", 30),
                (f"/api/v1/supplies/export/all?project_id={ids['project_id']}", 120),
                (f"/api/v1/labors/metrics?project_id={ids['project_id']}", 60),
                (f"/api/v1/projects/{ids['project_id']}/labors?page=1&per_page=50", 60),
                (f"/api/v1/projects/{ids['project_id']}/commercializations", 60),
                (f"/api/v1/projects/{ids['project_id']}/supply-movements", 60),
                (f"/api/v1/projects/{ids['project_id']}/supply-movements/providers", 60),
                (f"/api/v1/projects/{ids['project_id']}/supply-movements/export", 120),
                (f"/api/v1/projects/{ids['project_id']}/stocks/summary?cutoff_date=2024-01-01", 60),
                (f"/api/v1/projects/{ids['project_id']}/stocks/periods", 60),
                (f"/api/v1/projects/{ids['project_id']}/stocks/export", 120),
                (f"/api/v1/reports/investor-contribution?project_id={ids['project_id']}", 120),
                (f"/api/v1/reports/summary-results?project_id={ids['project_id']}", 120),
                (f"/api/v1/data-integrity/costs-check?project_id={ids['project_id']}", 180),
            ]
        )

    if ids["project_id"] and ids["field_id"]:
        endpoints.extend(
            [
                (
                    f"/api/v1/reports/field-crop?project_id={ids['project_id']}&field_id={ids['field_id']}",
                    120,
                ),
                (
                    f"/api/v1/labors/group/{ids['project_id']}?field_id={ids['field_id']}&page=1&page_size=50",
                    60,
                ),
            ]
        )

    if ids["customer_id"]:
        endpoints.append((f"/api/v1/projects/customers/{ids['customer_id']}?page=1&per_page=20", 30))

    if ids["project_name"]:
        encoded_name = urllib.parse.quote(ids["project_name"], safe="")
        endpoints.append((f"/api/v1/projects/search?name={encoded_name}&page=1&per_page=20", 30))

    if ids["lot_id"]:
        endpoints.append((f"/api/v1/lots/{ids['lot_id']}", 20))
    if ids["workorder_id"]:
        endpoints.append((f"/api/v1/work-orders/{ids['workorder_id']}", 20))
        endpoints.append((f"/api/v1/labors/{ids['workorder_id']}", 30))
        endpoints.append((f"/api/v1/invoices/{ids['workorder_id']}", 30))
    if ids["supply_id"]:
        endpoints.append((f"/api/v1/supplies/{ids['supply_id']}", 20))
    if ids["field_id"]:
        endpoints.append((f"/api/v1/fields/{ids['field_id']}", 20))
    if ids["crop_id"]:
        endpoints.append((f"/api/v1/crops/{ids['crop_id']}", 20))
    if ids["customer_id"]:
        endpoints.append((f"/api/v1/customers/{ids['customer_id']}", 20))
    if ids["investor_id"]:
        endpoints.append((f"/api/v1/investors/{ids['investor_id']}", 20))
    if ids["manager_id"]:
        endpoints.append((f"/api/v1/managers/{ids['manager_id']}", 20))
    if ids["lease_type_id"]:
        endpoints.append((f"/api/v1/lease-types/{ids['lease_type_id']}", 20))
    if ids["category_id"]:
        endpoints.append((f"/api/v1/categories/{ids['category_id']}", 20))
    if ids["type_id"]:
        endpoints.append((f"/api/v1/types/{ids['type_id']}", 20))
    if ids["parameter_key"]:
        endpoints.append((f"/api/v1/business-parameters/{ids['parameter_key']}", 20))
    if ids["parameter_category"]:
        endpoints.append((f"/api/v1/business-parameters/category/{ids['parameter_category']}", 20))
    if ids["project_id"] and ids["labor_type_id"]:
        endpoints.append((f"/api/v1/projects/{ids['project_id']}/labors/labor-categories/{ids['labor_type_id']}", 60))

    return endpoints, ids


def map_remote_path(local_path: str) -> str:
    # Remoto legacy: endpoints no normalizados
    if local_path.startswith("/healthz"):
        return ""

    if local_path.startswith("/api/v1/business-parameters"):
        return ""

    if local_path.startswith("/api/v1/work-orders"):
        return local_path.replace("/api/v1/work-orders", "/api/v1/workorders", 1)

    if local_path.startswith("/api/v1/invoices/"):
        return local_path.replace("/api/v1/invoices/", "/api/v1/invoice/", 1)

    if local_path.startswith("/api/v1/projects/customers/"):
        return local_path.replace("/api/v1/projects/customers/", "/api/v1/projects/customer/", 1)

    if "/api/v1/projects/" in local_path and "/supply-movements/providers" in local_path:
        return "/api/v1/providers"

    if local_path.startswith("/api/v1/labors/group/") and "field_id=" in local_path:
        return local_path.replace("field_id=", "fieldID=", 1)

    return local_path


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
    parser.add_argument("--labor-type-id", dest="labor_type_id")
    parser.add_argument("--workers", dest="workers", type=int, default=6)
    args = parser.parse_args()

    env = load_env_file(os.path.join(os.path.dirname(__file__), "..", ".env"))
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

    endpoints, ids = build_endpoints(args, local_base, headers)
    if not endpoints:
        print("ERROR: no hay endpoints configurados", file=sys.stderr)
        return 2

    ping_status, _, _, _, _ = fetch_json(f"{local_base}/ping", headers, 10)
    if ping_status == 0:
        print("ERROR: local no responde en /ping", file=sys.stderr)
        return 2

    print(f"Local:  {local_base}")
    print(f"Remote: {remote_base}")
    print(f"Headers: X-API-KEY={headers['X-API-KEY']} X-USER-ID={headers['X-USER-ID']}")
    print(f"IDs: {ids}")
    print(f"Total endpoints: {len(endpoints)}")
    print("")

    results = []

    def run_one(path: str, timeout: int):
        url_local = f"{local_base}{path}"
        remote_path = map_remote_path(path)
        if not remote_path:
            return path, timeout, 0.0, 0, "", None, "", 0, 0, "", None, "", 0, True
        url_remote = f"{remote_base}{remote_path}"
        t0 = time.time()
        st_l, raw_l, json_l, ct_l, len_l = fetch_json(url_local, headers, timeout)
        st_r, raw_r, json_r, ct_r, len_r = fetch_json(url_remote, headers, timeout)
        elapsed = time.time() - t0
        return path, timeout, elapsed, st_l, raw_l, json_l, ct_l, len_l, st_r, raw_r, json_r, ct_r, len_r, False

    with ThreadPoolExecutor(max_workers=args.workers) as pool:
        future_map = {pool.submit(run_one, p, t): (p, t) for p, t in endpoints}
        for fut in as_completed(future_map):
            results.append(fut.result())

    mismatches = []
    binary_keywords = ("/export",)
    skipped = 0
    for path, timeout, elapsed, st_l, raw_l, json_l, ct_l, len_l, st_r, raw_r, json_r, ct_r, len_r, is_skipped in sorted(results, key=lambda x: x[0]):
        if is_skipped:
            skipped += 1
            print(f"[SKIP] {path} (sin equivalente remoto)")
            continue
        status_match = st_l == st_r
        is_binary = any(key in path for key in binary_keywords)
        if is_binary:
            if st_l >= 400 and st_r >= 400:
                body_match = True
                diff = ""
            else:
                body_match = len_l == len_r
                diff = "" if body_match else f"binary length local={len_l} remote={len_r}"
        elif json_l is not None and json_r is not None:
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
    print(f"OK: {len(endpoints) - len(mismatches) - skipped} / {len(endpoints)} (SKIP: {skipped})")
    if mismatches:
        print("\n--- DIFFS (truncados) ---")
        for path, diff in mismatches:
            print(f"\n# {path}\n{diff}\n")
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
