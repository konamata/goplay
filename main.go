package main

/*
#include "hello.h"

#cgo CFLAGS: -I.
#cgo LDFLAGS: hello.o
*/
import "C"

func main() {
	C.sayHello()
}
