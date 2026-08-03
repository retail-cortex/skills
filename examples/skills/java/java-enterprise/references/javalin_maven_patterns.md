# Java 25 (LTS) Enterprise Microservice Architecture & TDD Guide

## 1. JUnit 5 (Jupiter) & Mockito TDD

Write parameterized unit tests in Java 25 isolating database and API layers with Mockito:

```java
package com.enterprise.service;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;
import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

class ServiceTest {

    @ParameterizedTest
    @ValueSource(strings = {"8080", "9090"})
    void testValidPortConfiguration(String port) {
        assertTrue(Integer.parseInt(port) > 1024);
    }

    @Test
    void testServiceStatusWithMock() {
        DataRepository mockRepo = mock(DataRepository.class);
        when(mockRepo.isHealthy()).thenReturn(true);

        HealthService service = new HealthService(mockRepo);
        assertTrue(service.checkHealth());
    }
}
```

## 2. GitHub Actions CI/CD (`.github/workflows/maven.yml`)

```yaml
name: Java 25 CI/CD

on: [push, pull_request]

jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          distribution: 'temurin'
          java-version: '25'
      - name: Verify with JaCoCo 85% Coverage and SpotBugs
        run: mvn clean verify
```
