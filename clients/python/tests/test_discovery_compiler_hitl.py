# Copyright 2026 Ryan McGuinness
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Unit tests for SkillCompiler, SkillDiscoveryEngine, and HITLEngine in clients/python."""

import unittest
from loader.compiler import SkillCompiler
from loader.discovery import SkillDiscoveryEngine, TFIDFVectorIndex
from loader.hitl import HITLEngine
from loader.loader import SkillRegistry
from loader.types import CompiledSkillReference, HITLPolicyTier, SkillDefinition


class TestDiscoveryCompilerHITL(unittest.TestCase):

    def setUp(self):
        self.sample_python_skill = SkillDefinition(
            name="python-core",
            description="Elite Python 3.13 backend SDLC using uv wrapped in Bazel.",
            instructions="Enforces pytest TDD with paired positive/negative tests, HTTP 429 rate limit backoff.",
            category="python",
            tags=["python", "uv", "pytest", "typing"],
            trigger_phrases=["write Python core code", "uv Python project setup"],
            execution_hints={
                "requires_human_approval": False,
                "hitl_tier": "TIER_1_AUTO_READ",
            },
            path="/tmp/skills/python-core",
        )
        self.sample_database_skill = SkillDefinition(
            name="database-postgres",
            description="Enterprise PostgreSQL schema migration and query optimization.",
            instructions="Execute SQL migration scripts and manage database connections safely.",
            category="database",
            tags=["database", "postgres", "sql", "migration"],
            trigger_phrases=["query postgres database", "migrate database schema"],
            execution_hints={
                "requires_human_approval": True,
                "hitl_tier": "TIER_3_MANDATORY_APPROVAL",
            },
            path="/tmp/skills/database-postgres",
        )

    def test_compiler_strict_and_permissive_schemas(self):
        # Test default strict compiler
        strict_compiler = SkillCompiler(strict_schemas=True)
        ref_strict = strict_compiler.compile(self.sample_python_skill)

        self.assertEqual(ref_strict.name, "python-core")
        self.assertTrue(ref_strict.strict_schema)
        self.assertFalse(ref_strict.json_schema["additionalProperties"])
        self.assertLess(ref_strict.estimated_tokens, 150)
        self.assertEqual(len(ref_strict.sha256_hash), 64)

        # Test permissive compiler with allowed additional properties
        permissive_compiler = SkillCompiler(
            strict_schemas=False, allow_additional_properties=["custom_context", "debug_mode"]
        )
        ref_permissive = permissive_compiler.compile(self.sample_python_skill)

        self.assertFalse(ref_permissive.strict_schema)
        self.assertEqual(ref_permissive.allowed_properties, ["custom_context", "debug_mode"])
        self.assertIn("additionalProperties", ref_permissive.json_schema)

    def test_discovery_engine_tfidf_search(self):
        engine = SkillDiscoveryEngine()
        engine.register_skill(self.sample_python_skill)
        engine.register_skill(self.sample_database_skill)
        engine.build_index()

        # Search for database query intent
        db_result = engine.search_skills("query postgres database", top_k=2)
        self.assertGreaterEqual(db_result.total_found, 1)
        top_db_match = db_result.matches[0]
        self.assertEqual(top_db_match.name, "database-postgres")

        # Search for Python SDLC intent
        py_result = engine.search_skills("write Python backend service", top_k=2)
        self.assertGreaterEqual(py_result.total_found, 1)
        top_py_match = py_result.matches[0]
        self.assertEqual(top_py_match.name, "python-core")

    def test_hitl_engine_policy_gates_and_bypass(self):
        hitl = HITLEngine(skip_hitl=False)
        compiler = SkillCompiler()

        ref_read = compiler.compile(self.sample_python_skill)
        ref_write = compiler.compile(self.sample_database_skill)

        # Tier 1 Auto Read
        gate_read = hitl.evaluate_gate(self.sample_python_skill, ref=ref_read)
        self.assertTrue(gate_read.approved)
        self.assertFalse(gate_read.bypassed)

        # Tier 3 Mandatory Approval without callback
        gate_write_no_cb = hitl.evaluate_gate(self.sample_database_skill, ref=ref_write)
        self.assertFalse(gate_write_no_cb.approved)
        self.assertTrue(gate_write_no_cb.requires_user_input)

        # Tier 3 Mandatory Approval with callback returning True
        def user_approver(skill_name: str, params: dict) -> bool:
            return True

        hitl_approver = HITLEngine(skip_hitl=False, approval_callback=user_approver)
        gate_write_approved = hitl_approver.evaluate_gate(self.sample_database_skill, ref=ref_write)
        self.assertTrue(gate_write_approved.approved)

        # Test Bypass via skip_hitl=True
        hitl_bypass = HITLEngine(skip_hitl=True)
        gate_bypassed = hitl_bypass.evaluate_gate(self.sample_database_skill, ref=ref_write)
        self.assertTrue(gate_bypassed.approved)
        self.assertTrue(gate_bypassed.bypassed)
        self.assertEqual(gate_bypassed.tier, HITLPolicyTier.TIER_0_BYPASS_ALL)

    def test_skill_registry_integration(self):
        registry = SkillRegistry.from_roots(roots=[])
        registry._skills = {
            "python-core": self.sample_python_skill,
            "database-postgres": self.sample_database_skill,
        }
        registry.discovery_engine.register_skill(self.sample_python_skill)
        registry.discovery_engine.register_skill(self.sample_database_skill)
        registry.discovery_engine.build_index()

        # Compile all skills
        compiled_map = registry.compile_all(strict_schemas=True)
        self.assertIn("python-core", compiled_map)
        self.assertIn("database-postgres", compiled_map)

        # Intent search via registry
        search_res = registry.search_intent("postgres SQL migration")
        self.assertGreaterEqual(search_res.total_found, 1)
        self.assertEqual(search_res.matches[0].name, "database-postgres")

        # Evaluate execution gate via registry
        gate_res = registry.evaluate_execution_gate("python-core")
        self.assertTrue(gate_res.approved)


if __name__ == "__main__":
    unittest.main()
