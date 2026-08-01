import importlib.util
import sys
import unittest
from pathlib import Path

MODULE_PATH = Path(__file__).with_name("detect_changed_services.py")
SPEC = importlib.util.spec_from_file_location("detect_changed_services", MODULE_PATH)
assert SPEC and SPEC.loader
DETECTOR = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = DETECTOR
SPEC.loader.exec_module(DETECTOR)


class SelectChangedServicesTests(unittest.TestCase):
    def setUp(self):
        self.services = [
            {
                "id": "service-a",
                "paths": ["services/a/", "shared/a.json"],
            },
            {
                "id": "service-b",
                "paths": ["services/b/", "shared/b.json"],
            },
        ]

    def test_uses_deployed_base_and_catches_changes_skipped_by_failed_ci(self):
        bases = {"service-a": "deployed-a", "service-b": "deployed-b"}
        changed = {
            "deployed-a": ["services/a/change-before-failed-ci.go"],
            "deployed-b": [],
        }

        selected, deployed_bases, warnings = DETECTOR.select_changed_services(
            self.services,
            "",
            lambda service: ("v1.0.0", bases[service["id"]]),
            lambda base: changed[base],
        )

        self.assertEqual(selected, ["service-a"])
        self.assertEqual(deployed_bases["service-a"], "deployed-a")
        self.assertEqual(warnings, [])

    def test_service_specific_shared_change_only_selects_dependents(self):
        selected, _, _ = DETECTOR.select_changed_services(
            self.services,
            "",
            lambda service: ("v1.0.0", service["id"]),
            lambda _base: ["shared/a.json"],
        )

        self.assertEqual(selected, ["service-a"])

    def test_node_lock_change_does_not_select_go_images(self):
        selected, _, _ = DETECTOR.select_changed_services(
            DETECTOR.SERVICES,
            "",
            lambda service: ("v1.0.0", service["id"]),
            lambda _base: ["pnpm-lock.yaml"],
        )

        self.assertEqual(
            selected,
            [
                "model-discovery",
                "model-discovery-migrate",
                "model-discovery-web",
                "support-admin",
                "support-admin-migrate",
                "docs",
                "website",
                "blog",
                "tenant-admin",
            ],
        )

    def test_contract_change_selects_only_dependent_go_images(self):
        selected, _, _ = DETECTOR.select_changed_services(
            DETECTOR.SERVICES,
            "",
            lambda service: ("v1.0.0", service["id"]),
            lambda _base: ["packages/contracts/platform/types.go"],
        )

        self.assertEqual(
            selected,
            ["control-plane", "silk-road", "tunnel-client", "tunnel-server"],
        )

    def test_transport_change_also_rebuilds_control_plane(self):
        selected, _, _ = DETECTOR.select_changed_services(
            DETECTOR.SERVICES,
            "",
            lambda service: ("v1.0.0", service["id"]),
            lambda _base: ["services/silk-road/pkg/server/server.go"],
        )

        self.assertEqual(selected, ["control-plane", "silk-road"])

    def test_turbo_configuration_does_not_rebuild_runtime_images(self):
        selected, _, _ = DETECTOR.select_changed_services(
            DETECTOR.SERVICES,
            "",
            lambda service: ("v1.0.0", service["id"]),
            lambda _base: ["turbo.json"],
        )

        self.assertEqual(selected, [])

    def test_missing_deployed_tag_selects_service_conservatively(self):
        def resolve(service):
            if service["id"] == "service-a":
                raise ValueError("missing tag")
            return "v1.0.0", "deployed-b"

        selected, _, warnings = DETECTOR.select_changed_services(
            self.services,
            "",
            resolve,
            lambda _base: [],
        )

        self.assertEqual(selected, ["service-a"])
        self.assertTrue(any("selecting conservatively" in item for item in warnings))

    def test_manual_selection_is_validated(self):
        with self.assertRaisesRegex(ValueError, "Unknown service id"):
            DETECTOR.select_changed_services(
                self.services,
                "unknown",
                lambda _service: ("v1.0.0", "base"),
                lambda _base: [],
            )


if __name__ == "__main__":
    unittest.main()
