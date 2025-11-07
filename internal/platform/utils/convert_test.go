package utils

import (
	"math"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestNumericFromFloat(t *testing.T) {
	tests := []struct {
		name    string
		input   float64
		wantErr bool
	}{
		{
			name:    "positive integer",
			input:   42.0,
			wantErr: false,
		},
		{
			name:    "positive decimal",
			input:   123.45,
			wantErr: false,
		},
		{
			name:    "negative decimal",
			input:   -67.89,
			wantErr: false,
		},
		{
			name:    "zero",
			input:   0.0,
			wantErr: false,
		},
		{
			name:    "very small decimal",
			input:   0.0001,
			wantErr: false,
		},
		{
			name:    "large number",
			input:   999999.99,
			wantErr: false,
		},
		{
			name:    "scientific notation compatible",
			input:   1.23e5,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NumericFromFloat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NumericFromFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the result is valid
				if !got.Valid {
					t.Error("NumericFromFloat() returned invalid pgtype.Numeric")
				}

				// For verification, we can convert it back and check it's close
				// This is a basic sanity check
				float8Val, err := got.Float64Value()
				if err != nil {
					t.Errorf("Failed to convert numeric back to float: %v", err)
				}
				
				if !float8Val.Valid {
					t.Error("Float64Value() returned invalid Float8")
				}

				// For most numbers, they should be approximately equal
				if !approximatelyEqual(float8Val.Float64, tt.input) {
					t.Errorf("NumericFromFloat() roundtrip failed: got %v, want %v", float8Val.Float64, tt.input)
				}
			}
		})
	}
}

func TestNumericFromFloat_SpecialValues(t *testing.T) {
	tests := []struct {
		name    string
		input   float64
		wantNaN bool
	}{
		{
			name:    "infinity",
			input:   math.Inf(1),
			wantNaN: false,
		},
		{
			name:    "negative infinity",
			input:   math.Inf(-1),
			wantNaN: false,
		},
		{
			name:    "NaN",
			input:   math.NaN(),
			wantNaN: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NumericFromFloat(tt.input)
			
			// For special values, NumericFromFloat may succeed
			// but we should check the result makes sense
			if err != nil {
				// Error is acceptable for special values
				return
			}
			
			if !got.Valid {
				// Invalid result is acceptable for special values
				return
			}
			
			// If we get a valid result, check if it's NaN when expected
			if tt.wantNaN && !got.NaN {
				t.Errorf("NumericFromFloat() NaN = %v, want true", got.NaN)
			}
		})
	}
}

func TestNumericFromFloat_Precision(t *testing.T) {
	// Test that precision is maintained for common use cases
	testCases := []float64{
		10.5,
		20.25,
		30.125,
		100.99,
		1234.5678,
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			numeric, err := NumericFromFloat(tc)
			if err != nil {
				t.Fatalf("NumericFromFloat(%v) error = %v", tc, err)
			}

			float8Val, err := numeric.Float64Value()
			if err != nil {
				t.Fatalf("Float64Value() error = %v", err)
			}
			
			if !float8Val.Valid {
				t.Fatal("Float64Value() returned invalid Float8")
			}

			if !approximatelyEqual(float8Val.Float64, tc) {
				t.Errorf("NumericFromFloat(%v) precision lost: got %v", tc, float8Val.Float64)
			}
		})
	}
}

// approximatelyEqual checks if two floats are approximately equal
func approximatelyEqual(a, b float64) bool {
	const epsilon = 1e-9
	diff := math.Abs(a - b)
	
	// Handle the case where both are zero
	if a == 0 && b == 0 {
		return true
	}
	
	// For very small numbers, use absolute difference
	if math.Abs(a) < epsilon || math.Abs(b) < epsilon {
		return diff < epsilon
	}
	
	// For larger numbers, use relative difference
	return diff/(math.Abs(a)+math.Abs(b)) < epsilon
}

func TestNumericFromFloat_ValidResult(t *testing.T) {
	// Test that Valid flag is set correctly
	n, err := NumericFromFloat(42.5)
	if err != nil {
		t.Fatalf("NumericFromFloat() error = %v", err)
	}

	if !n.Valid {
		t.Error("NumericFromFloat() should return Valid=true for valid input")
	}

	// Test the pgtype.Numeric can be used in database operations
	// by checking its internal state is reasonable
	var value pgtype.Numeric
	if err := value.Scan("42.5"); err != nil {
		t.Fatalf("Failed to scan test value: %v", err)
	}

	// Both should represent the same value
	float8Val1, err := n.Float64Value()
	if err != nil {
		t.Fatalf("Float64Value() error = %v", err)
	}
	float8Val2, err := value.Float64Value()
	if err != nil {
		t.Fatalf("Float64Value() error = %v", err)
	}
	
	if !float8Val1.Valid || !float8Val2.Valid {
		t.Fatal("Float64Value() returned invalid Float8")
	}

	if !approximatelyEqual(float8Val1.Float64, float8Val2.Float64) {
		t.Errorf("Values don't match: %v != %v", float8Val1.Float64, float8Val2.Float64)
	}
}
