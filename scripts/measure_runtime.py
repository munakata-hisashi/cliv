#!/usr/bin/env python3
"""Measure cliv wall-clock runtime without timing inventory verification."""
import argparse
from datetime import datetime, timezone
import hashlib
import json
import math
from pathlib import Path
import platform
import shutil
import statistics
import subprocess
import sys
import time

TARGETS = ("total", "brew", "mise", "npm", "uv", "manual")


def timed_run(command, timeout):
    start = time.perf_counter()
    try:
        proc = subprocess.run(command, capture_output=True, text=True, timeout=timeout)
        return {"command": command, "seconds": time.perf_counter() - start,
                "returncode": proc.returncode, "stdout": proc.stdout, "stderr": proc.stderr}
    except (OSError, subprocess.TimeoutExpired) as exc:
        return {"command": command, "seconds": time.perf_counter() - start,
                "returncode": None, "stdout": "", "stderr": str(exc)}


def sample_error(sample):
    if sample["returncode"] != 0:
        return f"command failed (exit={sample['returncode']}): {sample['stderr'].strip()}"
    # Older cliv binaries may exit zero after a collector failure.
    if sample["stderr"].strip():
        return f"cliv reported: {sample['stderr'].strip()}"
    return None


def summarize(seconds):
    return {"count": len(seconds), "median_seconds": statistics.median(seconds),
            "min_seconds": min(seconds), "max_seconds": max(seconds)}


def measure(command, runs, timeout):
    result = {"command": command, "status": "PASS", "first": timed_run(command, timeout),
              "samples": [], "summary": None}
    error = sample_error(result["first"])
    if not error:
        for _ in range(runs):
            sample = timed_run(command, timeout)
            result["samples"].append(sample)
            error = sample_error(sample)
            if error:
                break
    if error:
        result.update(status="ERROR", error=error)
    else:
        result["summary"] = summarize([s["seconds"] for s in result["samples"]])
    return result


def file_digest(path):
    try:
        return hashlib.sha256(path.read_bytes()).hexdigest()
    except OSError:
        return None


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cliv", required=True, help="path to the compiled cliv binary")
    parser.add_argument("--targets", nargs="+", choices=TARGETS, default=list(TARGETS))
    parser.add_argument("--runs", type=int, default=5, help="runs after the first run (default: 5)")
    parser.add_argument("--timeout", type=float, default=60, help="timeout per command in seconds")
    parser.add_argument("--all", action="store_true", help="include Homebrew dependencies")
    parser.add_argument("--json", action="store_true", help="benchmark JSON instead of plain table output")
    parser.add_argument("--output-dir", type=Path, help="save report.json in a new directory")
    args = parser.parse_args()
    if args.runs < 1:
        parser.error("--runs must be at least 1")
    if not math.isfinite(args.timeout) or args.timeout <= 0:
        parser.error("--timeout must be finite and positive")
    cliv = shutil.which(args.cliv)
    if not cliv:
        parser.error(f"cliv executable not found: {args.cliv}")
    cliv = str(Path(cliv).resolve())
    if args.output_dir:
        args.output_dir.mkdir(parents=True, exist_ok=False)
    config_path = Path.home() / ".config" / "cliv" / "config.toml"
    report = {
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "cliv": cliv, "cliv_sha256": file_digest(Path(cliv)),
        "platform": platform.platform(), "python": platform.python_version(),
        "cwd": str(Path.cwd()), "config_sha256": file_digest(config_path),
        "runs_after_first": args.runs, "timeout_seconds": args.timeout,
        "all": args.all, "format": "json" if args.json else "table",
        "stdout": "pipe", "targets": {},
    }
    exit_code = 0
    print("TARGET     FIRST(ms)  MEDIAN(ms)     MIN(ms)     MAX(ms)  RUNS  STATUS")
    for target in dict.fromkeys(args.targets):
        if target not in ("total", "manual") and not shutil.which(target):
            report["targets"][target] = {"status": "SKIP", "reason": "manager not installed"}
            print(f"{target:<10} SKIP (manager not installed)")
            continue
        command = [cliv, "--color", "never"]
        if target != "total":
            command.extend(["--source", target])
        if args.all:
            command.append("--all")
        if args.json:
            command.append("--json")
        result = measure(command, args.runs, args.timeout)
        report["targets"][target] = result
        if result["status"] == "ERROR":
            exit_code = 2
            print(f"{target:<10} ERROR: {result['error']}")
            continue
        summary = result["summary"]
        print(f"{target:<10} {result['first']['seconds'] * 1000:>9.1f} "
              f"{summary['median_seconds'] * 1000:>11.1f} "
              f"{summary['min_seconds'] * 1000:>11.1f} "
              f"{summary['max_seconds'] * 1000:>11.1f} {summary['count']:>5}  PASS")
        # An empty source can be intentional, but should not look like a fast collector.
        first_output = result["first"]["stdout"].strip()
        if target != "total" and first_output in ("", "[]"):
            result["note"] = "empty inventory; check collector configuration before comparing timings"
            print(f"  NOTE: {result['note']}")
    if all(r["status"] == "SKIP" for r in report["targets"].values()):
        exit_code = 2
        print("No targets measured.")
    # Collect version metadata AFTER timed runs so --version does not warm up managers first.
    report["cliv_version"] = timed_run([cliv, "--version"], args.timeout)
    report["manager_versions"] = {
        source: timed_run([source, "--version"], args.timeout)
        for source in ("brew", "mise", "npm", "uv") if shutil.which(source)
    }
    report["exit_code"] = exit_code
    if args.output_dir:
        (args.output_dir / "report.json").write_text(json.dumps(report, indent=2) + "\n")
        print(f"Evidence: {args.output_dir / 'report.json'}")
    return exit_code


if __name__ == "__main__":
    sys.exit(main())
