package bencode

import (
	"bufio"
	"errors"
	"strconv"
)

var (
	// ErrDictInvalid ...
	ErrDictInvalid error = errors.New("invalid dict")
	// ErrListInvalid ...
	ErrListInvalid error = errors.New("invalid list")
	// ErrIntInvalid ...
	ErrIntInvalid error = errors.New("invalid int")
	// ErrStringInvalid ...
	ErrStringInvalid error = errors.New("invalid string")
)

const stringSeparator = ':'

// ReadString reads a byte sequence which usually is a string.
//
// String in bencoding is represented as:
// <length of string>:<string>
//
// Example:
// 4:wiki
// is a string "wiki".
func ReadString(r *bufio.Reader) (string, error) {
	l, err := r.ReadBytes(stringSeparator)
	if err != nil {
		return "", err
	}
	length, err := strconv.Atoi(string(l[:len(l)-1]))
	if err != nil {
		return "", err
	}
	if length < 0 {
		return "", ErrStringInvalid
	}

	bs := []byte{}
	for i := 0; i < length; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		bs = append(bs, b)
	}

	return string(bs), nil
}

// ReadInt reads a byte sequence and returns an integer.
//
// Integers in bencoding are represented as:
// i<integer>e
//
// Example:
// i90e
// is an int 90.
func ReadInt(r *bufio.Reader) (int, error) {
	if b, _ := r.ReadByte(); b != 'i' {
		return 0, ErrIntInvalid
	}

	b, err := r.ReadBytes('e')
	if err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(string(b[:len(b)-1]))
	if err != nil {
		return 0, ErrIntInvalid
	}

	return i, nil
}

// ReadList reads a byte sequence and tries to interpret it
// as a []interface{}.
//
// Lists in bencoding are represented as:
// l[value 1][value2][...]e
//
// Example:
// l4:spam4:eggs12:cheeseburgere
// is a []string{spam, eggs, cheeseburger}
//
// However elements of the list are not necessarily are strings
// they can be any bencoding type, distionaries included.
func ReadList(r *bufio.Reader) ([]interface{}, error) {
	if b, _ := r.ReadByte(); b != 'l' {
		return nil, ErrListInvalid
	}

	l := []interface{}{}
	for {
		next, err := r.Peek(1)
		if err != nil {
			return nil, err
		}

		switch next[0] {
		case 'e':
			_, _ = r.ReadByte()
			return l, nil
		case 'l':
			list, err := ReadList(r)
			if err != nil {
				return nil, err
			}

			l = append(l, list)
		case 'd':
			dict, err := ReadDictionary(r)
			if err != nil {
				return nil, err
			}

			l = append(l, dict)
		case 'i':
			i, err := ReadInt(r)
			if err != nil {
				return nil, err
			}

			l = append(l, i)
		default:
			s, err := ReadString(r)
			if err != nil {
				return nil, err
			}

			l = append(l, s)
		}
	}
}

// ReadDictionary reads a byte sequence and tries to interpret it
// as a map[string]interface{}
//
// Dictionaries in bencoding are represented as:
// d[key1][value1][key2][value2][...]e
// Keys must be strings and must be ordered alphabetically.
// Values seem to by of any type.
//
// Example:
// d5:apple3:red6:banana6:yellow5:lemon6:yellow6:violet4:bluee
// is a []map[string]string{"apple":"red","banana":"yellow","violet":"blue"}
//
// Is the name ParseDictionary more suitable?
func ReadDictionary(r *bufio.Reader) (map[string]interface{}, error) {
	if b, _ := r.ReadByte(); b != 'd' {
		return nil, ErrDictInvalid
	}

	d := make(map[string]interface{})

	for {
		next, err := r.Peek(1)
		if next[0] == 'e' {
			_, _ = r.ReadByte()
			break
		}

		k, err := ReadString(r)
		if err != nil {
			return nil, err
		}

		next, err = r.Peek(1)
		if err != nil {
			return nil, err
		}

		var v interface{}
		switch next[0] {
		/*
			case 'e':
				_, _ = r.ReadByte()
				break*/
		case 'd':
			v, err = ReadDictionary(r)
		case 'i':
			v, err = ReadInt(r)
		case 'l':
			v, err = ReadList(r)
		default:
			v, err = ReadString(r)
		}

		if err != nil {
			return nil, err
		}

		d[k] = v
	}

	return d, nil
}
