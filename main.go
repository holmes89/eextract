package main

import (
	"fmt"

	"github.com/holmes89/eextract/lib"
)

func main() {
	entities := lib.FindEntities("./example/thing.pb.go")
	fmt.Println("====== Structs ======")
	for _, e := range entities {
		fmt.Println(e)
	}
	fmt.Println("=====================\n\n")
	services := lib.FindServices("./example/thing_grpc.pb.go")
	fmt.Println("====== Interfaces ===== ")
	for _, s := range services {
		fmt.Println(s)
	}
}
