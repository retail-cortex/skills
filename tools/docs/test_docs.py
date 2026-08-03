import unittest
from tools.docs import docs


class TestDocs(unittest.TestCase):
    def test_ensure_workspace_cwd(self) -> None:
        docs._ensure_workspace_cwd()
        self.assertTrue(True)


if __name__ == "__main__":
    unittest.main()
