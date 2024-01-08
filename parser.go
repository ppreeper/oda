package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type QueryDef struct {
	Model  string
	Filter string
	Offset int
	Limit  int
	Fields string
	Count  bool
}

func parseFields(field string) (fields []string) {
	if field != "" {
		fields = strings.Split(field, ",")
	} else {
		fields = []string{}
	}
	return
}

func parseFilter(filter string) (filters []any, err error) {
	filter = strings.TrimSpace(filter)

	// pre-parse
	sqBCount, sqBDepth, parenCount, _ := countBrackets(filter)
	if len(filter) == 0 {
		return
	}
	if !(len(filter) > 4) {
		return nil, errors.New("invalid filter length")
	}
	if !(sqBCount == 0 && sqBDepth == 1) {
		return nil, errors.New("invalid filter format")
	}
	if !(parenCount == 0) {
		return nil, errors.New("invalid filter format")
	}

	// lex
	tokens, err := lexer(filter)
	if err != nil {
		fmt.Printf("%v\n", err.Error())
	}

	parenCount = 0
	var arg []any
	var sList []any
	for i := 1; i < len(tokens)-1; i++ {
		f := tokens[i]
		switch {
		case f == "(":
			parenCount += 1
			if parenCount < 2 {
				arg = []any{}
			} else {
				sList = []any{}
			}
		case f == ")":
			parenCount -= 1
			if parenCount == 1 {
				arg = append(arg, sList)
			} else {
				filters = append(filters, arg)
			}
		case f == "&":
			filters = append(filters, f)
		case f == "|":
			filters = append(filters, f)
		case f == "!":
			filters = append(filters, f)
		default:
			if parenCount > 1 {
				if IsInt(f) {
					fi, _ := strconv.Atoi(f)
					sList = append(sList, fi)
				} else if IsNumeric(f) {
					fi, _ := strconv.ParseFloat(f, 64)
					sList = append(sList, fi)
				} else if IsBool(f) {
					fb, _ := strconv.ParseBool(f)
					sList = append(sList, fb)
				} else {
					sList = append(sList, f)
				}
			} else {
				if IsInt(f) {
					fi, _ := strconv.Atoi(f)
					arg = append(arg, fi)
				} else if IsNumeric(f) {
					fi, _ := strconv.ParseFloat(f, 64)
					arg = append(arg, fi)
				} else if IsBool(f) {
					fb, _ := strconv.ParseBool(f)
					arg = append(arg, fb)
				} else {
					arg = append(arg, f)
				}
			}
		}
	}
	return
}

func countBrackets(ff string) (sqBCount int, sqBDepth int, parenCount int, parenDepth int) {
	for _, f := range ff {
		switch {
		case string(f) == "[":
			sqBCount += 1
			sqBDepth += 1
		case string(f) == "]":
			sqBCount -= 1
		case string(f) == "(":
			parenCount += 1
			parenDepth += 1
		case string(f) == ")":
			parenCount -= 1
		}
	}
	return
}

func lexer(s string) ([]string, error) {
	tokens := []string{}
	bb := []byte(s)
	for i := 0; i < len(bb); i++ {
		b := bb[i]
		switch {
		case string(b) == ",":
			ffwd, sToken := lexToken(bb[i:len(bb)-1], []string{"(", "'", ",", ")"})
			if len(sToken) > 0 {
				tokens = append(tokens, strings.TrimSpace(sToken))
				if string(bb[i+ffwd-1]) == ")" {
					i += ffwd
				}
			}
		case string(b) == "'":
			ffwd, sToken := lexToken(bb[i:len(bb)-1], []string{"'"})
			if len(sToken) > 0 {
				tokens = append(tokens, strings.TrimSpace(sToken))
				i += ffwd
			}
		case string(b) == "[":
			tokens = append(tokens, string(b))
		case string(b) == "]":
			tokens = append(tokens, string(b))
		case string(b) == "(":
			tokens = append(tokens, string(b))
		case string(b) == ")":
			tokens = append(tokens, string(b))
		default:
			continue
		}
	}
	return tokens, nil
}

func lexToken(bb []byte, endTerms []string) (ffwd int, sToken string) {
	for i := 1; i < len(bb); i++ {
		b := bb[i]
		for _, t := range endTerms {
			if string(b) == t {
				return i, string(bb[1:i])
			}
		}
	}
	return
}

func IsInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func IsBool(s string) bool {
	_, err := strconv.ParseBool(s)
	return err == nil
}
