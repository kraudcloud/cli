package api

import (
	"database/sql/driver"
	"errors"
	"strconv"
	"strings"
)

type VATNumber struct {
	Number string
	Valid  bool
}

func ParseVATNumber(s string) (VATNumber, error) {
	var vn VATNumber
	if err := vn.parse(s); err != nil {
		return vn, err
	}

	vn.Number = s
	vn.Valid = true
	return vn, nil
}

func (vn VATNumber) String() string {
	return string(vn.Number)
}

func (vn *VATNumber) parse(s string) error {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return errors.New("vat number is empty")
	}

	if len(s) > 32 {
		return errors.New("vat number too long")
	}

	vn.Number = s
	vn.Valid = true

	return nil
}

func (vn VATNumber) MarshalJSON() ([]byte, error) {
	if !vn.Valid {
		return nil, nil
	}

	return []byte("\"" + vn.Number + "\""), nil
}

func (vn *VATNumber) UnmarshalJSON(text []byte) error {
	s, err := strconv.Unquote(string(text))
	if err != nil {
		return err
	}
	return vn.parse(s)
}

func (vn VATNumber) Value() (driver.Value, error) {
	return vn.Number, nil
}

func (vn *VATNumber) Scan(src interface{}) error {
	s := ""
	switch src := src.(type) {
	case string:
		s = src
	case []byte:
		s = string(src)
	case nil:
		s = ""
	default:
		return errors.New("incompatible type for VATNumber")
	}

	return vn.parse(s)
}

type CountryCode struct {
	CC    string
	Valid bool
}

func ParseCountryCode(s string) (CountryCode, error) {
	var cc CountryCode
	if err := cc.parse(s); err != nil {
		return cc, err
	}

	return cc, nil
}

// https://www.iso.org/obp/ui/#search
var validCountryCodes = map[string]struct{}{
	"AL": {}, "AD": {}, "AT": {}, "BY": {},
	"BE": {}, "BA": {}, "BG": {}, "HR": {},
	"CY": {}, "CZ": {}, "DK": {}, "EE": {},
	"FO": {}, "FI": {}, "FR": {}, "DE": {},
	"GI": {}, "GR": {}, "HU": {}, "IS": {},
	"IE": {}, "IM": {}, "IT": {}, "RS": {},
	"LV": {}, "LI": {}, "LT": {}, "LU": {},
	"MK": {}, "MT": {}, "MD": {}, "MC": {},
	"ME": {}, "NL": {}, "NO": {}, "PL": {},
	"PT": {}, "RO": {}, "RU": {}, "SM": {},
	"SK": {}, "SI": {}, "ES": {}, "SE": {},
	"CH": {}, "UA": {}, "GB": {}, "VA": {},
}

func (cc *CountryCode) parse(s string) error {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return errors.New("country code is empty")
	}

	if len(s) > 2 {
		return errors.New("country code too long")
	}

	if _, ok := validCountryCodes[s]; !ok {
		return errors.New("country code is not valid")
	}

	cc.CC = s
	cc.Valid = true

	return nil
}

func (cc CountryCode) String() string {
	return string(cc.CC)
}

func (cc CountryCode) MarshalJSON() ([]byte, error) {
	if !cc.Valid {
		return nil, nil
	}

	return []byte("\"" + cc.CC + "\""), nil
}

func (cc *CountryCode) UnmarshalJSON(text []byte) error {
	s, err := strconv.Unquote(string(text))
	if err != nil {
		return err
	}
	return cc.parse(s)
}

func (cc CountryCode) Value() (driver.Value, error) {
	return cc.String(), nil
}

func (cc *CountryCode) Scan(src interface{}) error {
	s := ""
	switch src := src.(type) {
	case string:
		s = src
	case []byte:
		s = string(src)
	case nil:
		s = ""
	default:
		return errors.New("incompatible type for CountryCode")
	}

	if s == "" {
		return nil
	}

	if err := cc.parse(s); err != nil {
		return err
	}

	return nil
}
