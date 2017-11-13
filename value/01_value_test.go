// /home/krylon/go/src/krylisp/value/value_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-11-13 21:13:36 krylon>

package value

import (
	"fmt"
	"krylisp/permission"
	"krylisp/types"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestTypeID(t *testing.T) {
	type testValue struct {
		input        LispValue
		expectedType types.ID
	}

	var values = []testValue{
		testValue{
			input:        IntValue(42),
			expectedType: types.Integer,
		},
		testValue{
			input:        StringValue("Wer das liest, ist doof."),
			expectedType: types.String,
		},
		testValue{
			input:        &ConsCell{nil, nil},
			expectedType: types.ConsCell,
		},
	}

	for idx, v := range values {
		if v.input.Type() != v.expectedType {
			t.Errorf("Wrong type ID for testValue %d: %s (expected %s)",
				idx,
				v.input.Type().String(),
				v.expectedType.String())

		}
	}
} // func TestTypeID(t *testing.T)

func TestString(t *testing.T) {
	type testValue struct {
		input    LispValue
		expected string
	}

	var testCases = []testValue{
		testValue{
			IntValue(42),
			"42",
		},
		testValue{
			StringValue("Wer das liest, ist doof."),
			"\"Wer das liest, ist doof.\"",
		},
		testValue{
			&ConsCell{IntValue(23), IntValue(42)},
			"(23 . 42)",
		},
		testValue{
			&ConsCell{IntValue(64), nil},
			"(64 . NIL)",
		},
	}

	for idx, val := range testCases {
		var s = val.input.String()

		if s != val.expected {
			t.Errorf("Unexpected string value for input %d: Expected \"%s\", got \"%s\"",
				idx,
				val.expected,
				s)
		}
	}
} // func TestString(t *testing.T)

func TestIsList(t *testing.T) {
	type testValue struct {
		input    *ConsCell
		expected bool
	}

	var testData = []testValue{
		testValue{
			&ConsCell{nil, nil},
			true,
		},
		testValue{ // ???
			nil,
			true,
		},
		testValue{
			&ConsCell{
				Car: IntValue(42),
				Cdr: &ConsCell{
					Car: IntValue(64),
					Cdr: IntValue(128),
				},
			},
			false,
		},
	}

	for idx, val := range testData {
		var res = val.input.IsList()

		if res != val.expected {
			t.Errorf("Error in test data #%d: Expected %t, got %t",
				idx,
				val.expected,
				res)
		}
	}
} // func TestIsList(t *testing.T)

func TestListString(t *testing.T) {
	type testValue struct {
		input    *List
		expected string
	}

	var testData = []testValue{
		testValue{
			input:    &List{},
			expected: "NIL",
		},
		testValue{
			input: &List{
				Car:    &ConsCell{IntValue(1), nil},
				Length: 1,
			},
			expected: "(1)",
		},
		testValue{
			input: &List{
				Car: &ConsCell{
					Car: IntValue(42),
					Cdr: &ConsCell{Car: IntValue(23)},
				},
				Length: 2,
			},
			expected: "(42 23)",
		},
		testValue{
			input: &List{
				Car: &ConsCell{
					Car: IntValue(1),
					Cdr: &ConsCell{
						Car: StringValue("Hallo"),
						Cdr: &ConsCell{
							Car: &List{
								Car: &ConsCell{
									Car: IntValue(-2),
									Cdr: &ConsCell{
										Car: IntValue(107),
										Cdr: nil,
									},
								},
								Length: 2,
							},
							Cdr: &ConsCell{
								Car: IntValue(42),
								Cdr: nil,
							},
						},
					},
				},
				Length: 4,
			},
			expected: `(1 "Hallo" (-2 107) 42)`,
		},
	}

	for idx, val := range testData {
		var s = val.input.String()

		if s != val.expected {
			t.Errorf("Invalid string (%d): Expected \"%s\", got \"%s\"",
				idx,
				val.expected,
				s)
		}
	}
} // func TestListString(t *testing.T)

// nolint: gocyclo
func TestListPush(t *testing.T) {
	var l = &List{}
	var cnt int

	l.Push(StringValue("doof"))
	if l.Length != 1 {
		t.Fatalf("Unexpected length after pushing one value: %d (expected 1)",
			l.Length)
	} else if cnt = l.ActualLength(); cnt != l.Length {
		t.Fatalf("Length field and actual count are different: Length = %d, Actual = %d",
			l.Length,
			cnt)
	}

	l.Push(StringValue("war"))
	if l.Length != 2 {
		t.Fatalf("Unexpected length after pushing second value: %d (expected 2)",
			l.Length)
	} else if cnt = l.ActualLength(); cnt != l.Length {
		t.Fatalf("Length field and actual count are different: Length = %d, Actual = %d",
			l.Length,
			cnt)
	}

	var v = l.Pop()

	if v == nil {
		t.Fatal("l.Pop returned nil!")
	} else if v.Type() != types.String {
		t.Fatalf("l.Pop returned unexpected type: %s (expected String)",
			v.Type().String())
	} else if v.(StringValue) != "war" {
		t.Fatalf("l.Pop returned unexpected value: %s (expected \"war\")",
			v)
	} else if l.Length != 1 {
		t.Fatalf("Unexpected length after popping one value: %d (expected 1)",
			l.Length)
	} else if cnt = l.ActualLength(); cnt != l.Length {
		t.Fatalf("Length field and actual count are different: Length = %d, Actual = %d",
			l.Length,
			cnt)
	}

	l.Push(StringValue("ist"))
	if l.Length != 2 {
		t.Fatalf("Unexpected length after pushing another value: %d (expected 2)",
			l.Length)
	} else if cnt = l.ActualLength(); cnt != l.Length {
		t.Fatalf("Length field and actual count are different: Length = %d, Actual = %d",
			l.Length,
			cnt)
	}

	for _, item := range []string{"liest", "das", "Wer"} {
		l.Push(StringValue(item))
	}

	if l.Length != 5 {
		t.Fatalf("Unexpected length after pushing remaining values: %d (expected 5)",
			l.Length)
	} else if cnt = l.ActualLength(); cnt != l.Length {
		t.Fatalf("Length field and actual count are different: Length = %d, Actual = %d",
			l.Length,
			cnt)
	}
} // func TestListPush(t *testing.T)

func TestListNth(t *testing.T) {
	type nthTest struct {
		list          *List
		index         int
		expectedValue LispValue
		expectedError bool
	}

	var testCases = []nthTest{
		nthTest{
			list:          nil,
			index:         3,
			expectedValue: nil,
			expectedError: true,
		},
		nthTest{
			list: &List{
				Car: &ConsCell{
					Car: IntValue(1),
					Cdr: &ConsCell{
						Car: IntValue(2),
						Cdr: NIL,
					},
				},
				Length: 2,
			},
			index:         1,
			expectedValue: IntValue(2),
		},
	}

	for idx, test := range testCases {
		var val LispValue
		var err error

		if val, err = test.list.Nth(test.index); err != nil {
			if !test.expectedError {
				t.Errorf("Unexpected error in test case #%d: %s",
					idx+1,
					err.Error())
			}
		} else if !test.expectedValue.Eq(val) {
			t.Errorf("Unexpected value returned from list: %s (expected %s)",
				val.String(),
				test.expectedValue.String())
		}
	}
} // func TestListNth(t *testing.T)

func TestFileHandle(t *testing.T) {
	var (
		fh                     *FileHandle
		err                    error
		path, filename, folder string
		status                 bool
	)

	folder = os.TempDir()
	filename = time.Now().Format("kryLisp_test_filehandle_20060102_150405")
	path = filepath.Join(folder, filename)

	if fh, err = OpenFile(path, 0600, permission.Read|permission.Write); err != nil {
		t.Fatalf("Error opening test file %s: %s",
			path,
			err.Error())
	} else {
		defer func() {
			if status {
				os.Remove(path)
			}
		}()
	}

	const nCount = 100

	var numbers = make([]int64, nCount)

	for idx := range numbers {
		var (
			n       = rand.Int63n(65536)
			nstring = fmt.Sprintf("%d\n", n)
		)

		if _, err = fh.Write([]byte(nstring)); err != nil {
			t.Fatalf("Error writing line no. %d to test file %s: %s",
				idx,
				path,
				err.Error())
		} else {
			numbers[idx] = n
		}
	}

	var pos int64

	if pos, err = fh.Seek(0, 0); err != nil {
		t.Fatalf("Error seeking to beginning of file: %s",
			err.Error())
	} else if pos != 0 {
		t.Errorf("Seek to beginning of file returned wrong offset %d. SRSLY?!?!",
			pos)
	}

	for _, val := range numbers {
		var (
			n       int64
			nstring string
		)

		if nstring, err = fh.ReadLine(); err != nil {
			t.Fatalf("Error reading line from %s: %s",
				path,
				err.Error())
		} else if n, err = strconv.ParseInt(strings.TrimRight(nstring, "\n"), 10, 64); err != nil {
			t.Errorf("Error reading number from string \"%s\": %s",
				nstring,
				err.Error())
			continue
		} else if n != val {
			t.Fatalf("Read wrong number from file: Expected %d, got %d",
				val,
				n)
		}
	}

	if err = fh.Close(); err != nil {
		t.Errorf("Error _CLOSING_ %s: %s",
			path,
			err.Error())
	} else {
		status = true
	}
} // func TestFileHandle(t *testing.T)

///////////////////
//// Utilities ////
///////////////////
