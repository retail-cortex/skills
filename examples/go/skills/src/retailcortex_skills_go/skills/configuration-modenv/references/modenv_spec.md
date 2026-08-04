# Hierarchical TOML Configuration, TDD & XOR Secret Specification

## 1. TDD Testing for Cascading Configuration Precedence

Write unit tests in Go or Python validating that local files override runtime files, which override base `.env.toml`:

```go
package config_test

import (
	"testing"
	"github.com/rrmcguinness/modenv/pkg/modenv"
	"github.com/stretchr/testify/assert"
)

func TestConfigurationPrecedence(t *testing.T) {
	var cfg AppConfig
	clone, err := modenv.Load(&cfg)
	assert.NoError(t, err)

	resolved := clone.(*AppConfig)
	assert.NotEmpty(t, resolved.Server.RestPort)
	// Assert that XOR secret decrypted in memory
	assert.NotContains(t, resolved.Database.Password, "xor:")
}
```

## 2. XOR Secret Encryption Algorithm

Properties starting with `xor:` are decrypted in memory upon load using the `MODENV_KEY` environment variable. Never write unencrypted passwords to Git repositories.
