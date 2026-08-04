import unittest
from main import run


class TestPythonClientExample(unittest.TestCase):

    def test_run(self):
        res = run()
        self.assertEqual(res["server_url"], "http://localhost:8080")
        self.assertEqual(res["api_key"], "python-secret-key-67890")
        self.assertEqual(res["scheme"], "github")
        self.assertEqual(res["target"], "google/skills@main")



if __name__ == "__main__":
    unittest.main()
