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

package com.company.example;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class ApplicationTest {

    @Test
    void testGetSkmServerUrl() {
        System.setProperty("skm.server.url", "http://test-server:8080");
        assertThat(Application.getSkmServerUrl()).isEqualTo("http://test-server:8080");
    }

    @Test
    void testGetSkmApiKey() {
        System.setProperty("skm.api.key", "test-key-123");
        assertThat(Application.getSkmApiKey()).isEqualTo("test-key-123");
    }
}
