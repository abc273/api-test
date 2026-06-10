#!/usr/bin/env python3
from __future__ import annotations

import importlib.util
import json
import os
import sys
import time
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[1]
FULL_SCRIPT = ROOT / "scripts" / "8liangai_full_retest.py"


def load_full_retest_module():
    spec = importlib.util.spec_from_file_location("retest", FULL_SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"failed to load {FULL_SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    sys.modules["retest"] = module
    spec.loader.exec_module(module)
    return module


def refund_delta(cost: dict[str, Any]) -> int | None:
    pre = cost.get("pre_consumed_quota")
    actual = cost.get("actual_quota")
    if pre is None or actual is None:
        return None
    return int(pre) - int(actual)


def stream_dims(item: dict[str, Any]) -> str:
    streams = (item.get("ffprobe") or {}).get("streams") or []
    if not streams:
        return ""
    return f"{streams[0].get('width')}x{streams[0].get('height')}"


def expected_dims(item: dict[str, Any]) -> str:
    dims = item.get("expected_dims")
    if not dims:
        return ""
    return f"{dims[0]}x{dims[1]}"


def sr_source_resolution(model: str, requested: str | None) -> str:
    if model not in {"seedance1.5-sr", "seedance2-sr", "seedance2.0fast-sr"}:
        return requested or ""
    if requested == "720p":
        return "480p"
    if requested == "1080p":
        return "720p"
    return requested or ""


def write_report(results: dict[str, Any], report_path: Path) -> None:
    videos = results.get("videos", [])
    images = results.get("images", [])
    assets = results.get("assets", [])
    video_cost = sum((x.get("cost") or {}).get("actual_cny") or 0 for x in videos)
    image_cost = sum((x.get("cost") or {}).get("actual_cny") or 0 for x in images)

    lines = [
        "# 8liangai.com 计费、退款与分辨率回归测试报告",
        "",
        f"- 测试时间：{results.get('started_at')} 至 {results.get('finished_at')}",
        f"- 测试站点：`{results.get('base_url')}`",
        "- 认证方式：`Authorization: Bearer sk-***`",
        "- 测试口径：真实提交任务，轮询业务状态与 OpenAI 状态，下载视频并用 `ffprobe` 校验输出宽高，读取 `/api/log/token` 核对预扣、实扣和差额结算。",
        f"- 原始结果：`{results.get('results_path')}`",
        "",
        "## 1. 总览",
        "",
        "| 类别 | 覆盖数 | 通过 | 失败 | 实扣金额 |",
        "| --- | ---: | ---: | ---: | ---: |",
        f"| 视频模型 | {len(videos)} | {sum(1 for x in videos if x.get('ok'))} | {sum(1 for x in videos if not x.get('ok'))} | {video_cost:.6f} CNY |",
        f"| 图片模型 | {len(images)} | {sum(1 for x in images if x.get('ok'))} | {sum(1 for x in images if not x.get('ok'))} | {image_cost:.6f} CNY |",
        f"| 资产接口 | {len(assets)} | {sum(1 for x in assets if x.get('ok'))} | {sum(1 for x in assets if not x.get('ok'))} | 0 CNY |",
        "",
        "## 2. 视频模型明细",
        "",
        "| 用例 | 模型 | 请求分辨率 | 内部输入分辨率 | 输出宽高 | 期望宽高 | 计费倍率 | 预扣 | 实扣 | 退回/补扣 | 结论 |",
        "| --- | --- | --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- |",
    ]

    for item in videos:
        case = item.get("case", {})
        cost = item.get("cost") or {}
        delta = refund_delta(cost)
        requested = case.get("resolution") or ""
        model = case.get("model") or ""
        lines.append(
            "| `{name}` | `{model}` | `{requested}` | `{source}` | `{actual_dims}` | `{expected}` | `{ratio}` | `{pre}` | `{actual}` | `{delta}` | {ok} |".format(
                name=case.get("name"),
                model=model,
                requested=requested,
                source=sr_source_resolution(model, requested),
                actual_dims=stream_dims(item),
                expected=expected_dims(item),
                ratio=cost.get("model_ratio") if cost.get("model_ratio") is not None else "",
                pre=cost.get("pre_consumed_quota") if cost.get("pre_consumed_quota") is not None else "",
                actual=cost.get("actual_quota") if cost.get("actual_quota") is not None else "",
                delta=delta if delta is not None else "",
                ok="通过" if item.get("ok") else "失败",
            )
        )

    lines.extend([
        "",
        "说明：",
        "",
        "- `退回/补扣` = `预扣 - 实扣`；正数表示任务完成后退回额度，负数表示补扣额度。",
        "- SR 模型的内部输入分辨率按产品规则记录：用户请求 `720p` 时内部使用 `480p` 基座，用户请求 `1080p` 时内部使用 `720p` 基座。",
        "- `adaptive` 画幅没有固定期望宽高，报告中只校验提交、状态、下载和计费。",
        "",
        "## 3. 图片模型明细",
        "",
        "| 用例 | 模型 | size | HTTP | 实扣金额 | 结论 |",
        "| --- | --- | --- | ---: | ---: | --- |",
    ])

    for item in images:
        case = item.get("case", {})
        cost = item.get("cost") or {}
        lines.append(
            f"| `{case.get('name')}` | `{case.get('model')}` | `{case.get('size')}` | {item.get('submit_status')} | `{cost.get('actual_cny')}` | {'通过' if item.get('ok') else '失败'} |"
        )

    lines.extend([
        "",
        "## 4. 资产接口明细",
        "",
        "| 用例 | 方法 | 路径 | 预期业务成功 | HTTP | 结论 |",
        "| --- | --- | --- | --- | ---: | --- |",
    ])
    for item in assets:
        lines.append(
            f"| `{item.get('name')}` | `{item.get('method')}` | `{item.get('path')}` | `{item.get('expect_success')}` | {item.get('status')} | {'通过' if item.get('ok') else '失败'} |"
        )

    failures = [x for x in [*videos, *images, *assets] if not x.get("ok")]
    lines.extend(["", "## 5. 失败项", ""])
    if not failures:
        lines.append("本轮未发现失败项。")
    for item in failures:
        name = item.get("case", {}).get("name") or item.get("name")
        lines.extend([
            f"### `{name}`",
            "",
            "```json",
            json.dumps(item, ensure_ascii=False, indent=2)[:6000],
            "```",
            "",
        ])

    report_path.write_text("\n".join(lines), encoding="utf-8")


def print_summary(item: dict[str, Any]) -> None:
    case = item.get("case", {})
    cost = item.get("cost") or {}
    print(
        "  ok={ok} dims={dims} expected={expected} ratio={ratio} pre={pre} actual={actual} refund_delta={delta} task={task}".format(
            ok=item.get("ok"),
            dims=stream_dims(item),
            expected=item.get("expected_dims"),
            ratio=cost.get("model_ratio"),
            pre=cost.get("pre_consumed_quota"),
            actual=cost.get("actual_quota"),
            delta=refund_delta(cost),
            task=item.get("task_id"),
        ),
        flush=True,
    )


def main() -> int:
    if not os.getenv("NEW_API_KEY"):
        print("NEW_API_KEY is required", file=sys.stderr)
        return 2

    mod = load_full_retest_module()
    resume_path = os.getenv("RESUME_RESULTS_PATH")
    if resume_path:
        results_path = Path(resume_path)
        out_dir = results_path.parent
    else:
        stamp = time.strftime("%Y%m%d_%H%M%S")
        out_dir = ROOT / "tmp" / f"8liangai_billing_resolution_{stamp}"
    mod.OUT_DIR = out_dir
    mod.DOWNLOAD_DIR = out_dir / "downloads"
    mod.REPORT_PATH = out_dir / "legacy_report.md"
    mod.RESULTS_PATH = out_dir / "results.json"
    report_path = out_dir / "billing-resolution-report.md"
    out_dir.mkdir(parents=True, exist_ok=True)
    mod.DOWNLOAD_DIR.mkdir(parents=True, exist_ok=True)

    C = mod.Case
    video_cases = [
        C("seedance15_480p_16x9", "video", "seedance1.5", "480p", "16:9", expected_ratio=8),
        C("seedance15_720p_16x9", "video", "seedance1.5", "720p", "16:9", expected_ratio=8),
        C("seedance15_1080p_16x9", "video", "seedance1.5", "1080p", "16:9", expected_ratio=8),
        C("seedance15sr_720p_16x9", "video", "seedance1.5-sr", "720p", "16:9", expected_ratio=10, expected_actual_resolution="720p"),
        C("seedance15sr_1080p_16x9", "video", "seedance1.5-sr", "1080p", "16:9", expected_ratio=10, expected_actual_resolution="1080p"),
        C("seedance2_480p_no_video_16x9", "video", "seedance2", "480p", "16:9", expected_ratio=23),
        C("seedance2_720p_no_video_16x9", "video", "seedance2", "720p", "16:9", expected_ratio=23),
        C("seedance2_1080p_no_video_16x9", "video", "seedance2", "1080p", "16:9", expected_ratio=25.5),
        C("seedance2sr_720p_16x9", "video", "seedance2-sr", "720p", "16:9", expected_ratio=24.5, expected_actual_resolution="720p"),
        C("seedance2sr_1080p_16x9", "video", "seedance2-sr", "1080p", "16:9", expected_ratio=24.5, expected_actual_resolution="1080p"),
        C("sd20fast_480p_16x9", "video", "sd2.0fast", "480p", "16:9", expected_ratio=18.5),
        C("sd20fast_720p_16x9", "video", "sd2.0fast", "720p", "16:9", expected_ratio=18.5),
        C("seedance20fastsr_720p_16x9", "video", "seedance2.0fast-sr", "720p", "16:9", expected_ratio=23.15, expected_actual_resolution="720p"),
        C("seedance20fastsr_1080p_16x9", "video", "seedance2.0fast-sr", "1080p", "16:9", expected_ratio=23.15, expected_actual_resolution="1080p"),
    ]
    for ratio in ["4:3", "1:1", "3:4", "9:16", "21:9", "adaptive"]:
        video_cases.append(C(f"ratio_seedance15_480p_{ratio.replace(':', 'x')}", "video", "seedance1.5", "480p", ratio, expected_ratio=8))

    if resume_path and mod.RESULTS_PATH.exists():
        results = json.loads(mod.RESULTS_PATH.read_text(encoding="utf-8"))
        results.setdefault("videos", [])
        results.setdefault("images", [])
        results.setdefault("assets", [])
        results["resumed_at"] = time.strftime("%Y-%m-%dT%H:%M:%S%z")
    else:
        results = {
            "base_url": mod.BASE_URL,
            "started_at": time.strftime("%Y-%m-%dT%H:%M:%S%z"),
            "matrix": "billing-refund-input-output-resolution",
            "results_path": str(mod.RESULTS_PATH),
            "videos": [],
            "images": [],
            "assets": [],
        }
    completed_video_names = {item.get("case", {}).get("name") for item in results.get("videos", [])}
    completed_image_names = {item.get("case", {}).get("name") for item in results.get("images", [])}

    reference_video_url = None
    for item in results.get("videos", []):
        if item.get("result_url"):
            reference_video_url = item["result_url"]
            break
    for idx, case in enumerate(video_cases, 1):
        if case.name in completed_video_names:
            print(f"[video {idx}/{len(video_cases)}] {case.name} skipped", flush=True)
            continue
        print(f"[video {idx}/{len(video_cases)}] {case.name}", flush=True)
        item = mod.submit_video(case, reference_video_url)
        results["videos"].append(item)
        if reference_video_url is None and item.get("result_url"):
            reference_video_url = item["result_url"]
        mod.RESULTS_PATH.write_text(json.dumps(results, ensure_ascii=False, indent=2), encoding="utf-8")
        print_summary(item)

    video_input_cases = [
        C("seedance2_480p_with_video_16x9", "video", "seedance2", "480p", "16:9", video_input=True, expected_ratio=15.5),
        C("seedance2_720p_with_video_16x9", "video", "seedance2", "720p", "16:9", video_input=True, expected_ratio=15.5),
        C("seedance2_1080p_with_video_16x9", "video", "seedance2", "1080p", "16:9", video_input=True, expected_ratio=17),
    ]
    for idx, case in enumerate(video_input_cases, 1):
        if case.name in completed_video_names:
            print(f"[video-input {idx}/{len(video_input_cases)}] {case.name} skipped", flush=True)
            continue
        print(f"[video-input {idx}/{len(video_input_cases)}] {case.name}", flush=True)
        item = mod.submit_video(case, reference_video_url)
        results["videos"].append(item)
        mod.RESULTS_PATH.write_text(json.dumps(results, ensure_ascii=False, indent=2), encoding="utf-8")
        print_summary(item)

    for idx, case in enumerate(mod.build_image_cases(), 1):
        if case.name in completed_image_names:
            print(f"[image {idx}/4] {case.name} skipped", flush=True)
            continue
        print(f"[image {idx}/4] {case.name}", flush=True)
        results["images"].append(mod.submit_image(case))
        mod.RESULTS_PATH.write_text(json.dumps(results, ensure_ascii=False, indent=2), encoding="utf-8")

    print("[assets] probing asset APIs", flush=True)
    results["assets"] = mod.probe_assets()
    results["finished_at"] = time.strftime("%Y-%m-%dT%H:%M:%S%z")
    mod.RESULTS_PATH.write_text(json.dumps(results, ensure_ascii=False, indent=2), encoding="utf-8")
    write_report(results, report_path)
    print("RESULTS=" + str(mod.RESULTS_PATH))
    print("REPORT=" + str(report_path))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
