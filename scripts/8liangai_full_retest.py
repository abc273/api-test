#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


BASE_URL = os.getenv("NEW_API_BASE_URL", "https://8liangai.com").rstrip("/")
API_KEY = os.getenv("NEW_API_KEY", "")
ROOT = Path(__file__).resolve().parents[1]
OUT_DIR = ROOT / "tmp" / f"8liangai_full_retest_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
DOWNLOAD_DIR = OUT_DIR / "downloads"
REPORT_PATH = OUT_DIR / "report.md"
RESULTS_PATH = OUT_DIR / "results.json"

QUOTA_PER_UNIT = 500000
POLL_INTERVAL = int(os.getenv("VIDEO_POLL_INTERVAL_SECONDS", "15"))
VIDEO_TIMEOUT = int(os.getenv("VIDEO_TIMEOUT_SECONDS", "1200"))
REQUEST_TIMEOUT = int(os.getenv("REQUEST_TIMEOUT_SECONDS", "180"))

RATIO_DIMS: dict[str, dict[str, tuple[int, int]]] = {
    "480p": {
        "16:9": (864, 496),
        "4:3": (752, 560),
        "1:1": (640, 640),
        "3:4": (560, 752),
        "9:16": (496, 864),
        "21:9": (992, 432),
    },
    "720p": {
        "16:9": (1280, 720),
        "4:3": (1112, 834),
        "1:1": (960, 960),
        "3:4": (834, 1112),
        "9:16": (720, 1280),
        "21:9": (1470, 630),
    },
    "1080p": {
        "16:9": (1920, 1080),
        "4:3": (1664, 1248),
        "1:1": (1440, 1440),
        "3:4": (1248, 1664),
        "9:16": (1080, 1920),
        "21:9": (2206, 946),
    },
}


@dataclass
class Case:
    name: str
    kind: str
    model: str
    resolution: str | None = None
    ratio: str | None = "16:9"
    duration: int = 5
    video_input: bool = False
    size: str | None = None
    expect_success: bool = True
    expected_ratio: float | None = None
    expected_price: float | None = None
    expected_actual_resolution: str | None = None
    notes: str = ""
    payload_extra: dict[str, Any] = field(default_factory=dict)


def now_ts() -> int:
    return int(time.time())


def auth_headers(content_type: str | None = None) -> dict[str, str]:
    headers = {"Authorization": f"Bearer {API_KEY}", "Accept": "application/json"}
    if content_type:
        headers["Content-Type"] = content_type
    return headers


def request(
    method: str,
    path: str,
    payload: dict[str, Any] | None = None,
    headers: dict[str, str] | None = None,
    timeout: int = REQUEST_TIMEOUT,
) -> tuple[int, Any, bytes, dict[str, str]]:
    body = None
    req_headers = auth_headers("application/json" if payload is not None else None)
    if headers:
        req_headers.update(headers)
    if payload is not None:
        body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    req = urllib.request.Request(BASE_URL + path, data=body, headers=req_headers, method=method)
    last_error: Exception | None = None
    for attempt in range(3):
        try:
            with urllib.request.urlopen(req, timeout=timeout) as resp:
                raw = resp.read()
                return resp.status, decode_json(raw), raw, dict(resp.headers)
        except urllib.error.HTTPError as exc:
            raw = exc.read()
            return exc.code, decode_json(raw), raw, dict(exc.headers)
        except (urllib.error.URLError, TimeoutError) as exc:
            last_error = exc
            if attempt < 2:
                time.sleep(3 * (attempt + 1))
                continue
    raise last_error if last_error is not None else RuntimeError("request failed")


def decode_json(raw: bytes) -> Any:
    text = raw.decode("utf-8", "replace")
    try:
        return json.loads(text)
    except Exception:
        return text


def get_logs(limit: int = 200) -> list[dict[str, Any]]:
    status, data, _, _ = request("GET", f"/api/log/token?p=1&page_size={limit}")
    if status != 200 or not isinstance(data, dict):
        return []
    payload = data.get("data")
    if isinstance(payload, list):
        return payload
    if isinstance(payload, dict) and isinstance(payload.get("items"), list):
        return payload["items"]
    return []


def parse_other(log: dict[str, Any]) -> dict[str, Any]:
    raw = log.get("other")
    if not isinstance(raw, str) or not raw:
        return {}
    try:
        return json.loads(raw)
    except Exception:
        return {}


def find_task_logs(task_id: str, since: int) -> list[dict[str, Any]]:
    logs = get_logs(300)
    matched = []
    for log in logs:
        if int(log.get("created_at") or 0) < since - 10:
            continue
        other = parse_other(log)
        if other.get("task_id") == task_id:
            matched.append(log)
    return matched


def summarize_task_cost(task_id: str, since: int) -> dict[str, Any]:
    logs = find_task_logs(task_id, since)
    submit_logs = [l for l in logs if l.get("type") == 2]
    settle_logs = [l for l in logs if l.get("type") == 6]
    error_logs = [l for l in logs if l.get("type") == 5]
    actual_quota = None
    model_ratio = None
    pre_consumed = None
    for log in settle_logs:
        other = parse_other(log)
        if other.get("actual_quota") is not None:
            actual_quota = int(other["actual_quota"])
        if other.get("model_ratio") is not None:
            model_ratio = float(other["model_ratio"])
        if other.get("pre_consumed_quota") is not None:
            pre_consumed = int(other["pre_consumed_quota"])
    if actual_quota is None and submit_logs:
        actual_quota = int(submit_logs[0].get("quota") or 0)
        other = parse_other(submit_logs[0])
        if other.get("model_ratio") is not None:
            model_ratio = float(other["model_ratio"])
    return {
        "logs": logs,
        "submit_quota": int(submit_logs[0].get("quota") or 0) if submit_logs else None,
        "settle_quota": int(settle_logs[0].get("quota") or 0) if settle_logs else None,
        "actual_quota": actual_quota,
        "actual_cny": round(actual_quota / QUOTA_PER_UNIT, 6) if actual_quota is not None else None,
        "model_ratio": model_ratio,
        "pre_consumed_quota": pre_consumed,
        "error_count": len(error_logs),
    }


def ffprobe(path: Path) -> dict[str, Any]:
    cmd = [
        "ffprobe",
        "-v",
        "error",
        "-select_streams",
        "v:0",
        "-show_entries",
        "stream=width,height,r_frame_rate,avg_frame_rate,duration",
        "-show_entries",
        "format=duration,size",
        "-of",
        "json",
        str(path),
    ]
    proc = subprocess.run(cmd, capture_output=True, text=True, timeout=60)
    if proc.returncode != 0:
        return {"error": proc.stderr.strip()}
    return json.loads(proc.stdout)


def expected_dims(case: Case) -> tuple[int, int] | None:
    resolution = case.expected_actual_resolution or case.resolution
    if not resolution or not case.ratio:
        return None
    return RATIO_DIMS.get(resolution, {}).get(case.ratio)


def dims_ok(meta: dict[str, Any], expected: tuple[int, int] | None) -> bool | None:
    if expected is None:
        return None
    streams = meta.get("streams") or []
    if not streams:
        return False
    width = streams[0].get("width")
    height = streams[0].get("height")
    exp_w, exp_h = expected
    return int(width or 0) == exp_w and int(height or 0) == exp_h


def submit_video(case: Case, reference_video_url: str | None) -> dict[str, Any]:
    metadata: dict[str, Any] = {"watermark": False}
    if case.video_input:
        if not reference_video_url:
            return {"ok": False, "error": "missing reference video url for video-input case"}
        metadata["content"] = [
            {
                "type": "video_url",
                "role": "reference_video",
                "video_url": {"url": reference_video_url},
            }
        ]
    metadata.update(case.payload_extra.get("metadata", {}))
    payload: dict[str, Any] = {
        "model": case.model,
        "prompt": f"接口完整回归测试 {case.name}: a calm red panda waves once in a clean studio, stable camera, no text, realistic motion.",
        "resolution": case.resolution,
        "ratio": case.ratio,
        "duration": case.duration,
        "metadata": metadata,
    }
    payload.update({k: v for k, v in case.payload_extra.items() if k != "metadata"})
    since = now_ts()
    status, data, raw, _ = request("POST", "/v1/video/generations", payload)
    result: dict[str, Any] = {
        "case": case.__dict__,
        "submit_status": status,
        "submit_response": data,
        "submit_raw": raw.decode("utf-8", "replace")[:2000],
        "started_at": since,
    }
    task_id = extract_task_id(data)
    result["task_id"] = task_id
    if status >= 400 or not task_id:
        result["ok"] = not case.expect_success
        result["cost"] = {"actual_quota": 0, "actual_cny": 0}
        return result
    deadline = time.time() + VIDEO_TIMEOUT
    business_status = None
    openai_status = None
    while time.time() < deadline:
        business_status = request("GET", f"/v1/video/generations/{task_id}")[1]
        openai_status = request("GET", f"/v1/videos/{task_id}")[1]
        b_status = normalize_business_status(business_status)
        o_status = normalize_openai_status(openai_status)
        if b_status in {"SUCCESS", "FAILURE"} or o_status in {"completed", "failed"}:
            break
        time.sleep(POLL_INTERVAL)
    result["business_status_response"] = business_status
    result["openai_status_response"] = openai_status
    result["business_status"] = normalize_business_status(business_status)
    result["openai_status"] = normalize_openai_status(openai_status)
    result["result_url"] = extract_result_url(business_status, openai_status)
    if result["business_status"] == "SUCCESS" or result["openai_status"] == "completed":
        download_path = DOWNLOAD_DIR / f"{case.name}_{task_id}.mp4"
        dl_status, _, dl_raw, dl_headers = request("GET", f"/v1/videos/{task_id}/content", timeout=300)
        result["download_status"] = dl_status
        result["download_headers"] = dl_headers
        if dl_status == 200 and dl_raw:
            download_path.write_bytes(dl_raw)
            result["download_path"] = str(download_path)
            result["ffprobe"] = ffprobe(download_path)
            exp = expected_dims(case)
            result["expected_dims"] = exp
            result["dims_ok"] = dims_ok(result["ffprobe"], exp)
        else:
            result["download_error"] = dl_raw.decode("utf-8", "replace")[:1000]
    result["cost"] = summarize_task_cost(task_id, since)
    result["ok"] = evaluate_video_result(case, result)
    return result


def extract_task_id(data: Any) -> str | None:
    if not isinstance(data, dict):
        return None
    for key in ("id", "task_id"):
        if isinstance(data.get(key), str):
            return data[key]
    nested = data.get("data")
    if isinstance(nested, dict):
        for key in ("id", "task_id"):
            if isinstance(nested.get(key), str):
                return nested[key]
    return None


def normalize_business_status(data: Any) -> str | None:
    if not isinstance(data, dict):
        return None
    nested = data.get("data")
    if isinstance(nested, dict) and nested.get("status"):
        return str(nested.get("status"))
    if data.get("status"):
        return str(data.get("status"))
    return None


def normalize_openai_status(data: Any) -> str | None:
    if not isinstance(data, dict):
        return None
    if data.get("status"):
        return str(data.get("status"))
    nested = data.get("data")
    if isinstance(nested, dict) and nested.get("status"):
        return str(nested.get("status"))
    return None


def extract_result_url(*responses: Any) -> str | None:
    for data in responses:
        if not isinstance(data, dict):
            continue
        nested = data.get("data")
        candidates = [data]
        if isinstance(nested, dict):
            candidates.append(nested)
        metadata = data.get("metadata")
        if isinstance(metadata, dict):
            candidates.append(metadata)
        for item in candidates:
            for key in ("url", "video_url", "output_url"):
                if isinstance(item.get(key), str) and item[key].startswith("http"):
                    return item[key]
            content = item.get("content")
            if isinstance(content, dict) and isinstance(content.get("video_url"), str):
                return content["video_url"]
    return None


def evaluate_video_result(case: Case, result: dict[str, Any]) -> bool:
    if not case.expect_success:
        return result.get("submit_status", 0) >= 400 or result.get("business_status") == "FAILURE"
    if result.get("submit_status") != 200:
        return False
    if result.get("business_status") != "SUCCESS" and result.get("openai_status") != "completed":
        return False
    if result.get("download_status") != 200:
        return False
    if result.get("dims_ok") is False:
        return False
    cost = result.get("cost") or {}
    actual_cny = cost.get("actual_cny")
    if actual_cny is None:
        return False
    if case.expected_ratio is not None and cost.get("model_ratio") is not None:
        if abs(float(cost["model_ratio"]) - case.expected_ratio) > 0.001:
            return False
    return True


def submit_image(case: Case) -> dict[str, Any]:
    payload = {
        "model": case.model,
        "prompt": f"接口完整回归测试 {case.name}: premium product photo of a ceramic teapot on a wooden table, natural light.",
        "size": case.size,
        "n": 1,
    }
    payload.update(case.payload_extra)
    since = now_ts()
    status, data, raw, _ = request("POST", "/v1/images/generations", payload, timeout=300)
    logs = get_logs(100)
    recent_model_logs = [
        l
        for l in logs
        if l.get("model_name") == case.model and int(l.get("created_at") or 0) >= since - 10
    ]
    cost_quota = None
    if recent_model_logs:
        cost_quota = int(recent_model_logs[0].get("quota") or 0)
    result = {
        "case": case.__dict__,
        "submit_status": status,
        "submit_response": data,
        "submit_raw": raw.decode("utf-8", "replace")[:2000],
        "logs": recent_model_logs[:5],
        "cost": {
            "actual_quota": cost_quota,
            "actual_cny": round(cost_quota / QUOTA_PER_UNIT, 6) if cost_quota is not None else None,
        },
    }
    success = status == 200 and isinstance(data, dict) and bool(data.get("data"))
    result["ok"] = success if case.expect_success else not success
    return result


def probe_assets() -> list[dict[str, Any]]:
    cases = [
        ("official_config", "GET", "/api/portrait_assets/official/config", None, True),
        ("official_jobs", "GET", "/api/portrait_assets/official/jobs?p=1&page_size=5", None, True),
        ("official_create_missing_name", "POST", "/api/portrait_assets/official/jobs", {}, False),
        ("virtual_config", "GET", "/api/portrait_assets/virtual/config", None, True),
        ("virtual_group", "GET", "/api/portrait_assets/virtual/group", None, True),
        ("virtual_assets", "GET", "/api/portrait_assets/virtual/assets?p=1&page_size=5", None, True),
        ("virtual_create_invalid", "POST", "/api/portrait_assets/virtual/assets", {"name": "", "asset_id": ""}, False),
    ]
    results = []
    for name, method, path, payload, expect_success in cases:
        status, data, raw, _ = request(method, path, payload)
        business_success = isinstance(data, dict) and data.get("success") is True
        ok = status < 500 and (business_success if expect_success else not business_success)
        results.append(
            {
                "name": name,
                "method": method,
                "path": path,
                "expect_success": expect_success,
                "status": status,
                "response": data,
                "raw": raw.decode("utf-8", "replace")[:2000],
                "ok": ok,
            }
        )
    return results


def build_cases(full_ratio: bool, include_adaptive: bool = False, video_input_all_ratios: bool = False) -> list[Case]:
    ratios = ["16:9"]
    if full_ratio:
        ratios = ["16:9", "4:3", "1:1", "3:4", "9:16", "21:9"]
    if include_adaptive:
        ratios = [*ratios, "adaptive"]
    cases: list[Case] = []
    for ratio in ratios:
        for res in ["480p", "720p", "1080p"]:
            cases.append(Case(f"seedance15_{res}_{ratio.replace(':','x')}", "video", "seedance1.5", res, ratio, expected_ratio=8))
        for res in ["720p", "1080p"]:
            cases.append(Case(f"seedance15sr_{res}_{ratio.replace(':','x')}", "video", "seedance1.5-sr", res, ratio, expected_ratio=10, expected_actual_resolution=res))
        for res in ["480p", "720p", "1080p"]:
            expected = 23 if res in {"480p", "720p"} else 25.5
            cases.append(Case(f"seedance2_{res}_no_video_{ratio.replace(':','x')}", "video", "seedance2", res, ratio, expected_ratio=expected))
        for res in ["720p", "1080p"]:
            cases.append(Case(f"seedance2sr_{res}_{ratio.replace(':','x')}", "video", "seedance2-sr", res, ratio, expected_ratio=24.5, expected_actual_resolution=res))
        for res in ["480p", "720p"]:
            cases.append(Case(f"sd20fast_{res}_{ratio.replace(':','x')}", "video", "sd2.0fast", res, ratio, expected_ratio=18.5))
        for res in ["720p", "1080p"]:
            cases.append(Case(f"seedance20fastsr_{res}_{ratio.replace(':','x')}", "video", "seedance2.0fast-sr", res, ratio, expected_ratio=23.15, expected_actual_resolution=res))
    video_input_ratios = ratios if video_input_all_ratios else ["16:9"]
    if not full_ratio or video_input_all_ratios:
        for ratio in video_input_ratios:
            for res in ["480p", "720p", "1080p"]:
                expected = 15.5 if res in {"480p", "720p"} else 17
                cases.append(Case(f"seedance2_{res}_with_video_{ratio.replace(':','x')}", "video", "seedance2", res, ratio, video_input=True, expected_ratio=expected))
    return cases


def build_image_cases() -> list[Case]:
    return [
        Case("seedream45_valid_2048", "image", "seedream4.5", size="2048x2048", expected_price=0.3),
        Case("seedream45_invalid_1024", "image", "seedream4.5", size="1024x1024", expect_success=False),
        Case("seedream50lite_valid_2048", "image", "seedream5.0lite", size="2048x2048", expected_price=0.25),
        Case("seedream50lite_invalid_1024", "image", "seedream5.0lite", size="1024x1024", expect_success=False),
    ]


def write_report(results: dict[str, Any]) -> None:
    lines = [
        "# 8liangai.com 接口完整回归测试报告",
        "",
        f"- 测试时间：{datetime.now().astimezone().isoformat(timespec='seconds')}",
        f"- 测试站点：`{BASE_URL}`",
        "- 认证方式：`Authorization: Bearer sk-***`",
        f"- 原始结果：`{RESULTS_PATH}`",
        f"- 下载目录：`{DOWNLOAD_DIR}`",
        "",
        "## 1. 结论总览",
        "",
        "| 类别 | 总数 | 通过 | 失败 |",
        "| --- | ---: | ---: | ---: |",
    ]
    for key in ["videos", "images", "assets"]:
        items = results.get(key, [])
        passed = sum(1 for item in items if item.get("ok"))
        lines.append(f"| {key} | {len(items)} | {passed} | {len(items) - passed} |")
    lines += [
        "",
        "## 2. 视频接口结果",
        "",
        "| 用例 | 模型 | 分辨率 | 视频输入 | 提交 | 业务状态 | OpenAI状态 | 下载 | 实际宽高 | 期望宽高 | 实际扣费(CNY) | 计费倍率 | 结论 |",
        "| --- | --- | --- | --- | ---: | --- | --- | ---: | --- | --- | ---: | ---: | --- |",
    ]
    for item in results.get("videos", []):
        case = item["case"]
        probe = item.get("ffprobe") or {}
        streams = probe.get("streams") or []
        actual_dims = ""
        if streams:
            actual_dims = f"{streams[0].get('width')}x{streams[0].get('height')}"
        exp = item.get("expected_dims")
        expected = f"{exp[0]}x{exp[1]}" if exp else ""
        cost = item.get("cost") or {}
        lines.append(
            f"| `{case['name']}` | `{case['model']}` | `{case.get('resolution')}` | {case.get('video_input')} | "
            f"{item.get('submit_status')} | {item.get('business_status') or ''} | {item.get('openai_status') or ''} | "
            f"{item.get('download_status') or ''} | {actual_dims} | {expected} | {cost.get('actual_cny') or ''} | "
            f"{cost.get('model_ratio') or ''} | {'通过' if item.get('ok') else '失败'} |"
        )
    lines += [
        "",
        "## 3. 图片接口结果",
        "",
        "| 用例 | 模型 | size | HTTP | 实际扣费(CNY) | 结论 |",
        "| --- | --- | --- | ---: | ---: | --- |",
    ]
    for item in results.get("images", []):
        case = item["case"]
        cost = item.get("cost") or {}
        lines.append(
            f"| `{case['name']}` | `{case['model']}` | `{case.get('size')}` | {item.get('submit_status')} | "
            f"{cost.get('actual_cny') or ''} | {'通过' if item.get('ok') else '失败'} |"
        )
    lines += [
        "",
        "## 4. 资产接口结果",
        "",
        "| 用例 | 方法 | 路径 | HTTP | 结论 |",
        "| --- | --- | --- | ---: | --- |",
    ]
    for item in results.get("assets", []):
        lines.append(
            f"| `{item['name']}` | `{item['method']}` | `{item['path']}` | {item['status']} | {'通过' if item.get('ok') else '失败'} |"
        )
    lines += [
        "",
        "## 5. 失败详情",
        "",
    ]
    failures = [item for group in ["videos", "images", "assets"] for item in results.get(group, []) if not item.get("ok")]
    if not failures:
        lines.append("本轮未发现失败用例。")
    for item in failures:
        name = item.get("name") or item.get("case", {}).get("name")
        lines.append(f"### `{name}`")
        lines.append("")
        lines.append("```json")
        lines.append(json.dumps(item, ensure_ascii=False, indent=2)[:6000])
        lines.append("```")
        lines.append("")
    REPORT_PATH.write_text("\n".join(lines), encoding="utf-8")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--full-ratio", action="store_true", help="run every supported ratio for every video model/resolution case")
    parser.add_argument("--include-adaptive", action="store_true", help="include adaptive ratio cases; fixed width/height assertion is skipped for adaptive")
    parser.add_argument("--video-input-all-ratios", action="store_true", help="run seedance2 video-input tier cases for every selected ratio")
    parser.add_argument("--video-input-only", action="store_true", help="only run seedance2 video-input tier cases")
    parser.add_argument("--skip-video", action="store_true")
    parser.add_argument("--skip-image", action="store_true")
    parser.add_argument("--skip-assets", action="store_true")
    args = parser.parse_args()

    if not API_KEY:
        print("NEW_API_KEY is required", file=sys.stderr)
        return 2
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    DOWNLOAD_DIR.mkdir(parents=True, exist_ok=True)

    status = request("GET", "/api/status")[1]
    pricing = request("GET", "/api/pricing")[1]
    models = request("GET", "/v1/models")[1]

    results: dict[str, Any] = {
        "base_url": BASE_URL,
        "started_at": datetime.now(timezone.utc).isoformat(),
        "status": status,
        "pricing": pricing,
        "models": models,
        "videos": [],
        "images": [],
        "assets": [],
    }
    reference_video_url = None
    if not args.skip_video:
        cases = build_cases(args.full_ratio, args.include_adaptive, args.video_input_all_ratios)
        if args.video_input_only:
            cases = [case for case in cases if case.video_input]
            reference_video_url = os.getenv("REFERENCE_VIDEO_URL") or None
            if not reference_video_url:
                print("REFERENCE_VIDEO_URL is required for --video-input-only", file=sys.stderr)
                return 2
        for idx, case in enumerate(cases, 1):
            print(f"[video {idx}/{len(cases)}] {case.name}", flush=True)
            item = submit_video(case, reference_video_url)
            results["videos"].append(item)
            if not reference_video_url and item.get("result_url"):
                reference_video_url = item["result_url"]
            RESULTS_PATH.write_text(json.dumps(results, ensure_ascii=False, indent=2), encoding="utf-8")
    if not args.skip_image:
        image_cases = build_image_cases()
        for idx, case in enumerate(image_cases, 1):
            print(f"[image {idx}/{len(image_cases)}] {case.name}", flush=True)
            results["images"].append(submit_image(case))
            RESULTS_PATH.write_text(json.dumps(results, ensure_ascii=False, indent=2), encoding="utf-8")
    if not args.skip_assets:
        print("[assets] probing asset APIs", flush=True)
        results["assets"] = probe_assets()
    results["finished_at"] = datetime.now(timezone.utc).isoformat()
    RESULTS_PATH.write_text(json.dumps(results, ensure_ascii=False, indent=2), encoding="utf-8")
    write_report(results)
    print(f"RESULTS={RESULTS_PATH}")
    print(f"REPORT={REPORT_PATH}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
