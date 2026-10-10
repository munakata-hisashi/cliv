#!/usr/bin/env python3
"""Read-only comparison of native package-manager inventories and cliv JSON."""
import argparse
from collections import Counter
from datetime import datetime, timezone
import json
from pathlib import Path
import re
import shutil
import subprocess
import sys

SOURCES = ("brew", "mise", "npm", "uv")
COMMANDS = {
    "brew": ["brew", "info", "--json=v2", "--installed"],
    "mise": ["mise", "ls", "--json"],
    "npm": ["npm", "list", "-g", "--depth=0", "--json"],
    "uv": ["uv", "tool", "list", "--color", "never"],
}


def inventory(source, text, include_dependencies=False):
    """Independent oracle: do not import or call cliv's collectors."""
    result = set()
    if source == "uv":
        for line in text.splitlines():
            if not line.strip() or re.match(r"^\s*-\s", line):
                continue
            match = re.fullmatch(r"\s*(\S+)\s+v(\d\S*)(?:\s+.*)?", line)
            if not match:
                raise ValueError(f"unexpected uv line: {line!r}")
            result.add((source, match[1], match[2]))
        return result
    data = json.loads(text)
    if source == "brew":
        for formula in data["formulae"]:
            for installed in formula["installed"]:
                if include_dependencies or installed.get("installed_on_request") is True:
                    result.add((source, formula["name"], installed.get("version") or "unknown"))
    elif source == "npm":
        # An empty inventory can omit dependencies. Error metadata is not an inventory.
        if data.get("error") or data.get("problems"):
            raise ValueError("npm reported inventory errors")
        for name, package in data.get("dependencies", {}).items():
            result.add((source, name, package.get("version") or "unknown"))
    elif source == "mise":
        if isinstance(data, dict):
            records = [(name, item) for name, items in data.items()
                       for item in (items if isinstance(items, list) else [items])]
        elif isinstance(data, list):
            records = [(next((item[k] for k in ("name", "tool", "plugin", "short")
                              if item.get(k)), None), item) for item in data]
        else:
            raise ValueError("unexpected mise JSON shape")
        for name, item in records:
            if item.get("installed") is False:
                continue
            if not isinstance(name, str) or not name.strip():
                raise ValueError("mise tool name missing")
            version = next((str(item[k]).strip() for k in
                            ("version", "installed_version", "requested_version")
                            if item.get(k) is not None and str(item[k]).strip()), "unknown")
            result.add((source, name.strip(), version.split()[0]))
    return result


def actual_inventory(source, text):
    data = json.loads(text)
    if not isinstance(data, list):
        raise ValueError("cliv output is not a JSON array")
    rows = []
    for item in data:
        if item.get("source") != source:
            raise ValueError("cliv returned a different source")
        if not all(isinstance(item.get(k), str) and item[k] for k in ("tool", "version")):
            raise ValueError("cliv tool/version missing")
        rows.append((source, item["tool"], item["version"]))
    return Counter(rows)


def compare(expected, actual):
    matched = expected & actual.keys()
    return {
        "expected": len(expected), "matched": len(matched),
        "coverage": len(matched) / len(expected) if expected else None,
        "missing": sorted(expected - actual.keys()),
        "extra": sorted(actual.keys() - expected),
        "duplicates": [{"entry": list(key), "count": count}
                       for key, count in sorted(actual.items()) if count > 1],
    }


def run(command, timeout):
    try:
        proc = subprocess.run(command, capture_output=True, text=True, timeout=timeout)
        return {"command": command, "returncode": proc.returncode,
                "stdout": proc.stdout, "stderr": proc.stderr}
    except (OSError, subprocess.TimeoutExpired) as exc:
        return {"command": command, "returncode": None, "stdout": "", "stderr": str(exc)}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cliv", required=True, help="path to the cliv binary to verify")
    parser.add_argument("--sources", nargs="+", choices=SOURCES, default=list(SOURCES))
    parser.add_argument("--all", action="store_true", help="include Homebrew dependencies")
    parser.add_argument("--output-dir", type=Path, help="save raw evidence and report (must not exist)")
    parser.add_argument("--timeout", type=float, default=60, help="timeout per command in seconds")
    args = parser.parse_args()
    if args.timeout <= 0:
        parser.error("--timeout must be positive")
    cliv = shutil.which(args.cliv)
    if not cliv:
        parser.error(f"cliv executable not found: {args.cliv}")
    if args.output_dir:
        args.output_dir.mkdir(parents=True, exist_ok=False)
    report = {"timestamp": datetime.now(timezone.utc).isoformat(),
              "cliv": str(Path(cliv).resolve()), "all": args.all,
              "cliv_version": run([cliv, "--version"], args.timeout), "sources": {}}
    exit_code = 0
    print("SOURCE  EXPECTED  MATCHED  MISSING  EXTRA  DUPLICATES  COVERAGE  STATUS")
    for source in dict.fromkeys(args.sources):
        record = {}
        report["sources"][source] = record
        if not shutil.which(source):
            record["status"] = "SKIP"
            print(f"{source:<7} {'-':>8} {'-':>8} {'-':>8} {'-':>6} {'-':>11} {'N/A':>9}  SKIP (not installed)")
            continue
        record["manager_version"] = run([source, "--version"], args.timeout)
        record["native"] = run(COMMANDS[source], args.timeout)
        command = [cliv, "--source", source, "--json"]
        if args.all:
            command.append("--all")
        record["cliv"] = run(command, args.timeout)
        try:
            for name in ("native", "cliv"):
                if record[name]["returncode"] != 0:
                    raise ValueError(f"{name} command failed: {record[name]['stderr'].strip()}")
            # Older cliv binaries may exit zero after a collector failure.
            if record["cliv"]["stderr"].strip():
                raise ValueError(f"cliv reported: {record['cliv']['stderr'].strip()}")
            expected = inventory(source, record["native"]["stdout"], args.all)
            actual = actual_inventory(source, record["cliv"]["stdout"])
            record["expected_inventory"] = sorted(expected)
            result = compare(expected, actual)
            record["comparison"] = result
            different = any(result[k] for k in ("missing", "extra", "duplicates"))
            record["status"] = "DIFF" if different else "PASS"
            if different:
                exit_code = max(exit_code, 1)
            coverage = f"{result['coverage']:.1%}" if result["coverage"] is not None else "N/A"
            print(f"{source:<7} {result['expected']:>8} {result['matched']:>8} {len(result['missing']):>8} "
                  f"{len(result['extra']):>6} {len(result['duplicates']):>11} {coverage:>9}  {record['status']}")
            for label in ("missing", "extra", "duplicates"):
                for entry in result[label]:
                    print(f"  {label.upper()}: {entry}")
        except (ValueError, KeyError, TypeError, AttributeError) as exc:
            record.update(status="ERROR", error=str(exc))
            exit_code = 2
            print(f"{source:<7} ERROR: {exc}")
    if all(r["status"] == "SKIP" for r in report["sources"].values()):
        exit_code = 2
        print("No sources measured.")
    report["exit_code"] = exit_code
    if args.output_dir:
        (args.output_dir / "report.json").write_text(json.dumps(report, indent=2) + "\n")
        print(f"Evidence: {args.output_dir / 'report.json'}")
    return exit_code


if __name__ == "__main__":
    sys.exit(main())
