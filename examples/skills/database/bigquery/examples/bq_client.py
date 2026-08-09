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
