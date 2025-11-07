package utils

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNumericFromFloat_PositiveNumber(t *testing.T) {
	f := 123.456
	
	n, err := NumericFromFloat(f)
	require.NoError(t, err)
	assert.True(t, n.Valid)
	
	// Verify the numeric was created successfully
	assert.IsType(t, pgtype.Numeric{}, n)
}

func TestNumericFromFloat_NegativeNumber(t *testing.T) {
	f := -987.654
	
	n, err := NumericFromFloat(f)
	require.NoError(t, err)
	assert.True(t, n.Valid)
}

func TestNumericFromFloat_Zero(t *testing.T) {
	f := 0.0
	
	n, err := NumericFromFloat(f)
	require.NoError(t, err)
	assert.True(t, n.Valid)
}

func TestNumericFromFloat_SmallNumber(t *testing.T) {
	f := 0.000123
	
	n, err := NumericFromFloat(f)
	require.NoError(t, err)
	assert.True(t, n.Valid)
}

func TestNumericFromFloat_LargeNumber(t *testing.T) {
	f := 12345.987654321
	
	n, err := NumericFromFloat(f)
	require.NoError(t, err)
	assert.True(t, n.Valid)
}

func TestNumericFromFloat_Integer(t *testing.T) {
	f := 100.0
	
	n, err := NumericFromFloat(f)
	require.NoError(t, err)
	assert.True(t, n.Valid)
}

func TestNumericFromFloat_ResultIsValid(t *testing.T) {
	tests := []float64{
		0.0,
		1.0,
		-1.0,
		3.14159,
		-273.15,
		10000.0,
		0.001,
	}
	
	for _, f := range tests {
		n, err := NumericFromFloat(f)
		require.NoError(t, err, "failed for %f", f)
		assert.True(t, n.Valid, "invalid numeric for %f", f)
		
		// Verify it's a valid pgtype.Numeric
		assert.IsType(t, pgtype.Numeric{}, n)
	}
}
