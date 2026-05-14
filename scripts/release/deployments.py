#!/usr/bin/env python3
"""GitHub Deployments helpers for Ponti release workflows.

The workflows use GitHub Deployments as the small source of truth for
staging approvals and rollback targets. Technical identity remains the
immutable artifact SHA; these helpers add the human-operable layer.
"""

from __future__ import annotations

import datetime as dt
import json
import os
import re
import subprocess
import sys
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any


API_VERSION = "2022-11-28"


def die(message: str) -> None:
    print(f"ERROR: {message}", file=sys.stderr)
    sys.exit(1)


def env(name: str, default: str = "", required: bool = False) -> str:
    value = os.getenv(name, default)
    if required and not value:
        die(f"missing required env var {name}")
    return value


def now() -> str:
    return dt.datetime.now(dt.timezone.utc).replace(microsecond=0).isoformat()


def token() -> str:
    return env("GH_TOKEN") or env("GITHUB_TOKEN", required=True)


def repo() -> str:
    return env("GITHUB_REPOSITORY", required=True)


def api_request(method: str, path: str, payload: dict[str, Any] | None = None) -> Any:
    base = env("GITHUB_API_URL", "https://api.github.com").rstrip("/")
    data = None if payload is None else json.dumps(payload).encode("utf-8")
    request = urllib.request.Request(
        f"{base}{path}",
        data=data,
        method=method,
        headers={
            "Accept": "application/vnd.github+json",
            "Authorization": f"Bearer {token()}",
            "Content-Type": "application/json",
            "X-GitHub-Api-Version": API_VERSION,
        },
    )
    try:
        with urllib.request.urlopen(request, timeout=30) as response:
            raw = response.read().decode("utf-8")
            return json.loads(raw) if raw else {}
    except urllib.error.HTTPError as error:
        raw = error.read().decode("utf-8", errors="replace")
        die(f"GitHub API {method} {path} failed: {error.code} {raw[:1000]}")


def write_outputs(values: dict[str, Any]) -> None:
    output_path = os.getenv("GITHUB_OUTPUT")
    if not output_path:
        return
    with open(output_path, "a", encoding="utf-8") as fh:
        for key, value in values.items():
            if value is None:
                value = ""
            fh.write(f"{key}={value}\n")


def append_summary(markdown: str) -> None:
    summary_path = os.getenv("GITHUB_STEP_SUMMARY")
    if summary_path:
        with open(summary_path, "a", encoding="utf-8") as fh:
            fh.write(markdown.rstrip() + "\n")


def latest_migration_version(path: str) -> str:
    if not path:
        return ""
    migrations_dir = Path(path)
    if not migrations_dir.is_dir():
        return ""
    highest = 0
    for item in migrations_dir.iterdir():
        match = re.match(r"(\d+)_.*\.up\.sql$", item.name)
        if match:
            highest = max(highest, int(match.group(1)))
    return str(highest) if highest else ""


def clean_payload(payload: Any) -> dict[str, Any]:
    if isinstance(payload, dict):
        return payload
    if isinstance(payload, str) and payload.strip():
        try:
            parsed = json.loads(payload)
            return parsed if isinstance(parsed, dict) else {}
        except json.JSONDecodeError:
            return {}
    return {}


def require_deploy_sha() -> str:
    value = env("DEPLOY_SHA", required=True)
    if not re.fullmatch(r"[0-9a-f]{40}", value):
        die(f"invalid DEPLOY_SHA: {value}")
    return value


def base_payload(status_label: str, sha: str) -> dict[str, Any]:
    migrations_dir = env("MIGRATIONS_DIR_FOR_RELEASE") or env("MIGRATIONS_DIR")
    migration_version = env("MIGRATION_VERSION") or latest_migration_version(migrations_dir)
    return {
        "status_label": status_label,
        "sha": sha,
        "short_sha": sha[:12],
        "image_uri": env("IMAGE_URI"),
        "service_url": env("SERVICE_URL") or env("DEPLOY_URL"),
        "cloud_run_revision": env("CLOUD_RUN_REVISION"),
        "firebase_hosting_version": env("FIREBASE_HOSTING_VERSION"),
        "ui_artifact_name": env("UI_ARTIFACT_NAME"),
        "ui_artifact_run_id": env("UI_ARTIFACT_RUN_ID") or env("GITHUB_RUN_ID"),
        "migration_version": migration_version,
        "release_version": env("RELEASE_VERSION") or env("SERVICE_VERSION"),
        "build_time": env("SERVICE_BUILD_TIME"),
        "workflow": env("GITHUB_WORKFLOW"),
        "workflow_run_id": env("GITHUB_RUN_ID"),
        "workflow_run_url": workflow_run_url(),
        "actor": env("GITHUB_ACTOR"),
        "recorded_at": now(),
    }


def workflow_run_url() -> str:
    server = env("GITHUB_SERVER_URL", "https://github.com").rstrip("/")
    repository = env("GITHUB_REPOSITORY")
    run_id = env("GITHUB_RUN_ID")
    return f"{server}/{repository}/actions/runs/{run_id}" if repository and run_id else ""


def create_deployment(environment: str, sha: str, payload: dict[str, Any], description: str) -> dict[str, Any]:
    return api_request(
        "POST",
        f"/repos/{repo()}/deployments",
        {
            "ref": sha,
            "environment": environment,
            "auto_merge": False,
            "required_contexts": [],
            "description": description[:140],
            "payload": payload,
            "transient_environment": environment in {"dev", "staging", "staging-approved"},
            "production_environment": environment == "prod",
        },
    )


def create_status(
    deployment_id: int,
    state: str,
    environment: str,
    description: str,
    environment_url: str = "",
) -> dict[str, Any]:
    return api_request(
        "POST",
        f"/repos/{repo()}/deployments/{deployment_id}/statuses",
        {
            "state": state,
            "environment": environment,
            "description": description[:140],
            "log_url": workflow_run_url(),
            "environment_url": environment_url or env("SERVICE_URL") or env("DEPLOY_URL"),
            "auto_inactive": False,
        },
    )


def deployment_statuses(deployment_id: int) -> list[dict[str, Any]]:
    return api_request("GET", f"/repos/{repo()}/deployments/{deployment_id}/statuses?per_page=30")


def latest_status(deployment_id: int) -> dict[str, Any]:
    statuses = deployment_statuses(deployment_id)
    return statuses[0] if statuses else {}


def list_deployments(environment: str) -> list[dict[str, Any]]:
    return api_request("GET", f"/repos/{repo()}/deployments?environment={environment}&per_page=100")


def successful(deployment: dict[str, Any]) -> bool:
    return latest_status(int(deployment["id"])).get("state") == "success"


def find_latest(
    environment: str,
    labels: set[str] | None = None,
    sha: str = "",
    exclude_sha: str = "",
) -> dict[str, Any] | None:
    for deployment in list_deployments(environment):
        payload = clean_payload(deployment.get("payload"))
        deployment_sha = payload.get("sha") or deployment.get("sha") or deployment.get("ref", "")
        if sha and deployment_sha != sha:
            continue
        if exclude_sha and deployment_sha == exclude_sha:
            continue
        if labels and payload.get("status_label") not in labels:
            continue
        if not successful(deployment):
            continue
        deployment["_payload"] = payload
        return deployment
    return None


def command_record() -> None:
    environment = env("DEPLOY_ENV", required=True)
    sha = require_deploy_sha()
    label = env("DEPLOY_STATUS_LABEL", "DEPLOYED")
    state = env("DEPLOY_STATE") or ("failure" if label.endswith("FAILED") else "success")
    payload = base_payload(label, sha)
    description = env("DEPLOY_DESCRIPTION", f"{environment} {label} {sha[:12]}")
    deployment = create_deployment(environment, sha, payload, description)
    create_status(int(deployment["id"]), state, environment, description, payload.get("service_url", ""))
    outputs = {
        "deployment_id": deployment["id"],
        "deploy_sha": sha,
        "status_label": label,
        "image_uri": payload.get("image_uri", ""),
        "migration_version": payload.get("migration_version", ""),
    }
    write_outputs(outputs)
    append_summary(
        f"""### Deployment recorded

| Field | Value |
| --- | --- |
| Environment | `{environment}` |
| Status | `{label}` |
| SHA | `{sha}` |
| Image | `{payload.get("image_uri", "")}` |
| Service URL | `{payload.get("service_url", "")}` |
| Migration version | `{payload.get("migration_version", "")}` |
"""
    )


def command_approve_staging() -> None:
    source = find_latest("staging", labels={"SMOKE_OK", "ROLLED_BACK"})
    if not source:
        die("no successful staging deployment with status SMOKE_OK or ROLLED_BACK was found")
    payload = dict(source["_payload"])
    sha = payload.get("sha") or source.get("sha")
    payload.update(
        {
            "status_label": "QA_APPROVED",
            "approved_at": now(),
            "approved_by": env("GITHUB_ACTOR"),
            "source_deployment_id": source["id"],
        }
    )
    description = f"QA_APPROVED staging {str(sha)[:12]}"
    deployment = create_deployment("staging-approved", sha, payload, description)
    create_status(int(deployment["id"]), "success", "staging-approved", description, payload.get("service_url", ""))
    write_outputs(
        {
            "deployment_id": deployment["id"],
            "deploy_sha": sha,
            "image_uri": payload.get("image_uri", ""),
            "ui_artifact_name": payload.get("ui_artifact_name", ""),
            "ui_artifact_run_id": payload.get("ui_artifact_run_id", ""),
            "migration_version": payload.get("migration_version", ""),
        }
    )
    append_summary(
        f"""### Staging approved

| Field | Value |
| --- | --- |
| Approved SHA | `{sha}` |
| Source deployment | `{source["id"]}` |
| Approved deployment | `{deployment["id"]}` |
| Image | `{payload.get("image_uri", "")}` |
| UI artifact | `{payload.get("ui_artifact_name", "")}` |
| Migration version | `{payload.get("migration_version", "")}` |
"""
    )


def command_resolve() -> None:
    environment = env("RESOLVE_ENV", "staging-approved")
    labels = {x.strip() for x in env("RESOLVE_LABELS").split(",") if x.strip()} or None
    requested_sha = env("RESOLVE_SHA")
    exclude_sha = env("EXCLUDE_SHA")
    if env("EXCLUDE_CURRENT_SHA") == "1":
        current = find_latest(env("CURRENT_ENV", "staging"))
        if current:
            current_payload = clean_payload(current.get("payload"))
            exclude_sha = current_payload.get("sha") or current.get("sha") or exclude_sha
    deployment = find_latest(environment, labels=labels, sha=requested_sha, exclude_sha=exclude_sha)
    if not deployment:
        detail = f"environment={environment}"
        if requested_sha:
            detail += f" sha={requested_sha}"
        if exclude_sha:
            detail += f" exclude_sha={exclude_sha}"
        die(f"no matching deployment found ({detail})")
    payload = deployment["_payload"]
    sha = payload.get("sha") or deployment.get("sha")
    outputs = {
        "deployment_id": deployment["id"],
        "deploy_sha": sha,
        "image_uri": payload.get("image_uri", ""),
        "service_url": payload.get("service_url", ""),
        "cloud_run_revision": payload.get("cloud_run_revision", ""),
        "firebase_hosting_version": payload.get("firebase_hosting_version", ""),
        "ui_artifact_name": payload.get("ui_artifact_name", ""),
        "ui_artifact_run_id": payload.get("ui_artifact_run_id", ""),
        "migration_version": payload.get("migration_version", ""),
        "release_version": payload.get("release_version", ""),
    }
    write_outputs(outputs)
    append_summary(
        f"""### Deployment resolved

| Field | Value |
| --- | --- |
| Source environment | `{environment}` |
| Deployment | `{deployment["id"]}` |
| SHA | `{sha}` |
| Image | `{payload.get("image_uri", "")}` |
| UI artifact | `{payload.get("ui_artifact_name", "")}` |
| Migration version | `{payload.get("migration_version", "")}` |
"""
    )


def run_git(args: list[str]) -> str:
    return subprocess.check_output(["git", *args], text=True).strip()


def parse_semver(tag: str) -> tuple[int, int, int]:
    match = re.fullmatch(r"v(\d+)\.(\d+)\.(\d+)", tag.strip())
    if not match:
        die(f"invalid semver tag: {tag}")
    return tuple(int(x) for x in match.groups())


def semver_tags() -> list[str]:
    tags = run_git(["tag", "--list", "v[0-9]*.[0-9]*.[0-9]*"]).splitlines()
    valid = []
    for tag in tags:
        if re.fullmatch(r"v\d+\.\d+\.\d+", tag):
            valid.append(tag)
    return sorted(valid, key=parse_semver, reverse=True)


def command_next_version() -> None:
    target_sha = env("TARGET_SHA", required=True)
    bump = env("BUMP", "auto")
    tags = semver_tags()
    latest = tags[0] if tags else ""
    if latest:
        log_range = f"{latest}..{target_sha}"
        current = parse_semver(latest)
    else:
        log_range = target_sha
        current = (0, 0, 0)
    subjects = run_git(["log", "--format=%s", log_range]).splitlines()
    bodies = run_git(["log", "--format=%B", log_range])
    if bump == "auto":
        if "BREAKING CHANGE" in bodies or any(re.match(r"^[a-zA-Z]+(?:\([^)]+\))?!:", s) for s in subjects):
            bump = "major"
        elif any(s.startswith("feat") for s in subjects):
            bump = "minor"
        else:
            bump = "patch"
    major, minor, patch = current
    if bump == "major":
        major, minor, patch = major + 1, 0, 0
    elif bump == "minor":
        minor, patch = minor + 1, 0
    elif bump == "patch":
        patch += 1
    else:
        die("BUMP must be auto, patch, minor, or major")
    next_version = f"v{major}.{minor}.{patch}"
    if next_version in tags:
        die(f"tag already exists: {next_version}")

    notes_path = Path(env("RELEASE_NOTES_PATH", "release-notes.md"))
    features = [s for s in subjects if s.startswith("feat")]
    fixes = [s for s in subjects if s.startswith("fix")]
    others = [s for s in subjects if s and s not in features and s not in fixes]
    lines = [
        f"## {next_version}",
        "",
        f"- Target SHA: `{target_sha}`",
        f"- Previous tag: `{latest or 'none'}`",
        f"- Bump: `{bump}`",
        "",
    ]
    for title, items in (("Features", features), ("Fixes", fixes), ("Other changes", others)):
        if items:
            lines.append(f"### {title}")
            lines.extend(f"- {item}" for item in items)
            lines.append("")
    notes_path.write_text("\n".join(lines).rstrip() + "\n", encoding="utf-8")
    write_outputs({"next_version": next_version, "previous_tag": latest, "bump": bump, "notes_path": str(notes_path)})
    append_summary(f"### Next version\n\n`{next_version}` from `{latest or 'none'}` using bump `{bump}`.\n")


def main() -> None:
    if len(sys.argv) != 2:
        die("usage: deployments.py <record|approve-staging|resolve|next-version>")
    command = sys.argv[1]
    if command == "record":
        command_record()
    elif command == "approve-staging":
        command_approve_staging()
    elif command == "resolve":
        command_resolve()
    elif command == "next-version":
        command_next_version()
    else:
        die(f"unknown command: {command}")


if __name__ == "__main__":
    main()
