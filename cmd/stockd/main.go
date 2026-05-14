package main

import (
	"fmt"

	"stock/pkg/version"
)

func main() {
	fmt.Printf("stockd %s\n", version.GetShortVersion())
}
