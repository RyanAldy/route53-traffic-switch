package main

import (
	"fmt"
	"os"
)

func main() {
	config, err := New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	a, err := NewApp(*config)

	if err != nil {
		fmt.Printf("Error running application due to %s", err)
	}

	message, err := a.handler()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(message)
}
