package bencode

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadInt(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		expectedInt int
		expectedErr error
	}{
		// Positive cases
		{
			name:        "valid: i0e is a valid 0",
			in:          "i0e",
			expectedInt: 0,
		},
		{
			name:        "valid: i1e is a valid 1",
			in:          "i1e",
			expectedInt: 1,
		},
		{
			name:        "valid: i1ee is a valid 1",
			in:          "i1ee",
			expectedInt: 1,
		},
		{
			name:        "i-1e is a valid -1",
			in:          "i-1e",
			expectedInt: -1,
		},
		{
			name:        "i000000000000000000000e is a valid 0",
			in:          "i000000000000000000000e",
			expectedInt: 0,
		},

		// Negative cases
		{
			name: "invalid: i0 is not a valid int",
			in:   "i0",
			// io.EOF
			expectedErr: ErrIntInvalid,
		},
		{
			name: "invalid: a is not a valid int",
			in:   "a",
			// strconv.ErrSyntax
			expectedErr: ErrIntInvalid,
		},
		{
			name: "invalid: ie is not a valid int",
			in:   "ie",
			// io.EOF
			expectedErr: ErrIntInvalid,
		},
		{
			name: "invalid: iae not a valid int",
			in:   "iae",
			// strconv.ErrSyntax
			expectedErr: ErrIntInvalid,
		},
		{
			name: "invalid: 0e is invalid",
			in:   "0e",
			// io.EOF
			expectedErr: ErrIntInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(test.in))
			i, err := ReadInt(r)

			if test.expectedErr != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedInt, i)
			}
		})
	}
}

func TestReadString(t *testing.T) {
	tests := []struct {
		name           string
		in             string
		expectedString string
		expectedErr    error
	}{
		// Positive cases
		{
			name:           "0: is a valid string",
			in:             "0:",
			expectedString: "",
		},
		{
			name:           "1:a is a valid string",
			in:             "1:a",
			expectedString: "a",
		},
		{
			name:           "1:ab is a valid a",
			in:             "1:ab",
			expectedString: "a",
		},

		// Negative cases
		{
			name:        "aaaa is not a valid string",
			in:          "aaaa",
			expectedErr: ErrStringInvalid,
		},
		{
			name:        ":aaaa is not a valid string",
			in:          ":aaaa",
			expectedErr: ErrStringInvalid,
		},
		{
			name:        "-5:aaaa is not a valid string",
			in:          "-5:aaaaa",
			expectedErr: ErrStringInvalid,
		},
		{
			name: "invalid: 3:a is not a valid string",
			in:   "5:a",
			// io.EOF
			expectedErr: ErrStringInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(test.in))
			s, err := ReadString(r)
			if test.expectedErr != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedString, s)
			}
		})
	}
}

func TestReadList(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		expectedList []interface{}
		expectedErr  error
	}{
		// Positive cases
		{
			name:         "valid: le is a valid empty list",
			in:           "le",
			expectedList: []interface{}{},
		},
		// List of ints
		{
			name:         "valid: li0ee is a valid list of ints",
			in:           "li0ee",
			expectedList: []interface{}{0},
		},
		{
			name:         "valid: li0ee is a valid list of ints",
			in:           "li0ei1ee",
			expectedList: []interface{}{0, 1},
		},
		// List of strings
		{
			name:         "valid: l1:a2:abe is a valid list of strings",
			in:           "l1:a2:eee",
			expectedList: []interface{}{"a", "ee"},
		},
		// List of lists
		{
			name:         "valid: lli0eee is a valid list of lists of ints",
			in:           "lli0eee",
			expectedList: []interface{}{[]interface{}{0}},
		},
		// List of dicts
		{
			name:         "valid: list of dicts, empty",
			in:           "ldee",
			expectedList: []interface{}{map[string]interface{}{}},
		},

		// Negative cases
		{
			name:        "invalid: i0ee is not a valid list",
			in:          "i0ee",
			expectedErr: ErrListInvalid,
		},
		// Unexpected EOF
		{
			name:        "invalid: l is not a valid list",
			in:          "l",
			expectedErr: io.EOF,
		},
		{
			name: "invalid: li0 is not a valid list",
			in:   "li0",
			// io.EOF
			expectedErr: ErrIntInvalid,
		},
		{
			name:        "invalid: nested list is not closed",
			in:          "lli1e",
			expectedErr: io.EOF,
		},
		{
			name:        "invalid: the outer list is not closed",
			in:          "lli0ee",
			expectedErr: io.EOF,
		},
		// List of strings
		{
			name: "invalid: l3:a is not a valid list",
			in:   "l3:a",
			// io.EOF
			expectedErr: ErrStringInvalid,
		},
		// List of dicts
		{
			name:        "invalid: nested dict is not closed",
			in:          "ld",
			expectedErr: io.EOF,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(test.in))
			l, err := ReadList(r)

			if test.expectedErr != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedList, l)
			}
		})
	}
}

func TestReadDictionary(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		expectedMap map[string]interface{}
		expectedErr error
	}{
		// Positive cases
		{
			name:        "valid: empty dictionary",
			in:          "de",
			expectedMap: map[string]interface{}{},
		},
		{
			name: "valid: the value is nil",
			in:   "d1:ae",
			expectedMap: map[string]interface{}{
				"a": nil,
			},
		},
		// String value
		{
			name: "valid: map[string]string with one element",
			in:   "d1:a1:be",
			expectedMap: map[string]interface{}{
				"a": "b",
			},
		},
		// Int value
		{
			name: "valid: map[string]int with one element",
			in:   "d1:ai1ee",
			expectedMap: map[string]interface{}{
				"a": 1,
			},
		},
		// List value
		{
			name: "valid: map[string][]string with an empty slice",
			in:   "d1:alee",
			expectedMap: map[string]interface{}{
				"a": []interface{}{},
			},
		},
		{
			name: "valid: map[string][]string with one element",
			in:   "d1:ali1eee",
			expectedMap: map[string]interface{}{
				"a": []interface{}{1},
			},
		},
		// Dict value
		{
			name: "valid: dict of dict",
			in:   "d1:ad1:a1:bee",
			expectedMap: map[string]interface{}{
				"a": map[string]interface{}{
					"a": "b",
				},
			},
		},

		// Negative cases
		{
			name: "invalid: dict doesn't start with 'd'",
			in:   "e",
			// io.EOF
			expectedErr: ErrDictInvalid,
		},
		{
			name: "invalid: int can't be a key",
			in:   "di1e1:ae",
			// strconv.ErrSyntax
			expectedErr: ErrStringInvalid,
		},
		{
			name: "invalid: invalid string as a value",
			in:   "d1:a2:e",
			// io.EOF
			expectedErr: ErrStringInvalid,
		},
		{
			name: "invalid: invalid int",
			in:   "d1:aiee",
			// io.EOF
			expectedErr: ErrIntInvalid,
		},
		{
			name:        "invalid: ends after the key",
			in:          "d1:a",
			expectedErr: io.EOF,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(test.in))
			d, err := ReadDictionary(r)

			if err != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedMap, d)
			}
		})
	}
}
