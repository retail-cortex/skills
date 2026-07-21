# Agent-to-Agent (A2A) Protocol Specification

## Message Envelope Structure

All inter-agent communication under the A2A protocol must adhere to a standardized, strongly-typed JSON envelope containing routing headers, cryptographic verification signatures, and payload data:

```json
{
  "protocol_version": "0.2.1",
  "message_id": "msg_01HX9AZ78B4K2M",
  "correlation_id": "corr_01HX9AZ78B4K2M",
  "timestamp": "2026-07-21T08:57:36Z",
  "sender": {
    "agent_id": "agent_analyzer_01",
    "role": "data_analyzer",
    "signature": "hmac_sha256_sig_hex"
  },
  "recipient": {
    "agent_id": "agent_orchestrator_01",
    "role": "orchestrator"
  },
  "payload": {
    "action": "QUERY_RESULTS",
    "data": {
      "query_id": "bq_job_9921",
      "row_count": 1420
    }
  }
}
```

## Security & CWE Invariants

- **CWE-306 (Missing Authentication for Critical Function)**: Every message envelope must carry a valid HMAC-SHA256 signature generated using pre-shared agent credentials to prevent unauthorized agent impersonation.
- **CWE-269 (Improper Privilege Management)**: Tools requiring external services (e.g. BigQuery CAPI, Cloud Storage) must retrieve end-user delegation tokens via `ToolContext.state["user_token"]` rather than relying on shared service account keys.
- **CWE-400 (Uncontrolled Resource Consumption)**: Message dispatchers must enforce strict payload limits (maximum 1 MB) and recursion depth limits to prevent cyclic agent invocation loops.

## HTTP 429 Rate Limit & Exponential Backoff Invariants

Autonomous agents interacting over A2A routers must handle downstream API exhaustion and rate limits:
- **Token Bucket Rate Limiting**: Inbound dispatchers restrict request frequency per sender agent.
- **Exponential Backoff with Jitter**: When encountering HTTP 429 responses, agents apply backoff using `tenacity` or `retryablehttp` before re-dispatching messages.
