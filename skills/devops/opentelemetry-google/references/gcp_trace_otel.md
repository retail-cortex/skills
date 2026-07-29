# OpenTelemetry & Google Cloud Trace Instrumentation

## BatchSpanProcessor Lifecycle

In asynchronous services and short-lived tasks, spans are batched in background memory. You must call `force_flush()` or `shutdown()` on application exit:

```python
import os
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.cloud_trace import CloudTraceSpanExporter

def init_gcp_telemetry() -> TracerProvider:
    project_id = os.environ.get("GOOGLE_CLOUD_PROJECT")
    if not project_id:
        raise ValueError("GOOGLE_CLOUD_PROJECT environment variable is required.")
    
    provider = TracerProvider()
    exporter = CloudTraceSpanExporter(project_id=project_id)
    processor = BatchSpanProcessor(exporter)
    
    provider.add_span_processor(processor)
    trace.set_tracer_provider(provider)
    return provider
```

## Trace Attributes for LLM Agents

When tracing AI agent execution, attach standardized attributes:
- `gen_ai.system`: `google_genai` or `adk`
- `gen_ai.request.model`: `gemini-2.5-flash` or `gemini-2.5-pro`
- `gen_ai.agent.session_id`: Unique user session UUID
- `gen_ai.response.token_count`: Integer token usage
