// Copyright 2026 Ryan McGuinness
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package com.retailcortex.castor.client;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class CastorClientTest {

    @Test
    void testDefaultInitialization() {
        CastorClient client = new CastorClient();
        assertThat(client.getServerUrl()).isNotBlank();
        assertThat(client.getRegistry()).isNotNull();
    }

    @Test
    void testCustomInitialization() {
        CastorClient client = new CastorClient("http://custom:9090", "test-key");
        assertThat(client.getServerUrl()).isEqualTo("http://custom:9090");
        assertThat(client.getApiKey()).isEqualTo("test-key");
    }

    @Test
    void testParseUri() {
        CastorClient client = new CastorClient();
        var parsed = client.parseUri("castor://skills/example.com/testing/test-skill/1.0.0");
        assertThat(parsed.scheme()).isEqualTo("castor");
        assertThat(parsed.target()).isEqualTo("example.com/testing/test-skill/1.0.0");
    }
}
