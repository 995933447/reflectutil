package reflectutil

import (
	"fmt"
	"testing"
)

func TestCopySameFields(t *testing.T) {
	type A struct {
		A string
		b string
		B uint64
		C bool
		D func()
	}

	type B struct {
		A string
		C bool
		D func()
	}

	a := &A{
		A: "123",
		b: "234",
		B: 90,
		C: true,
		D: func() {
			fmt.Println("Hello world")
		},
	}

	var b B = B {

	}
	if err := CopySameFields(a, &b); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%+v\n", b)

	b.D()
}
