package main

import (
	"fmt"
	"os"
)

func main() {
	path, err := os.Executable()
	if err != nil {
		fmt.Println(err)
		os.Exit(1) // want "call os.Exit()"
	}
	fmt.Println(path)
	osExit(0)
}

func osExit(code int) {
	os.Exit(code)

}
