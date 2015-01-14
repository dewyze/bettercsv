// Copyright 2014 John DeWyze. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bettercsv

import (
	"reflect"
	"strings"
	"testing"
)

type BetterCsvTesting struct {
	t *testing.T
}

func NewBetterCsvTesting(t *testing.T) *BetterCsvTesting {
	return &BetterCsvTesting{
		t: t,
	}
}

type Test struct {
	Name               string
	Input              string
	Output             [][]string
	OutputMap          []map[string]string
	Headers            []string
	UseFieldsPerRecord bool // false (default) means FieldsPerRecord is -1
	UseHeaders         bool // true means use Headers methods for reading
	UseHeadersAndErrs  bool // true means use HeadersAndErrors methods for reading

	// These fields are copied into the Reader
	Comma            rune
	Comment          rune
	FieldsPerRecord  int
	LazyQuotes       bool
	TrailingComma    bool
	TrimLeadingSpace bool
	SkipLineOnErr    bool

	Error  string
	Line   int // Expected error line if != 0
	Column int // Expected error column if line != 0
	Errors []string
}

var readTests = []Test{
	{
		Name:   "Simple",
		Input:  "a,b,c\n",
		Output: [][]string{{"a", "b", "c"}},
	},
	{
		Name:   "CRLF",
		Input:  "a,b\r\nc,d\r\n",
		Output: [][]string{{"a", "b"}, {"c", "d"}},
	},
	{
		Name:   "BareCR",
		Input:  "a,b\rc,d\r\n",
		Output: [][]string{{"a", "b\rc", "d"}},
	},
	{
		Name:               "RFC4180test",
		UseFieldsPerRecord: true,
		Input: `#field1,field2,field3
"aaa","bb
b","ccc"
"a,a","b""bb","ccc"
zzz,yyy,xxx
`,
		Output: [][]string{
			{"#field1", "field2", "field3"},
			{"aaa", "bb\nb", "ccc"},
			{"a,a", `b"bb`, "ccc"},
			{"zzz", "yyy", "xxx"},
		},
	},
	{
		Name:   "NoEOLTest",
		Input:  "a,b,c",
		Output: [][]string{{"a", "b", "c"}},
	},
	{
		Name:   "Semicolon",
		Comma:  ';',
		Input:  "a;b;c\n",
		Output: [][]string{{"a", "b", "c"}},
	},
	{
		Name: "MultiLine",
		Input: `"two
line","one line","three
line
field"`,
		Output: [][]string{{"two\nline", "one line", "three\nline\nfield"}},
	},
	{
		Name:  "BlankLine",
		Input: "a,b,c\n\nd,e,f\n\n",
		Output: [][]string{
			{"a", "b", "c"},
			{"d", "e", "f"},
		},
	},
	{
		Name:             "TrimSpace",
		Input:            " a,  b,   c\n",
		TrimLeadingSpace: true,
		Output:           [][]string{{"a", "b", "c"}},
	},
	{
		Name:   "LeadingSpace",
		Input:  " a,  b,   c\n",
		Output: [][]string{{" a", "  b", "   c"}},
	},
	{
		Name:    "Comment",
		Comment: '#',
		Input:   "#1,2,3\na,b,c\n#comment",
		Output:  [][]string{{"a", "b", "c"}},
	},
	{
		Name:   "NoComment",
		Input:  "#1,2,3\na,b,c",
		Output: [][]string{{"#1", "2", "3"}, {"a", "b", "c"}},
	},
	{
		Name:       "LazyQuotes",
		LazyQuotes: true,
		Input:      `a "word","1"2",a","b`,
		Output:     [][]string{{`a "word"`, `1"2`, `a"`, `b`}},
	},
	{
		Name:       "BareQuotes",
		LazyQuotes: true,
		Input:      `a "word","1"2",a"`,
		Output:     [][]string{{`a "word"`, `1"2`, `a"`}},
	},
	{
		Name:       "BareDoubleQuotes",
		LazyQuotes: true,
		Input:      `a""b,c`,
		Output:     [][]string{{`a""b`, `c`}},
	},
	{
		Name:  "BadDoubleQuotes",
		Input: `a""b,c`,
		Error: `bare " in non-quoted-field`, Line: 1, Column: 1,
	},
	{
		Name:             "TrimQuote",
		Input:            ` "a"," b",c`,
		TrimLeadingSpace: true,
		Output:           [][]string{{"a", " b", "c"}},
	},
	{
		Name:  "BadBareQuote",
		Input: `a "word","b"`,
		Error: `bare " in non-quoted-field`, Line: 1, Column: 2,
	},
	{
		Name:  "BadTrailingQuote",
		Input: `"a word",b"`,
		Error: `bare " in non-quoted-field`, Line: 1, Column: 10,
	},
	{
		Name:  "ExtraneousQuote",
		Input: `"a "word","b"`,
		Error: `extraneous " in field`, Line: 1, Column: 3,
	},
	{
		Name:               "BadFieldCount",
		UseFieldsPerRecord: true,
		Input:              "a,b,c\nd,e",
		Error:              "wrong number of fields", Line: 2,
	},
	{
		Name:               "BadFieldCount1",
		UseFieldsPerRecord: true,
		FieldsPerRecord:    2,
		Input:              `a,b,c`,
		Error:              "wrong number of fields", Line: 1,
	},
	{
		Name:   "FieldCount",
		Input:  "a,b,c\nd,e",
		Output: [][]string{{"a", "b", "c"}, {"d", "e"}},
	},
	{
		Name:   "TrailingCommaEOF",
		Input:  "a,b,c,",
		Output: [][]string{{"a", "b", "c", ""}},
	},
	{
		Name:   "TrailingCommaEOL",
		Input:  "a,b,c,\n",
		Output: [][]string{{"a", "b", "c", ""}},
	},
	{
		Name:             "TrailingCommaSpaceEOF",
		TrimLeadingSpace: true,
		Input:            "a,b,c, ",
		Output:           [][]string{{"a", "b", "c", ""}},
	},
	{
		Name:             "TrailingCommaSpaceEOL",
		TrimLeadingSpace: true,
		Input:            "a,b,c, \n",
		Output:           [][]string{{"a", "b", "c", ""}},
	},
	{
		Name:             "TrailingCommaLine3",
		TrimLeadingSpace: true,
		Input:            "a,b,c\nd,e,f\ng,hi,",
		Output:           [][]string{{"a", "b", "c"}, {"d", "e", "f"}, {"g", "hi", ""}},
	},
	{
		Name:   "NotTrailingComma3",
		Input:  "a,b,c, \n",
		Output: [][]string{{"a", "b", "c", " "}},
	},
	{
		Name:          "CommaFieldTest",
		TrailingComma: true,
		Input: `x,y,z,w
x,y,z,
x,y,,
x,,,
,,,
"x","y","z","w"
"x","y","z",""
"x","y","",""
"x","","",""
"","","",""
`,
		Output: [][]string{
			{"x", "y", "z", "w"},
			{"x", "y", "z", ""},
			{"x", "y", "", ""},
			{"x", "", "", ""},
			{"", "", "", ""},
			{"x", "y", "z", "w"},
			{"x", "y", "z", ""},
			{"x", "y", "", ""},
			{"x", "", "", ""},
			{"", "", "", ""},
		},
	},
	{
		Name:             "TrailingCommaIneffective1",
		TrailingComma:    true,
		TrimLeadingSpace: true,
		Input:            "a,b,\nc,d,e",
		Output: [][]string{
			{"a", "b", ""},
			{"c", "d", "e"},
		},
	},
	{
		Name:             "TrailingCommaIneffective2",
		TrailingComma:    false,
		TrimLeadingSpace: true,
		Input:            "a,b,\nc,d,e",
		Output: [][]string{
			{"a", "b", ""},
			{"c", "d", "e"},
		},
	},
	{
		Name:          "SkipLine1DoubleQuote",
		SkipLineOnErr: true,
		Input:         "a\nb\"\nc",
		Output:        [][]string{{"a"}, {"c"}},
		Errors:        []string{"line 2, column 2: bare \" in non-quoted-field"},
	},
	{
		Name:          "SkipLine2DoubleQuote",
		SkipLineOnErr: true,
		Input:         "a\nb\"b\"\nc",
		Output:        [][]string{{"a"}, {"c"}},
		Errors:        []string{"line 2, column 4: bare \" in non-quoted-field"},
	},
	{
		Name:               "SkipLineNoOfArgs",
		SkipLineOnErr:      true,
		UseFieldsPerRecord: true,
		Input:              "a,b,c\nd,e,f,g\nh,i,j",
		Output:             [][]string{{"a", "b", "c"}, {"h", "i", "j"}},
		Errors:             []string{"line 2, column 0: wrong number of fields in line"},
	},
	{
		Name:          "SkipLineExtraneousQuote",
		SkipLineOnErr: true,
		Input:         "a,b,c\nd,\"e\"e\",f\ng,h,i",
		Output:        [][]string{{"a", "b", "c"}, {"g", "h", "i"}},
		Errors:        []string{"line 2, column 8: extraneous \" in field"},
	},
	{
		Name:               "SkipLineMultilineFieldWithErrors",
		SkipLineOnErr:      true,
		UseFieldsPerRecord: true,
		Input:              "a,b,c\nd,\"e\"\nf\",g\nh,i,j",
		Output:             [][]string{{"a", "b", "c"}, {"h", "i", "j"}},
		Errors:             []string{"line 2, column 0: wrong number of fields in line", "line 3, column 4: bare \" in non-quoted-field"},
	},
	{
		Name:               "GetHeaders",
		UseFieldsPerRecord: true,
		Input:              "a,b,c\n1,2,3",
		Headers:            []string{"a", "b", "c"},
	},
	{
		Name:               "ReadAllToMaps",
		UseFieldsPerRecord: true,
		UseHeaders:         true,
		Input:              "a,b,c\n1,2,3\n4,5,6",
		OutputMap: []map[string]string{
			{"a": "a", "b": "b", "c": "c"},
			{"a": "1", "b": "2", "c": "3"},
			{"a": "4", "b": "5", "c": "6"}},
	},
	{
		Name:               "ReadAllToMapsWithErrors",
		UseFieldsPerRecord: true,
		UseHeadersAndErrs:  true,
		Input:              "a,b,c\n1,2\",3\n4,5,6\n7,8,9,10\n11,12,13",
		Errors:             []string{"line 2, column 6: bare \" in non-quoted-field", "line 4, column 0: wrong number of fields in line"},
		OutputMap: []map[string]string{
			{"a": "a", "b": "b", "c": "c"},
			{"a": "4", "b": "5", "c": "6"},
			{"a": "11", "b": "12", "c": "13"}},
	},
}

func (t *BetterCsvTesting) DeepCompareAllAndPrint(out [][]string, test Test) {
	if !reflect.DeepEqual(out, test.Output) {
		t.t.Errorf("%s: out=%q want %q", test.Name, out, test.Output)
	}
}

func (t *BetterCsvTesting) DeepCompareErrorAndPrint(errors []error, test Test) {
	var errorStrings []string
	for _, err := range errors {
		errorStrings = append(errorStrings, err.Error())
	}
	if !reflect.DeepEqual(errorStrings, test.Errors) {
		t.t.Errorf("%s: errors=%q want %q", test.Name, errorStrings, test.Errors)
	}
}

func (t *BetterCsvTesting) DeepCompareMapAndPrint(out []map[string]string, test Test) {
	if !reflect.DeepEqual(out, test.OutputMap) {
		t.t.Errorf("%s: out=%q want %q", test.Name, out, test.OutputMap)
	}
}

func TestRead(t *testing.T) {
	betterCsvTests := NewBetterCsvTesting(t)
	for _, tt := range readTests {
		r := NewReader(strings.NewReader(tt.Input))
		r.Comment = tt.Comment
		if tt.UseFieldsPerRecord {
			r.FieldsPerRecord = tt.FieldsPerRecord
		} else {
			r.FieldsPerRecord = -1
		}
		r.LazyQuotes = tt.LazyQuotes
		r.TrailingComma = tt.TrailingComma
		r.TrimLeadingSpace = tt.TrimLeadingSpace
		r.SkipLineOnErr = tt.SkipLineOnErr
		if tt.Comma != 0 {
			r.Comma = tt.Comma
		}
		if tt.Name == "GetHeaders" {
			r.ReadAllToMaps()
			if !reflect.DeepEqual(r.Headers, tt.Headers) {
				t.Errorf("%s: headers=%q, want=%q", tt.Name, r.Headers, tt.Headers)
			}
		} else if tt.SkipLineOnErr {
			out, errors := r.ReadAllWithErrors()
			betterCsvTests.DeepCompareErrorAndPrint(errors, tt)
			betterCsvTests.DeepCompareAllAndPrint(out, tt)
		} else if tt.UseHeaders {
			out, err := r.ReadAllToMaps()
			if err != nil {
				t.Errorf("%s: unexpected error %v", tt.Name, err)
			} else {
				betterCsvTests.DeepCompareMapAndPrint(out, tt)
			}
		} else if tt.UseHeadersAndErrs {
			out, errs := r.ReadAllToMapsWithErrors()
			betterCsvTests.DeepCompareMapAndPrint(out, tt)
			betterCsvTests.DeepCompareErrorAndPrint(errs, tt)
		} else {
			out, err := r.ReadAll()
			perr, _ := err.(*ParseError)
			if tt.Error != "" {
				if err == nil || !strings.Contains(err.Error(), tt.Error) {
					t.Errorf("%s: error %v, want error %q", tt.Name, err, tt.Error)
				} else if tt.Line != 0 && (tt.Line != perr.Line || tt.Column != perr.Column) {
					t.Errorf("%s: error at %d:%d expected %d:%d", tt.Name, perr.Line, perr.Column, tt.Line, tt.Column)
				}
			} else if err != nil {
				t.Errorf("%s: unexpected error %v", tt.Name, err)
			} else if !reflect.DeepEqual(out, tt.Output) {
				t.Errorf("%s: out=%q want %q", tt.Name, out, tt.Output)
			}
		}
	}
}
