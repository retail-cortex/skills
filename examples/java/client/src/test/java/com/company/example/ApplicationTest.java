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
