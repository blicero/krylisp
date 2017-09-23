// /home/krylon/go/src/krylisp/value/value_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-22 17:06:50 krylon>

package value

import (
	"krylisp/types"
	"testing"
)

func TestTypeID(t *testing.T) {
	type testValue struct {
		input        LispValue
		expectedType types.ID
	}

	var values = []testValue{
		testValue{
			input:        IntValue(42),
			expectedType: types.Number,
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

///////////////////
//// Utilities ////
///////////////////
