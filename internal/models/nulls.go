package models

import "database/sql"

func NullInt64FromPtr(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}
