import unittest
import json
from pathlib import Path
from collections import Counter

from check_collection import actual_inventory, compare, inventory


class CollectionTests(unittest.TestCase):
    def test_reviewed_fixtures(self):
        fixtures = Path(__file__).resolve().parents[1] / "internal" / "collector" / "testdata"
        for source, all_mode in [("brew", False), ("brew", True), ("mise", False), ("npm", False), ("uv", False)]:
            with self.subTest(source=source, all=all_mode):
                extension = "txt" if source == "uv" else "json"
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

    def test_actual(self):
        self.assertEqual(actual_inventory("uv", '[{"tool":"httpie","version":"3","source":"uv"}]'), Counter({("uv", "httpie", "3"): 1}))
        with self.assertRaises(ValueError):
            actual_inventory("uv", '[{"tool":"httpie","version":"3","source":"mise"}]')


if __name__ == "__main__":
    unittest.main()
