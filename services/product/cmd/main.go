package main

import (
	"fmt"
	"github.com/dzhordano/ecom-thing/services/product/internal/config"
)

func main() {
	cfg := config.MustNew()

	fmt.Printf("config: %+v", cfg)
}
