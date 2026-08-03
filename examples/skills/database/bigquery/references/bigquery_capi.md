# BigQuery Conversational Analytics (CAPI) & Protobuf Normalization

## Protobuf to Plain Python Object Normalization

Protobuf message objects returned by `google.cloud.geminidataanalytics` crash standard JSON and ADK serializers. Convert them recursively:

```python
def to_plain(obj):
    if obj is None or isinstance(obj, (str, int, float, bool)):
        return obj
    if isinstance(obj, dict):
        return {k: to_plain(v) for k, v in obj.items()}
    if isinstance(obj, (list, tuple)):
        return [to_plain(v) for v in obj]
    try:
        return {k: to_plain(v) for k, v in dict(obj).items()}
    except (TypeError, ValueError):
        pass
    try:
        return [to_plain(v) for v in list(obj)]
    except (TypeError, ValueError):
        pass
    return str(obj)
```

## CAPI Client Streaming Invariants

```python
from google.cloud import geminidataanalytics

# CRITICAL: Location MUST be global
request = geminidataanalytics.ChatRequest(
    parent=f"projects/{project_id}/locations/global",
    messages=[
        geminidataanalytics.Message(
            user_message=geminidataanalytics.UserMessage(text=user_query)
        )
    ],
)

client = geminidataanalytics.DataChatServiceClient(credentials=user_creds)
stream = client.chat(request=request, timeout=300)
```
