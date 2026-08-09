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

from typing import Dict, Any

def process_multi_asset(asset_manifest: Dict[str, Any]) -> bool:
    supported_types = ["image/png", "application/pdf", "image/svg+xml", "text/csv", "application/json"]
    return all(asset.get("mime_type") in supported_types for asset in asset_manifest.get("assets", []))
