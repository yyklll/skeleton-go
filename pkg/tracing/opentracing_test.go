package tracing

import (
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

func TestExtractMapFromString(t *testing.T) {
	expected := make(opentracing.TextMapCarrier)
	expected["apa"] = "12"
	expected["banan"] = "x-tracing-backend-12"
	result, err := extractMapFromString("apa=12:banan=x-tracing-backend-12")
	assert.True(t, err)
	assert.Equal(t, expected, result)
}

func TestErrorConditions(t *testing.T) {
	_, err := extractMapFromString("")
	assert.False(t, err)

	_, err = extractMapFromString("key=value:keywithnovalue")
	assert.False(t, err)
}
