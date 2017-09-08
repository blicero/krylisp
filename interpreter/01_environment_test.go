// /home/krylon/go/src/krylisp/interpreter/01_environment_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-08 21:53:39 krylon>

package interpreter

import (
	"krylisp/value"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

var myEnv *Environment

func TestFreshEnv(t *testing.T) {
	// I know it feels kind of silly to test for the obvious.
	// But the purpose of automated testing is mainly to catch
	// errors that creep in later on. If I ever change the creation
	// or initialization of Environments, this might help
	// catch bugs.
	if myEnv = NewEnvironment(nil); myEnv == nil {
		t.Fatal("NewEnvironment must not return nil!")
	} else if myEnv.Level != 0 {
		t.Fatalf("Level of Environment without parent should be 0, not %d!",
			myEnv.Level)
	} else if myEnv.Parent != nil {
		t.Fatal("Parent of fresh Environment must be nil!")
	}
} // func TestFreshEnv(t *testing.T)

func TestSet(t *testing.T) {
	myEnv.Set("x", value.IntValue(10))
	myEnv.Set("y", value.StringValue("Hallo, Welt"))

	if v, ok := myEnv.Get("x"); !ok {
		t.Fatalf("Could not find key \"x\" after I just set it to 10!")
	} else if i, ok := v.(value.IntValue); !ok {
		t.Fatalf("Lookup of key \"x\" should have return an IntValue, not a %T (%s)",
			v,
			spew.Sdump(v))
	} else if i != 10 {
		t.Fatalf("Binding for \"x\" should be an IntValue of 10, not %d",
			i)
	}

	if v, ok := myEnv.Get("y"); !ok {
		t.Fatalf("Could not find key \"y\" after I just set it")
	} else if s, ok := v.(value.StringValue); !ok {
		t.Fatalf("Lookup of key \"y\" should have return a StringValue, not a %T (%s)",
			v,
			spew.Sdump(v))
	} else if s != "Hallo, Welt" {
		t.Fatalf("Binding for y should be the string \"Hallo, Welt\", not %s",
			s)
	}
} // func TestSet(t *testing.T)

func TestLookupChain(t *testing.T) {
	var child = NewEnvironment(myEnv)

	if v, ok := child.Get("x"); !ok {
		t.Fatalf("Could not find key \"x\" after I just set it to 10!")
	} else if i, ok := v.(value.IntValue); !ok {
		t.Fatalf("Lookup of key \"x\" should have return an IntValue, not a %T (%s)",
			v,
			spew.Sdump(v))
	} else if i != 10 {
		t.Fatalf("Binding for \"x\" should be an IntValue of 10, not %d",
			i)
	}

	if v, ok := child.Get("y"); !ok {
		t.Fatalf("Could not find key \"y\" after I just set it")
	} else if s, ok := v.(value.StringValue); !ok {
		t.Fatalf("Lookup of key \"y\" should have return a StringValue, not a %T (%s)",
			v,
			spew.Sdump(v))
	} else if s != "Hallo, Welt" {
		t.Fatalf("Binding for y should be the string \"Hallo, Welt\", not %s",
			s)
	}
} // func TestLookupChain(t *testing.T)

func TestSetChain(t *testing.T) {
	var child = NewEnvironment(myEnv)

	child.Set("x", value.IntValue(42))

	if _, ok := child.Data["x"]; ok {
		t.Fatal("Binding should have been set in parent Environment, not in current")
	}

	child.Ins("x", value.IntValue(128))

	if v, ok := child.Get("x"); !ok {
		t.Fatal("After inserting new binding for \"x\", no binding for \"x\" was found")
	} else if i, ok := v.(value.IntValue); !ok {
		t.Fatalf("Binding for \"x\" should be an IntValue, not a %T (%s)",
			v,
			spew.Sdump(v))
	} else if i != 128 {
		t.Fatalf("Expected vaue for \"x\" after masking previous binding was 128, not %d",
			i)
	}

	child.Del("x")

	if v, ok := child.Get("x"); !ok {
		t.Fatal("After deleting new binding for \"x\", no binding for \"x\" was found")
	} else if i, ok := v.(value.IntValue); !ok {
		t.Fatalf("Binding for \"x\" should be an IntValue, not a %T (%s)",
			v,
			spew.Sdump(v))
	} else if i != 42 {
		t.Fatalf("Expected vaue for \"x\" after masking previous binding was 42, not %d",
			i)
	}
} // func TestSetChain(t *testing.T)
