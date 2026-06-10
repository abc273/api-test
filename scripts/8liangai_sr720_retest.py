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
        print("NEW_API_KEY is required", file=sys.stderr, flush=True)
        return 2

    mod = load_full_retest_module()
    stamp = time.strftime("%Y%m%d_%H%M%S")
    out_dir = ROOT / "tmp" / f"8liangai_sr720_normalized_{stamp}"
    mod.OUT_DIR = out_dir
    mod.DOWNLOAD_DIR = out_dir / "downloads"
    mod.REPORT_PATH = out_dir / "report.md"
    mod.RESULTS_PATH = out_dir / "results.json"
    mod.OUT_DIR.mkdir(parents=True, exist_ok=True)
    mod.DOWNLOAD_DIR.mkdir(parents=True, exist_ok=True)

    C = mod.Case
    cases = [
        C("seedance15sr_720p_16x9_normalized", "video", "seedance1.5-sr", "720p", "16:9", expected_ratio=10, expected_actual_resolution="720p"),
        C("seedance2sr_720p_16x9_normalized", "video", "seedance2-sr", "720p", "16:9", expected_ratio=24.5, expected_actual_resolution="720p"),
        C("seedance20fastsr_720p_16x9_normalized", "video", "seedance2.0fast-sr", "720p", "16:9", expected_ratio=23.15, expected_actual_resolution="720p"),
    ]
    selected = {x.strip() for x in os.getenv("SR720_CASES", "").split(",") if x.strip()}
    if selected:
        cases = [case for case in cases if case.name in selected]
    if not cases:
        print("no SR720 cases selected", file=sys.stderr, flush=True)
        return 2
    results = {
        "base_url": mod.BASE_URL,
        "started_at": time.strftime("%Y-%m-%dT%H:%M:%S%z"),
        "matrix": "sr-720-normalized-regression",
        "results_path": str(mod.RESULTS_PATH),
        "videos": [],
    }

    print("OUT_DIR=" + str(out_dir), flush=True)
    for idx, case in enumerate(cases, 1):
        print(f"[sr720 {idx}/{len(cases)}] {case.name}", flush=True)
        item = mod.submit_video(case, None)
        results["videos"].append(item)
        mod.RESULTS_PATH.write_text(json.dumps(results, ensure_ascii=False, indent=2), encoding="utf-8")

        cost = item.get("cost") or {}
        streams = (item.get("ffprobe") or {}).get("streams") or []
        dims = ""
        if streams:
            dims = f"{streams[0].get('width')}x{streams[0].get('height')}"
        pre = cost.get("pre_consumed_quota")
        actual = cost.get("actual_quota")
        delta = None if pre is None or actual is None else int(pre) - int(actual)
        print(
            "  ok={ok} dims={dims} expected={expected} ratio={ratio} pre={pre} actual={actual} refund_delta={delta} task={task}".format(
                ok=item.get("ok"),
                dims=dims,
                expected=item.get("expected_dims"),
                ratio=cost.get("model_ratio"),
                pre=pre,
                actual=actual,
                delta=delta,
                task=item.get("task_id"),
            ),
            flush=True,
        )

    results["finished_at"] = time.strftime("%Y-%m-%dT%H:%M:%S%z")
    mod.RESULTS_PATH.write_text(json.dumps(results, ensure_ascii=False, indent=2), encoding="utf-8")
    print("RESULTS=" + str(mod.RESULTS_PATH), flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
