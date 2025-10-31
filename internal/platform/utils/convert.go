package utils

import (
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
)

func NumericFromFloat(f float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	str := strconv.FormatFloat(f, 'g', -1, 64)
	if err := n.Scan(str); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}
