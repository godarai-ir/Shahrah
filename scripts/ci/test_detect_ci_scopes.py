import importlib.util
import sys
import unittest
from pathlib import Path

MODULE_PATH = Path(__file__).with_name("detect_ci_scopes.py")
SPEC = importlib.util.spec_from_file_location("detect_ci_scopes", MODULE_PATH)
assert SPEC and SPEC.loader
DETECTOR = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = DETECTOR
SPEC.loader.exec_module(DETECTOR)


class DetectCiScopesTests(unittest.TestCase):
    def test_go_change_only_selects_go(self):
        scopes = DETECTOR.detect_scopes(
            ["services/control-plane/internal/server/server.go"]
        )

        self.assertEqual(
            scopes,
            {
                "go": True,
                "tenant": False,
                "model": False,
                "javascript": False,
                "tenant_dom": False,
            },
        )

    def test_tenant_change_selects_node_and_dom_checks(self):
        scopes = DETECTOR.detect_scopes(
            ["apps/tenant-admin/components/app-ui/Button.tsx"]
        )

        self.assertFalse(scopes["go"])
        self.assertTrue(scopes["tenant"])
        self.assertFalse(scopes["model"])
        self.assertTrue(scopes["javascript"])
        self.assertTrue(scopes["tenant_dom"])

    def test_model_change_does_not_select_tenant_dom(self):
        scopes = DETECTOR.detect_scopes(["apps/model-discovery/src/main.ts"])

        self.assertFalse(scopes["go"])
        self.assertFalse(scopes["tenant"])
        self.assertTrue(scopes["model"])
        self.assertTrue(scopes["javascript"])
        self.assertFalse(scopes["tenant_dom"])

    def test_lockfile_change_selects_both_javascript_projects(self):
        scopes = DETECTOR.detect_scopes(["pnpm-lock.yaml"])

        self.assertFalse(scopes["go"])
        self.assertTrue(scopes["tenant"])
        self.assertTrue(scopes["model"])
        self.assertTrue(scopes["javascript"])
        self.assertTrue(scopes["tenant_dom"])

    def test_unrelated_docs_change_skips_expensive_checks(self):
        scopes = DETECTOR.detect_scopes(["plans/example.md"])

        self.assertEqual(scopes, DETECTOR.all_scopes(False))

    def test_ci_workflow_change_and_missing_base_select_everything(self):
        self.assertEqual(
            DETECTOR.detect_scopes([".github/workflows/ci.yml"]),
            DETECTOR.all_scopes(True),
        )
        self.assertEqual(
            DETECTOR.detect_scopes([], force_all=True),
            DETECTOR.all_scopes(True),
        )

    def test_zero_sha_is_not_a_usable_base(self):
        self.assertTrue(DETECTOR.is_missing_base("0" * 40))
        self.assertTrue(DETECTOR.is_missing_base(""))
        self.assertFalse(DETECTOR.is_missing_base("abc123"))


if __name__ == "__main__":
    unittest.main()
