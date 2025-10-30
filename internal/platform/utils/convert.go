package utils

import "github.com/jackc/pgx/v5/pgtype"

func NumericFromFloat(v float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(v); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}
