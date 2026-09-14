import unittest
from pathlib import Path

from miplang_runtime import load_ir


class RuntimeTest(unittest.TestCase):
    def test_loads_shared_ir(self):
        path = Path(__file__).parents[2] / "testdata" / "transport.ir.json"
        model = load_ir(path)
        self.assertEqual(model.schema_version, "miplang.ir/v1alpha2")
        self.assertEqual(model.name, "anonymous")
        self.assertEqual(model.sets, ("TRIPS",))


if __name__ == "__main__":
    unittest.main()
