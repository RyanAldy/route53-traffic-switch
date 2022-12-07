package main

import "fmt"

func main() {
	a, err := New()

	if err != nil {
		fmt.Printf("Error running application due to %s", err)
	}

	a.handler()
}
