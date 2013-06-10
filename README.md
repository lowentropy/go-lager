go-lager
========

Serialize Go data with cyclic pointers

[Documentation](http://godoc.org/github.com/lowentropy/go-lager)

Purpose
=======

Go's [gob](http://golang.org/pkg/encoding/gob/) encoding packs and unpacks data structures from a stream of bytes.
It handles pointers by recursively writing the object pointed to. This means identical pointers, when read back,
are no longer identical, and there will be multiple copies of some objects. It also cannot handle recursive pointers.

go-lager solves this by only writing each object once. Pointers are not initialized until every object has been read.

Usage
=====

```go
import lager "github.com/lowentropy/go-lager"
struct Foo { Ptr *Foo }
```

Writing
-------

```go
foo := new(Foo)                           // make an object to encode
foo.Ptr = foo                             // you can have self-references
writer := ...                             // this is your output stream
encoder := lager.NewEncoder(writer)       // create the encoder
encoder.Write(foo)                        // write the object to the stream
encoder.Finish()                          // flush and terminate encoding
```

Reading
-------

```go
lager.Register(Foo{})                     // register the Foo type
reader := ...                             // this is your input stream
decoder, err := lager.NewDecoder(reader)  // create the decoder
thing, err := decoder.Read()              // read the next object
foo := thing.(*Foo)                       // cast to static type
```

Encoding Details
================

TODO

Caveats
=======

This is pre-alpha quality software and is emphatically NOT ready for production use!

go-lager's binary format is also not as space-efficient or flexible as gob's.

Like gob, only exported struct fields (the ones that start with an upper-case letter) are encoded.

TODO
====

 * Document encoding
 * Make better use of bufio
