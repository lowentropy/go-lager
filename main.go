package main

import (
	"bufio"
	"fmt"
	"os"
)

type payload struct {
	Str string
	num int
}

type Cyclic struct {
	Parent *Cyclic
	payload
}

func main() {
	// set up data: two structs pointing to each other
	a := &Cyclic{nil, payload{"foo", 3}}
	b := &Cyclic{a, payload{"bar", 5}}
	a.Parent = b

	// encode to a file
	file, _ := os.Create("out.bin")
	enc := NewEncoder(file)
	enc.Write(a)
	enc.Finish()
	file.Close()

	// read back from file
	file, _ = os.Open("out.bin")
	dec, _ := NewDecoder(bufio.NewReader(file))
	raw, _ := dec.Read()
	ptr := raw.(*Cyclic)

	// show off the cyclic pointerse
	for i := 0; i < 10; i++ {
		fmt.Println(ptr.payload.Str)
		ptr = ptr.Parent
	}
}
