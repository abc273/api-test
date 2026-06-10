#!/usr/bin/env python3
from __future__ import annotations

import importlib.util
import json
import os
import sys
import time
from pathlib import Path


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


def main() -> int:
    if not os.getenv("NEW_API_KEY"):
        print("NEW_API_KEY is required", file=sys.stderr)
        return 2

    mod = load_full_retest_module()
    stamp = time.strftime("%Y%m%d_%H%M%S")
    out_dir = ROOT / "tmp" / f"8liangai_targeted_regression_{stamp}"
    mod.OUT_DIR = out_dir
    mod.DOWNLOAD_DIR = out_dir / "downloads"
    mod.REPORT_PATH = out_dir / "report.md"
    mod.RESULTS_PATH = out_dir / "results.json"
    mod.OUT_DIR.mkdir(parents=True, exist_ok=True)
    mod.DOWNLOAD_DIR.mkdir(parents=True, exist_ok=True)

    C = mod.Case

    # Main model/resolution coverage. This covers every public video model and
    # every supported user-facing resolution without exploding into all ratios.
    video_cases = [
        C("main_seedance15_480p_16x9", "video", "seedance1.5", "480p", "16:9", expected_ratio=8),
        C("main_seedance15_720p_16x9", "video", "seedance1.5", "720p", "16:9", expected_ratio=8),
        C("main_seedance15_1080p_16x9", "video", "seedance1.5", "1080p", "16:9", expected_ratio=8),
        C("main_seedance15sr_720p_16x9", "video", "seedance1.5-sr", "720p", "16:9", expected_ratio=10, expected_actual_resolution="720p"),
        C("main_seedance15sr_1080p_16x9", "video", "seedance1.5-sr", "1080p", "16:9", expected_ratio=10, expected_actual_resolution="1080p"),
        C("main_seedance2_480p_no_video_16x9", "video", "seedance2", "480p", "16:9", expected_ratio=23),
        C("main_seedance2_720p_no_video_16x9", "video", "seedance2", "720p", "16:9", expected_ratio=23),
        C("main_seedance2_1080p_no_video_16x9", "video", "seedance2", "1080p", "16:9", expected_ratio=25.5),
        C("main_seedance2sr_720p_16x9", "video", "seedance2-sr", "720p", "16:9", expected_ratio=24.5, expected_actual_resolution="720p"),
        C("main_seedance2sr_1080p_16x9", "video", "seedance2-sr", "1080p", "16:9", expected_ratio=24.5, expected_actual_resolution="1080p"),
        C("main_sd20fast_480p_16x9", "video", "sd2.0fast", "480p", "16:9", expected_ratio=18.5),
        C("main_sd20fast_720p_16x9", "video", "sd2.0fast", "720p", "16:9", expected_ratio=18.5),
        C("main_seedance20fastsr_720p_16x9", "video", "seedance2.0fast-sr", "720p", "16:9", expected_ratio=23.15, expected_actual_resolution="720p"),
        C("main_seedance20fastsr_1080p_16x9", "video", "seedance2.0fast-sr", "1080p", "16:9", expected_ratio=23.15, expected_actual_resolution="1080p"),
    ]

    # Ratio coverage is sampled on the lowest-cost base model and 480p.
    # The gateway payload conversion and upstream ratio support are shared, so
    # this detects ratio regressions without multiplying cost by every model.
    for ratio in ["4:3", "1:1", "3:4", "9:16", "21:9", "adaptive"]:
        video_cases.append(
            C(
                f"ratio_seedance15_480p_{ratio.replace(':', 'x')}",
                "video",
                "seedance1.5",
                "480p",
                ratio,
                expected_ratio=8,
            )
        )

    results = {
        "base_url": mod.BASE_URL,
        "started_at": time.strftime("%Y-%m-%dT%H:%M:%S%z"),
        "matrix": "targeted-cost-controlled",
        "videos": [],
        "images": [],
        "assets": [],
    }

    reference_video_url = None
    for idx, case in enumerate(video_cases, 1):
        print(f"[video {idx}/{len(video_cases)}] {case.name}", flush=True)
        item = mod.submit_video(case, reference_video_url)
        results["videos"].append(item)
        if reference_video_url is None and item.get("result_url"):
            reference_video_url = item["result_url"]
        mod.RESULTS_PATH.write_text(json.dumps(results, ensure_ascii=False, indent=2), encoding="utf-8")
        print_summary(item)

    # Video-input tier coverage requires a valid reference video generated above.
    video_input_cases = [
        C("tier_seedance2_480p_with_video_16x9", "video", "seedance2", "480p", "16:9", video_input=True, expected_ratio=15.5),
        C("tier_seedance2_720p_with_video_16x9", "video", "seedance2", "720p", "16:9", video_input=True, expected_ratio=15.5),
        C("tier_seedance2_1080p_with_video_16x9", "video", "seedance2", "1080p", "16:9", video_input=True, expected_ratio=17),
    ]
    for idx, case in enumerate(video_input_cases, 1):
        print(f"[video-input {idx}/{len(video_input_cases)}] {case.name}", flush=True)
        item = mod.submit_video(case, reference_video_url)
        results["videos"].append(item)
        mod.RESULTS_PATH.write_text(json.dumps(results, ensure_ascii=False, indent=2), encoding="utf-8")
        print_summary(item)

    image_cases = mod.build_image_cases()
    for idx, case in enumerate(image_cases, 1):
        print(f"[image {idx}/{len(image_cases)}] {case.name}", flush=True)
        item = mod.submit_image(case)
        results["images"].append(item)
        mod.RESULTS_PATH.write_text(json.dumps(results, ensure_ascii=False, indent=2), encoding="utf-8")

    print("[assets] probing asset APIs", flush=True)
    results["assets"] = mod.probe_assets()
    results["finished_at"] = time.strftime("%Y-%m-%dT%H:%M:%S%z")
    mod.RESULTS_PATH.write_text(json.dumps(results, ensure_ascii=False, indent=2), encoding="utf-8")
    print("RESULTS=" + str(mod.RESULTS_PATH))
    return 0


def print_summary(item: dict) -> None:
    case = item.get("case", {})
    streams = (item.get("ffprobe") or {}).get("streams") or []
    dims = f"{streams[0].get('width')}x{streams[0].get('height')}" if streams else ""
    cost = item.get("cost") or {}
    print(
        "  ok={ok} dims={dims} expected={expected} ratio={ratio} cny={cny} task={task}".format(
            ok=item.get("ok"),
            dims=dims,
            expected=item.get("expected_dims"),
            ratio=cost.get("model_ratio"),
            cny=cost.get("actual_cny"),
            task=item.get("task_id"),
        ),
        flush=True,
    )


if __name__ == "__main__":
    raise SystemExit(main())
