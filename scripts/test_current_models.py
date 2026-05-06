#!/usr/bin/env python3
"""
Test the currently enabled models on this deployment.

Current model list was read from the server at 2026-05-06:
  - seedance2
  - sd2.0fast
  - seedream4.5
  - seedream5.0lite

Notes:
  - `seedance2` maps to the same upstream endpoint as `doubao-seedance-2.0`
  - `sd2.0fast` maps to the same upstream endpoint as `doubao-seedance-2.0-fast`
  - `seedream4.5` and `seedream5.0lite` are tested as image models
  - `seedance2` and `sd2.0fast` are tested as video models
"""

from __future__ import annotations

import getpass
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from typing import Any


BASE_URL = os.getenv("NEW_API_BASE_URL", "http://116.62.175.161:3000").rstrip("/")

# Set to 0 if you only want to verify that video tasks can be submitted.
POLL_VIDEO_TASKS = os.getenv("POLL_VIDEO_TASKS", "1") != "0"
VIDEO_TIMEOUT_SECONDS = int(os.getenv("VIDEO_TIMEOUT_SECONDS", "900"))
VIDEO_POLL_INTERVAL_SECONDS = int(os.getenv("VIDEO_POLL_INTERVAL_SECONDS", "15"))
REQUEST_TIMEOUT_SECONDS = int(os.getenv("REQUEST_TIMEOUT_SECONDS", "120"))


MODEL_CASES: list[dict[str, Any]] = [
    {
        "model": "seedance2",
        "kind": "video",
        "prompt": "A young man turns his head slightly and waves to the camera. Natural motion, cinematic lighting, realistic style.",
        "metadata": {"duration": 5, "resolution": "720p", "ratio": "16:9", "watermark": False},
    },
    {
        "model": "sd2.0fast",
        "kind": "video",
        "prompt": "A person slowly walks forward and nods once. Realistic facial details, stable camera, cinematic look.",
        "metadata": {"duration": 5, "resolution": "720p", "ratio": "16:9", "watermark": False},
    },
    {
        "model": "seedream4.5",
        "kind": "image",
        "prompt": "A cinematic portrait of a traveler at sunrise, highly detailed, natural skin texture, soft light, realistic photography.",
        "size": "2048x2048",
    },
    {
        "model": "seedream5.0lite",
        "kind": "image",
        "prompt": "A clean studio portrait with soft lighting, realistic facial details, premium commercial photography, highly detailed.",
        "size": "2048x2048",
    },
]


def request_json(
    method: str,
    path: str,
    api_key: str,
    payload: dict[str, Any] | None = None,
) -> tuple[int, dict[str, Any]]:
    url = BASE_URL + path
    body = None
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Accept": "application/json",
    }
    if payload is not None:
        body = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"

    req = urllib.request.Request(url, data=body, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=REQUEST_TIMEOUT_SECONDS) as resp:
            raw = resp.read().decode("utf-8", errors="replace")
            return resp.status, json.loads(raw) if raw else {}
    except urllib.error.HTTPError as exc:
        raw = exc.read().decode("utf-8", errors="replace")
        try:
            data = json.loads(raw) if raw else {}
        except json.JSONDecodeError:
            data = {"raw": raw}
        return exc.code, data
    except urllib.error.URLError as exc:
        return 0, {"error": str(exc)}


def list_models(api_key: str) -> set[str]:
    status, data = request_json("GET", "/v1/models", api_key)
    if status != 200:
        raise RuntimeError(f"GET /v1/models failed: HTTP {status} {json.dumps(data, ensure_ascii=False)}")
    items = data.get("data") or []
    model_ids = set()
    for item in items:
        if isinstance(item, dict):
            model_id = item.get("id")
            if isinstance(model_id, str) and model_id.strip():
                model_ids.add(model_id.strip())
    return model_ids


def prompt_api_key() -> str:
    api_key = os.getenv("NEW_API_KEY", "").strip()
    if api_key:
        return api_key

    print("请输入 NEW API 的 API Key。输入时不会明文显示。")
    try:
        api_key = getpass.getpass("API Key: ").strip()
    except (EOFError, KeyboardInterrupt):
        print("\n已取消。", file=sys.stderr)
        return ""
    return api_key


def test_image_model(case: dict[str, Any], api_key: str) -> tuple[bool, str]:
    payload = {
        "model": case["model"],
        "prompt": case["prompt"],
        "n": 1,
        "size": case.get("size", "1024x1024"),
    }
    status, data = request_json("POST", "/v1/images/generations", api_key, payload)
    if status != 200:
        return False, f"HTTP {status}: {json.dumps(data, ensure_ascii=False)}"
    images = data.get("data") or []
    if not isinstance(images, list) or not images:
        return False, f"empty image result: {json.dumps(data, ensure_ascii=False)}"
    first = images[0] if isinstance(images[0], dict) else {}
    image_url = first.get("url") or first.get("b64_json")
    if not image_url:
        return False, f"missing image payload: {json.dumps(data, ensure_ascii=False)}"
    return True, "image generation succeeded"


def submit_video_model(case: dict[str, Any], api_key: str) -> tuple[bool, str, str | None]:
    payload = {
        "model": case["model"],
        "prompt": case["prompt"],
        "metadata": case.get("metadata", {}),
    }
    status, data = request_json("POST", "/v1/video/generations", api_key, payload)
    if status != 200:
        return False, f"HTTP {status}: {json.dumps(data, ensure_ascii=False)}", None
    task_id = data.get("task_id") or data.get("id")
    if not isinstance(task_id, str) or not task_id.strip():
        return False, f"missing task id: {json.dumps(data, ensure_ascii=False)}", None
    task_id = task_id.strip()
    task_status = str(data.get("status") or "").strip().lower()
    if task_status == "failed":
        return False, f"video submit failed immediately: {json.dumps(data, ensure_ascii=False)}", task_id
    return True, f"video task submitted: {task_id}", task_id


def poll_video_task(model_name: str, task_id: str, api_key: str) -> tuple[bool, str]:
    deadline = time.time() + VIDEO_TIMEOUT_SECONDS
    last_status = "unknown"
    while time.time() < deadline:
        path = "/v1/video/generations/" + urllib.parse.quote(task_id, safe="")
        status, data = request_json("GET", path, api_key)
        if status != 200:
            return False, f"{model_name} poll failed: HTTP {status}: {json.dumps(data, ensure_ascii=False)}"

        last_status = str(data.get("status") or "").strip().lower() or "unknown"
        if last_status in {"completed", "succeeded", "ready"}:
            return True, f"video task completed: {task_id}"
        if last_status in {"failed", "error", "cancelled"}:
            return False, f"video task failed: {json.dumps(data, ensure_ascii=False)}"

        time.sleep(VIDEO_POLL_INTERVAL_SECONDS)

    return False, f"video task timeout after {VIDEO_TIMEOUT_SECONDS}s (last status: {last_status}, task_id: {task_id})"


def print_summary(results: list[dict[str, str]]) -> None:
    print("测试结果")
    print("--------")
    for item in results:
        print(f"{item['model']}: {item['status']}")
        print(f"  {item['message']}")


def main() -> int:
    api_key = prompt_api_key()
    if not api_key:
        print("未输入 API Key。", file=sys.stderr)
        return 2

    print(f"Base URL: {BASE_URL}")
    print(f"Video polling: {'on' if POLL_VIDEO_TASKS else 'off'}")
    print()

    try:
        available_models = list_models(api_key)
    except Exception as exc:
        print(f"[FATAL] Failed to list models: {exc}", file=sys.stderr)
        return 1

    failures: list[str] = []
    results: list[dict[str, str]] = []

    for case in MODEL_CASES:
        model_name = case["model"]
        kind = case["kind"]
        listed = model_name in available_models
        print(f"==> Testing {model_name} [{kind}] {'(listed)' if listed else '(not listed by /v1/models)'}")

        if kind == "image":
            ok, message = test_image_model(case, api_key)
        else:
            ok, message, task_id = submit_video_model(case, api_key)
            if ok and POLL_VIDEO_TASKS and task_id:
                ok, message = poll_video_task(model_name, task_id, api_key)

        prefix = "[OK]" if ok else "[FAIL]"
        print(f"{prefix} {message}")
        print()

        if not listed:
            message = f"{message}; model not listed by /v1/models"

        results.append(
            {
                "model": model_name,
                "status": "可用" if ok else "失败",
                "message": message,
            }
        )
        if not ok:
            failures.append(f"{model_name}: {message}")

    print_summary(results)
    if failures:
        print()
        print("失败项")
        print("------")
        for item in failures:
            print(f"- {item}")
        return 1

    print()
    print("这 4 个模型都通过了测试。")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
