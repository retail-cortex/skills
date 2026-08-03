import json
from typing import Any, Dict, List, Optional

def build_a2ui_srcdoc_payload(title: str, columns: List[Dict[str, str]], rows: List[Dict[str, Any]]) -> str:
    """Composes a self-contained sandboxed A2UI WebFrameSrcdoc HTML document."""
    
    table_headers = "".join([f'<th class="px-4 py-2 text-left text-xs font-semibold text-slate-400">{col["label"]}</th>' for col in columns])
    # Ensure action columns are right-aligned
    table_headers += '<th class="px-4 py-2 text-right text-xs font-semibold text-slate-400">Actions</th>'
    
    table_rows = ""
    for row in rows:
        row_cells = "".join([f'<td class="px-4 py-3 text-sm text-slate-200">{row.get(col["key"], "")}</td>' for col in columns])
        row_cells += '<td class="px-4 py-3 text-right"><div class="flex justify-end gap-2"><button class="px-2 py-1 bg-blue-600 hover:bg-blue-500 rounded text-xs">View</button></div></td>'
        table_rows += f'<tr class="border-b border-slate-800 hover:bg-slate-900/50">{row_cells}</tr>'

    html_template = f"""<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-slate-950 text-slate-100 p-6 font-sans antialiased">
  <div class="max-w-6xl mx-auto">
    <div class="mb-6 flex justify-between items-center">
      <h2 class="text-2xl font-bold tracking-tight text-white">{title}</h2>
      <span class="px-3 py-1 bg-emerald-500/10 text-emerald-400 text-xs font-medium rounded-full border border-emerald-500/20">Live A2UI Payload</span>
    </div>
    
    <div class="overflow-x-auto rounded-xl border border-slate-800 bg-slate-900/60 shadow-2xl">
      <table class="min-w-full divide-y divide-slate-800">
        <thead class="bg-slate-900/80">
          <tr>{table_headers}</tr>
        </thead>
        <tbody class="divide-y divide-slate-800/60">
          {table_rows}
        </tbody>
      </table>
    </div>
  </div>
</body>
</html>"""
    return html_template

def build_a2ui_v08_card_json(card_id: str, title: str, child_component_id: str) -> Dict[str, Any]:
    """Generates a Gemini Enterprise A2UI v0.8 compliant Card JSON structure."""
    return {
        "id": card_id,
        "component": {
            "Card": {
                "title": title,
                "child": child_component_id
            }
        }
    }
