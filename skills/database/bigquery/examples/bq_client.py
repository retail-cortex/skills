import os
from typing import Any, Dict
from google.cloud import bigquery, geminidataanalytics

# 1. Standard BigQuery SDK Execution
def run_sql_query(project_id: str, sql: str) -> list[dict]:
    client = bigquery.Client(project=project_id)
    # Always ensure fully qualified table names are used in sql
    query_job = client.query(sql)
    rows = query_job.result()
    return [dict(row) for row in rows]

# 2. Conversational Analytics (CAPI) Execution
async def run_capi_analytics(user_query: str, project_id: str, creds=None) -> Dict[str, Any]:
    client = geminidataanalytics.DataChatServiceClient(credentials=creds)

    messages = [geminidataanalytics.Message()]
    messages[0].user_message = geminidataanalytics.UserMessage(text=user_query)

    request = geminidataanalytics.ChatRequest(
        parent=f"projects/{project_id}/locations/global", # Global location invariant
        messages=messages,
    )

    output = {"sql": "", "text": "", "rows": []}
    stream = client.chat(request=request, timeout=300)

    for response in stream:
        msg = response.system_message
        if msg.text and msg.text.parts:
            output["text"] += "".join(msg.text.parts)
        if msg.data and msg.data.generated_sql:
            output["sql"] = msg.data.generated_sql

    return output
