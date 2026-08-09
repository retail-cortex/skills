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

import os
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.cloud_trace import CloudTraceSpanExporter

def setup_telemetry(service_name: str = "adk-agent-service") -> TracerProvider:
    project_id = os.environ.get("GOOGLE_CLOUD_PROJECT")
    if not project_id:
        print("Warning: GOOGLE_CLOUD_PROJECT not set. Disabling Cloud Trace exporter.")
        return None

    provider = TracerProvider()
    processor = BatchSpanProcessor(CloudTraceSpanExporter(project_id=project_id))
    provider.add_span_processor(processor)
    trace.set_tracer_provider(provider)
    print(f"OpenTelemetry Cloud Trace initialized for project: {project_id}")
    return provider

def trace_agent_run(session_id: str, prompt: str):
    tracer = trace.get_tracer("enterprise.agent.runner")
    with tracer.start_as_current_span("agent_execution_span") as span:
        span.set_attribute("adk.session_id", session_id)
        span.set_attribute("adk.prompt_length", len(prompt))
        span.set_attribute("gcp.project_id", os.environ.get("GOOGLE_CLOUD_PROJECT", ""))
        
        # Simulate business logic
        print(f"Executing agent run under trace: {span.get_span_context().trace_id}")
