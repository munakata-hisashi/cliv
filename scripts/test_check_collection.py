import json
import os
import unittest
from pathlib import Path
from collections import Counter

from check_collection import actual_inventory, compare, go_bin_dir, inventory


class CollectionTests(unittest.TestCase):
    def test_reviewed_fixtures(self):
        fixtures = Path(__file__).resolve().parents[1] / "internal" / "collector" / "testdata"
        for source, all_mode in [("brew", False), ("brew", True), ("mise", False), ("npm", False), ("uv", False), ("cargo", False), ("go", False)]:
            with self.subTest(source=source, all=all_mode):
                extension = "txt" if source in ("uv", "cargo", "go") else "json"
                text = (fixtures / f"{source}.input.{extension}").read_text()
                name = "brew-all" if all_mode else source
                expected = json.loads((fixtures / f"{name}.expected.json").read_text())
                keys = {(item["source"], item["tool"], item["version"]) for item in expected}
                self.assertEqual(inventory(source, text, all_mode), keys)

    def test_comparison(self):
        expected = {("mise", "pi", "1"), ("mise", "codex", "2")}
        actual = Counter({("mise", "codex", "2"): 2, ("mise", "extra", "3"): 1})
        result = compare(expected, actual)
        self.assertEqual(result["coverage"], 0.5)
        self.assertEqual(result["missing"], [("mise", "pi", "1")])
        self.assertEqual(result["extra"], [("mise", "extra", "3")])
        self.assertEqual(result["duplicates"][0]["count"], 2)

    def test_empty(self):
        result = compare(set(), Counter())
        self.assertIsNone(result["coverage"])
        self.assertEqual(result["missing"], [])

    def test_brew(self):
        text = '{"formulae":[{"name":"cmake","installed":[{"version":"4","installed_on_request":true}]},{"name":"lib","installed":[{"version":"1","installed_on_request":false}]}]}'
        self.assertEqual(inventory("brew", text), {("brew", "cmake", "4")})
        self.assertEqual(len(inventory("brew", text, True)), 2)

    def test_mise(self):
        text = '{"npm:pi":[{"version":"1","installed":true,"active":false},{"version":"2","installed":true},{"version":"3","installed":false}]}'
        self.assertEqual(inventory("mise", text), {("mise", "npm:pi", "1"), ("mise", "npm:pi", "2")})

    def test_npm(self):
        self.assertEqual(inventory("npm", '{"dependencies":{"library":{"version":"1"}}}'), {("npm", "library", "1")})
        with self.assertRaises(ValueError):
            inventory("npm", '{"error":{"code":"FAIL"}}')

    def test_uv(self):
        self.assertEqual(inventory("uv", "httpie v3.2.4\n- http\n- https\n"), {("uv", "httpie", "3.2.4")})
        with self.assertRaises(ValueError):
            inventory("uv", "unrecognized output")

    def test_cargo(self):
        text = "ripgrep v14.1.1:\n    rg\nlocal-tool v0.1.0 (/home/user/src/local-tool):\n    local-tool\n"
        self.assertEqual(inventory("cargo", text), {("cargo", "ripgrep", "14.1.1"), ("cargo", "local-tool", "0.1.0")})
        self.assertEqual(inventory("cargo", ""), set())
        for line in ("ripgrep 14.1.1:", "ripgrep v14.1.1", "garbage"):
            with self.assertRaises(ValueError):
                inventory("cargo", line)

    def test_go(self):
        text = ("/home/user/go/bin/gopls: go1.23.2\n"
                "\tpath\tgolang.org/x/tools/gopls\n"
                "\tmod\tgolang.org/x/tools/gopls\tv0.16.2\th1:abc=\n"
                "/home/user/go/bin/mytool: go1.23.2\n"
                "\tpath\texample.com/mytool\n"
                "\tmod\texample.com/mytool\t(devel)\n"
                "/tmp/go: go1.23.2\n"
                "\tpath\tcmd/go\n")
        self.assertEqual(inventory("go", text), {
            ("go", "golang.org/x/tools/gopls", "0.16.2"),
            ("go", "example.com/mytool", "(devel)"),
            ("go", "cmd/go", "unknown"),
        })
        self.assertEqual(inventory("go", ""), set())
        for line in ("garbage", "\tpath\tx"):
            with self.assertRaises(ValueError):
                inventory("go", line)

    def test_go_bin_dir(self):
        sep = os.pathsep
        self.assertEqual(go_bin_dir("/custom/bin\n/home/u/go\n"), "/custom/bin")
        self.assertEqual(go_bin_dir("\n/home/u/go" + sep + "/other\n"), str(Path("/home/u/go") / "bin"))
        self.assertEqual(go_bin_dir("\n\n"), "")

    def test_actual(self):
        self.assertEqual(actual_inventory("uv", '[{"tool":"httpie","version":"3","source":"uv"}]'), Counter({("uv", "httpie", "3"): 1}))
        with self.assertRaises(ValueError):
            actual_inventory("uv", '[{"tool":"httpie","version":"3","source":"mise"}]')


if __name__ == "__main__":
    unittest.main()
