import subprocess
import unittest
from unittest.mock import patch

from measure_runtime import measure, sample_error, summarize, timed_run


def sample(seconds, returncode=0, stderr=""):
    return {"command": ["cliv"], "seconds": seconds, "returncode": returncode,
            "stdout": "uv\nTOOL VERSION\ntool 1\n", "stderr": stderr}


class RuntimeTests(unittest.TestCase):
    def test_summary(self):
        self.assertEqual(summarize([0.3, 0.1, 0.2]), {
            "count": 3, "median_seconds": 0.2, "min_seconds": 0.1, "max_seconds": 0.3,
        })

    @patch("measure_runtime.timed_run")
    def test_first_excluded_from_summary(self, run):
        run.side_effect = [sample(9), sample(0.3), sample(0.1), sample(0.2)]
        result = measure(["cliv"], 3, 60)
        self.assertEqual(result["status"], "PASS")
        self.assertEqual(result["first"]["seconds"], 9)
        self.assertEqual(result["summary"]["median_seconds"], 0.2)
        self.assertEqual(run.call_count, 4)

    @patch("measure_runtime.timed_run")
    def test_failed_first_not_benchmarked(self, run):
        run.return_value = sample(0.1, returncode=1, stderr="failed")
        result = measure(["cliv"], 5, 60)
        self.assertEqual(result["status"], "ERROR")
        self.assertIsNone(result["summary"])
        self.assertEqual(result["samples"], [])
        run.assert_called_once()

    @patch("measure_runtime.timed_run")
    def test_collector_failure_invalidates_summary(self, run):
        run.side_effect = [sample(1), sample(0.1), sample(0.01, stderr="brew: failed")]
        result = measure(["cliv"], 5, 60)
        self.assertEqual(result["status"], "ERROR")
        self.assertIsNone(result["summary"])
        self.assertEqual(len(result["samples"]), 2)
        self.assertEqual(run.call_count, 3)

    def test_whitespace_stderr(self):
        self.assertIsNone(sample_error(sample(1, stderr="\n")))

    @patch("measure_runtime.time.perf_counter", side_effect=[10.0, 10.25])
    @patch("measure_runtime.subprocess.run")
    def test_wall_clock(self, run, clock):
        run.return_value = subprocess.CompletedProcess(["cliv"], 0, "output", "")
        result = timed_run(["cliv"], 60)
        self.assertEqual(result["seconds"], 0.25)
        self.assertEqual(result["stdout"], "output")
        run.assert_called_once_with(["cliv"], capture_output=True, text=True, timeout=60)

    @patch("measure_runtime.subprocess.run", side_effect=subprocess.TimeoutExpired(["cliv"], 1))
    def test_timeout(self, run):
        result = timed_run(["cliv"], 1)
        self.assertIsNone(result["returncode"])
        self.assertIn("timed out", result["stderr"])
        self.assertIsNotNone(sample_error(result))

    @patch("measure_runtime.subprocess.run", side_effect=OSError("not found"))
    def test_spawn_failure(self, run):
        result = timed_run(["cliv"], 1)
        self.assertIsNone(result["returncode"])
        self.assertIn("not found", result["stderr"])


if __name__ == "__main__":
    unittest.main()
