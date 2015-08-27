package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetric(t *testing.T) {
	m := New("TestMetric")
	assert.Equal(t, m.Name, "TestMetric")
	assert.Equal(t, m.Value, 0.0)
	assert.Equal(t, m.MetricType, "gauge")
	assert.Equal(t, len(m.Dimensions), 0)
}

func TestAddDimension(t *testing.T) {
	m := New("TestMetric")
	m.AddDimension("TestDimension", "test value")
	assert.Equal(t, len(m.Dimensions), 1)
	assert.Equal(t, m.Dimensions["TestDimension"], "test value")
}

func TestGetDimensionsWithNoDimensions(t *testing.T) {
	defaultDimensions := make(map[string]string, 0)
	m := New("TestMetric")
	assert.Equal(t, len(m.GetDimensions(defaultDimensions)), 0)
}

func TestGetDimensionsWithDimensions(t *testing.T) {
	defaultDimensions := make(map[string]string, 0)
	defaultDimensions["DefaultDim"] = "default value"
	m := New("TestMetric")
	m.AddDimension("TestDimension", "test value")
	assert.Equal(t, len(m.GetDimensions(defaultDimensions)), 2)
}

func TestGetDimensionValueFound(t *testing.T) {
	defaultDimensions := make(map[string]string, 0)
	m := New("TestMetric")
	m.AddDimension("TestDimension", "test value")
	value, ok := m.GetDimensionValue("TestDimension", defaultDimensions)

	assert.Equal(t, value, "test value")
	assert.Equal(t, ok, true)
}

func TestGetDimensionValueNotFound(t *testing.T) {
	defaultDimensions := make(map[string]string, 0)
	m := New("TestMetric")
	value, ok := m.GetDimensionValue("TestDimension", defaultDimensions)

	assert.Equal(t, value, "")
	assert.Equal(t, ok, false)
}
