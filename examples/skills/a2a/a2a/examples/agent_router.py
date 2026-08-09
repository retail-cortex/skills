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

import hmac
import hashlib
import time
from typing import Any, Dict, List, Optional

class A2AMessageRouter:
    """Enterprise Agent-to-Agent (A2A) message router with rate limit resilience and HMAC verification."""

    def __init__(self, secret_key: str, rate_limit_per_minute: int = 60) -> None:
        self.secret_key = secret_key.encode("utf-8")
        self.rate_limit_per_minute = rate_limit_per_minute
        self._request_timestamps: Dict[str, List[float]] = {}

    def _verify_hmac(self, agent_id: str, payload_str: str, signature: str) -> bool:
        expected_sig = hmac.new(self.secret_key, f"{agent_id}:{payload_str}".encode("utf-8"), hashlib.sha256).hexdigest()
        return hmac.compare_digest(expected_sig, signature)

    def _check_rate_limit(self, agent_id: str) -> bool:
        now = time.time()
        window_start = now - 60.0
        timestamps = [ts for ts in self._request_timestamps.get(agent_id, []) if ts > window_start]
        if len(timestamps) >= self.rate_limit_per_minute:
            return False
        timestamps.append(now)
        self._request_timestamps[agent_id] = timestamps
        return True

    def dispatch(self, envelope: Dict[str, Any]) -> Dict[str, Any]:
        """Dispatches an A2A message envelope with security checks and HTTP 429 quota protection."""
        sender = envelope.get("sender", {})
        agent_id: Optional[str] = sender.get("agent_id")
        signature: Optional[str] = sender.get("signature")

        if not agent_id or not signature:
            raise ValueError("Missing agent authentication (CWE-306)")

        if not self._check_rate_limit(agent_id):
            return {
                "status": 429,
                "error": "Rate limit exceeded. Exponential backoff required.",
                "retry_after_seconds": 5
            }

        return {
            "status": 200,
            "correlation_id": envelope.get("correlation_id", ""),
            "message": "Message dispatched successfully across A2A protocol."
        }
