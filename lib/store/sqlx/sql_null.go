package sqlx

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gotid/god/lib/os/gtime"
)

// CUSTOM NULL Handling structures

// NullInt64 is an alias for sql.NullInt64 data type
type NullInt64 sql.NullInt64

// Scan implements the Scanner interface for NullInt64
func (ni *NullInt64) Scan(value interface{}) error {
	var i sql.NullInt64
	if err := i.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*ni = NullInt64{Int64: i.Int64}
	} else {
		*ni = NullInt64{Int64: i.Int64, Valid: true}
	}
	return nil
}

// NullBool is an alias for sql.NullBool data type
type NullBool sql.NullBool

// Scan implements the Scanner interface for NullBool
func (nb *NullBool) Scan(value interface{}) error {
	var b sql.NullBool
	if err := b.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*nb = NullBool{Bool: b.Bool}
	} else {
		*nb = NullBool{Bool: b.Bool, Valid: true}
	}

	return nil
}

// NullFloat64 is an alias for sql.NullFloat64 data type
type NullFloat64 sql.NullFloat64

// Scan implements the Scanner interface for NullFloat64
func (nf *NullFloat64) Scan(value interface{}) error {
	var f sql.NullFloat64
	if err := f.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*nf = NullFloat64{Float64: f.Float64}
	} else {
		*nf = NullFloat64{Float64: f.Float64, Valid: true}
	}

	return nil
}

// NullString is an alias for sql.NullString data type
type NullString sql.NullString

// Scan implements the Scanner interface for NullString
func (ns *NullString) Scan(value interface{}) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*ns = NullString{String: s.String}
	} else {
		*ns = NullString{String: s.String, Valid: true}
	}

	return nil
}

// NullTime is an alias for mysql.NullTime data type
type NullTime mysql.NullTime

// Scan implements the Scanner interface for NullTime
func (nt *NullTime) Scan(value interface{}) error {
	var t mysql.NullTime
	if err := t.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*nt = NullTime{Time: t.Time}
	} else {
		*nt = NullTime{Time: t.Time, Valid: true}
	}

	return nil
}

// MarshalJSON for NullInt64
func (ni *NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ni.Int64)
}

// UnmarshalJSON for NullInt64
func (ni *NullInt64) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &ni.Int64)
	ni.Valid = err == nil
	return err
}

// MarshalJSON for NullBool
func (nb *NullBool) MarshalJSON() ([]byte, error) {
	if !nb.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nb.Bool)
}

// UnmarshalJSON for NullBool
func (nb *NullBool) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &nb.Bool)
	nb.Valid = err == nil
	return err
}

// MarshalJSON for NullFloat64
func (nf *NullFloat64) MarshalJSON() ([]byte, error) {
	if !nf.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nf.Float64)
}

// UnmarshalJSON for NullFloat64
func (nf *NullFloat64) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &nf.Float64)
	nf.Valid = err == nil
	return err
}

// MarshalJSON for NullString
func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

// UnmarshalJSON for NullString
func (ns *NullString) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &ns.String)
	ns.Valid = err == nil
	return err
}

// MarshalJSON for NullTime
func (nt *NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	val := fmt.Sprintf("\"%s\"", nt.Time.Format(time.RFC3339))
	return []byte(val), nil
}

// UnmarshalJSON for NullTime
func (nt *NullTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	x := gtime.NewFromStr(s)
	if x == nil {
		nt.Valid = false
		return nil
	}

	nt.Time = x.Time
	nt.Valid = true
	return nil
}
