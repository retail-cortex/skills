"""OpenTelemetry reporting module for Google Cloud Platform."""

import logging
import os
from typing import Optional

from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource

logger = logging.getLogger("skills_service.telemetry")

_tracer_provider: Optional[TracerProvider] = None


def setup_telemetry(
    enable_telemetry: Optional[bool] = None,
    service_name: Optional[str] = None,
    gcp_project_id: Optional[str] = None,
) -> Optional[TracerProvider]:
    """Initializes OpenTelemetry tracing and exports to Google Cloud Trace if configured."""
    global _tracer_provider

    enabled = enable_telemetry if enable_telemetry is not None else (os.getenv("ENABLE_OPENTELEMETRY", "false").lower() == "true")
    if not enabled:
        logger.info("OpenTelemetry disabled via configuration.")
        return None

    try:
        from opentelemetry.exporter.cloud_trace import CloudTraceSpanExporter

        sname = service_name or os.getenv("OTEL_SERVICE_NAME", "skills-service")
        resource = Resource.create({"service.name": sname})
        provider = TracerProvider(resource=resource)

        proj_id = gcp_project_id or os.getenv("GCP_PROJECT_ID") or None
        exporter = CloudTraceSpanExporter(project_id=proj_id)
        processor = BatchSpanProcessor(exporter)
        provider.add_span_processor(processor)

        trace.set_tracer_provider(provider)
        _tracer_provider = provider
        logger.info("OpenTelemetry Google Cloud Trace exporter initialized successfully.")
        return provider
    except Exception as exc:
        logger.warning("OpenTelemetry GCP setup failed gracefully: %s", exc)
        return None


def get_tracer(name: str = "skills_service") -> trace.Tracer:
    """Returns an OpenTelemetry tracer instance."""
    return trace.get_tracer(name)
