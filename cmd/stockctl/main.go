package main

import (
	"fmt"

	"stock/pkg/version"
)

func main() {
	fmt.Printf("stockctl %s\n", version.GetShortVersion())
}
