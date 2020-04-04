package jsontotype

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io"
	"math"
	"strings"
)

// https://github.com/golang/go/wiki/CodeReviewComments#initialisms
var initialisms = map[string]bool{
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"CSV":   true,
	"DB":    true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"KYC":   true,
	"LHS":   true,
	"NTP":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SSH":   true,
	"TLS":   true,
	"TTL":   true,
	"UI":    true,
	"UID":   true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"UUID":  true,
	"VM":    true,
	"XML":   true,
}

type jsonType int

const (
	_ jsonType = iota
	jsonTypeStr
	jsonTypeNum
	jsonTypeNull
	jsonTypeBool
	jsonTypeObj
	jsonTypeArr
)

type jsonTok struct {
	jsonType jsonType
	key      string
	val      interface{}
}

/*
tok = str | num | nul | bool | obj | arr
obj = "{" str ":" tok ("," str ":" expr)* "}"
arr = "[" tok ("," tok)* "]"
*/

func Exec(r io.Reader, pkgName string, typeName string) (string, error) {
	dec := json.NewDecoder(r)
	tok, err := tokenize(dec)
	if err != nil {
		return "", err
	}
	typeBody, err := parse(tok)
	if err != nil {
		return "", err
	}
	raw := fmt.Sprintf("package %s;type %s %s", pkgName, typeName, typeBody)
	b, err := format.Source([]byte(raw))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func tokenize(dec *json.Decoder) (*jsonTok, error) {
	return tok(dec)
}

func tok(dec *json.Decoder) (*jsonTok, error) {
	t, err := dec.Token()
	if err == io.EOF {
		return nil, errors.New("no value")
	}
	if err != nil {
		return nil, err
	}

	switch v := t.(type) {
	case string:
		return &jsonTok{jsonType: jsonTypeStr, val: v}, nil
	case float64:
		return &jsonTok{jsonType: jsonTypeNum, val: v}, nil
	case nil:
		return &jsonTok{jsonType: jsonTypeNull, val: v}, nil
	case bool:
		return &jsonTok{jsonType: jsonTypeBool, val: v}, nil
	case json.Delim:
		switch v {
		case '{':
			node, err := obj(dec)
			if err != nil {
				return nil, err
			}
			t, err := dec.Token()
			if err == io.EOF {
				return nil, errors.New("no value")
			}
			if err != nil {
				return nil, err
			}
			delim, ok := t.(json.Delim)
			if !ok {
				return nil, errors.New("Expect }")
			}
			if delim != '}' {
				return nil, errors.New("Expect }")
			}
			return node, nil

		case '[':
			node, err := arr(dec)
			if err != nil {
				return nil, err
			}
			t, err := dec.Token()
			if err == io.EOF {
				return nil, errors.New("no value")
			}
			if err != nil {
				return nil, err
			}
			delim, ok := t.(json.Delim)
			if !ok {
				return nil, errors.New("Expect }")
			}
			if delim != ']' {
				return nil, errors.New("Expect }")
			}
			return node, nil
		}
	}
	return nil, nil
}

// obj = "{" str ":" tok ("," str ":" expr)* "}"
func obj(dec *json.Decoder) (*jsonTok, error) {
	var o []*jsonTok
	for {
		if !dec.More() {
			return &jsonTok{jsonType: jsonTypeObj, val: o}, nil
		}
		t, err := dec.Token()
		if err == io.EOF {
			return nil, errors.New("no value")
		}
		if err != nil {
			return nil, err
		}
		key, ok := t.(string)
		if !ok {
			return nil, errors.New("object key must be string")
		}
		current, err := tok(dec)
		if err != nil {
			return nil, err
		}
		current.key = key
		o = append(o, current)
	}
}

func arr(dec *json.Decoder) (*jsonTok, error) {
	var a []*jsonTok
	for {
		if !dec.More() {
			return &jsonTok{jsonType: jsonTypeArr, val: a}, nil
		}
		current, err := tok(dec)
		if err != nil {
			return nil, err
		}
		a = append(a, current)
	}
}

func parse(tok *jsonTok) (string, error) {
	// var body
	switch tok.jsonType {
	case jsonTypeStr:
		return "string", nil

	case jsonTypeNum:
		v, ok := tok.val.(float64)
		if !ok {
			return "", errors.New("parse error number")
		}
		return getNumberType(v), nil

	case jsonTypeNull:
		return "", errors.New("can't use null")

	case jsonTypeBool:
		return "bool", nil

	case jsonTypeObj:
		vs, ok := tok.val.([]*jsonTok)
		if !ok {
			return "", errors.New("parse error object")
		}
		if len(vs) == 0 {
			return "", errors.New("unexpect empty object")
		}
		var os string
		for _, v := range vs {
			o, err := parse(v)
			if err != nil {
				return "", err
			}
			os += fmt.Sprintf("%s %s `json:\"%s\"`;", toCamelCase(v.key), o, v.key)
		}
		return fmt.Sprintf("struct{%s}", os), nil

	case jsonTypeArr:
		vs, ok := tok.val.([]*jsonTok)
		if !ok {
			return "", errors.New("parse error array")
		}
		if len(vs) == 0 {
			return "", errors.New("unexpect empty array")
		}
		o, err := parse(vs[0])
		if err != nil {
			return "", err
		}
		return "[]" + o, nil
	}
	return "", nil
}

func getNumberType(f float64) string {
	if math.Floor(f) == f {
		return "int64"
	}
	return "float64"
}

func toCamelCase(raw string) string {
	var (
		txt       string
		nowWord   string
		isToUpper bool
	)

	for i, r := range raw {
		// first rune must be upper case
		if i == 0 {
			nowWord = strings.ToUpper(string(r))
			continue
		}

		// for snake case
		if isToUpper {
			nowWord += strings.ToUpper(string(r))
			isToUpper = false
			continue
		}

		// for snake case
		if r == '_' {
			isToUpper = true
			if upNowWord := strings.ToUpper(nowWord); initialisms[upNowWord] {
				nowWord = upNowWord
			}
			txt += nowWord
			nowWord = ""
			continue
		}

		// check initialisms
		if string(r) == strings.ToUpper(string(r)) {
			if upNowWord := strings.ToUpper(nowWord); initialisms[upNowWord] {
				nowWord = upNowWord
			}
			txt += nowWord
			nowWord = string(r)
			continue
		}

		// normal rune
		nowWord += string(r)
	}

	// check initialisms
	if upNowWord := strings.ToUpper(nowWord); initialisms[upNowWord] {
		nowWord = upNowWord
	}

	return txt + nowWord
}
