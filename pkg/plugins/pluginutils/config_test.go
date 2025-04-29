package pluginutils

import "testing"

func TestLoadPluginConfig(t *testing.T) {
	conf, err := LoadPluginConfig()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if len(conf) != 0 {
		t.FailNow()
	}
}
