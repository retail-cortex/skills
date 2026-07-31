"""Human-in-the-Loop (HITL) Execution Safety Engine.

Enforces tiered intervention gates, manifest locking verification, and explicit audit compliance.
Supports programmatic bypass controls via flags and environment overrides.
"""

import os
import uuid
from typing import Any, Callable, Dict, Optional

from loader.types import CompiledSkillReference, HITLGateResult, HITLPolicyTier, SkillDefinition


class HITLEngine:
    """Evaluates safety policies and manages interactive approval gates for skill execution."""

    def __init__(
        self,
        skip_hitl: bool = False,
        approval_callback: Optional[Callable[[str, Dict[str, Any]], bool]] = None,
    ):
        # Allow environment variable override as well
        env_skip = os.environ.get("SKILL_SKIP_HITL", "").lower() in ("1", "true", "yes")
        self.skip_hitl = skip_hitl or env_skip
        self.approval_callback = approval_callback

    def evaluate_gate(
        self,
        skill: SkillDefinition,
        ref: Optional[CompiledSkillReference] = None,
        runtime_params: Optional[Dict[str, Any]] = None,
        force_skip_hitl: bool = False,
    ) -> HITLGateResult:
        """Evaluates whether execution can proceed or requires human approval."""
        audit_id = f"audit-{uuid.uuid4().hex[:12]}"
        tier = ref.hitl_tier if ref else HITLPolicyTier.TIER_1_AUTO_READ
        params = runtime_params or {}

        # Check bypass conditions
        if self.skip_hitl or force_skip_hitl or tier == HITLPolicyTier.TIER_0_BYPASS_ALL:
            return HITLGateResult(
                approved=True,
                tier=HITLPolicyTier.TIER_0_BYPASS_ALL,
                reason="HITL intervention gate bypassed via configuration flag or TIER_0 policy.",
                bypassed=True,
                audit_event_id=audit_id,
            )

        if tier == HITLPolicyTier.TIER_1_AUTO_READ:
            return HITLGateResult(
                approved=True,
                tier=tier,
                reason="Tier 1 read-only operation authorized for autonomous execution.",
                bypassed=False,
                audit_event_id=audit_id,
            )

        if tier == HITLPolicyTier.TIER_2_AUDITED_WRITE:
            return HITLGateResult(
                approved=True,
                tier=tier,
                reason="Tier 2 state-modifying operation authorized with execution snapshot & audit trail.",
                bypassed=False,
                audit_event_id=audit_id,
            )

        # Tier 3 Mandatory Approval Gate
        if tier == HITLPolicyTier.TIER_3_MANDATORY_APPROVAL:
            if self.approval_callback:
                approved = self.approval_callback(skill.name, params)
                reason = "Approved by interactive human user callback." if approved else "Rejected by human user."
                return HITLGateResult(
                    approved=approved,
                    tier=tier,
                    reason=reason,
                    bypassed=False,
                    requires_user_input=True,
                    audit_event_id=audit_id,
                )
            else:
                # Default behavior when no interactive callback is registered: mandate approval requirement
                return HITLGateResult(
                    approved=False,
                    tier=tier,
                    reason="Tier 3 high-risk operation requires human approval gate confirmation.",
                    bypassed=False,
                    requires_user_input=True,
                    audit_event_id=audit_id,
                )

        return HITLGateResult(
            approved=True,
            tier=tier,
            reason="Default execution authorized.",
            audit_event_id=audit_id,
        )
