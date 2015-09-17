package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcStatusConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})

	ps := NewProcStatus(nil, 123, nil)
	ps.Configure(config)

	assert.Equal(t,
		ps.Interval(),
		123,
		"should be the default collection interval",
	)
	assert.Equal(t,
		ps.ProcessName(),
		"",
		"should be the default process name",
	)
}

func TestProcStatusConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999
	config["processName"] = "fullerite"

	ps := NewProcStatus(nil, 123, nil)
	ps.Configure(config)

	assert.Equal(t,
		ps.Interval(),
		9999,
		"should be the defined interval",
	)

	assert.Equal(t,
		ps.ProcessName(),
		"fullerite",
		"should be the defined process name",
	)
}
