package main

/*
#include "hello.h"
*/
import "C"

func main() {
	// C'deki sayHello fonksiyonunu çağır
	C.sayHello()
}
