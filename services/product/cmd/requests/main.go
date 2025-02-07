package main

import (
	"context"
	"fmt"
	product_v1 "github.com/dzhordano/ecom-thing/services/product/pkg/grpc/product/v1"
	"github.com/tjarratt/babble"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"math/rand/v2"
	"sync"
	"time"
)

func main() {
	conn, err := grpc.NewClient("0.0.0.0:50001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Println("failed to close connection:", err)
		}
	}()

	c := product_v1.NewProductServiceV1Client(conn)

	b := babble.NewBabbler()
	b.Separator = " "
	b.Count = 2

	_, err = c.CreateProduct(context.Background(), &product_v1.CreateProductRequest{
		Name:     b.Babble(),
		Desc:     "test",
		Category: "test",
		Price:    1.0,
	})

	timeout := int(50 * time.Millisecond)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for i := 0; i < 25000; i++ {
			_, err = c.CreateProduct(context.Background(), &product_v1.CreateProductRequest{
				Name:     b.Babble(),
				Desc:     "test",
				Category: "test",
				Price:    1.0,
			})
			if err != nil {
				fmt.Println("failed to create grpc:", err)
			}
			time.Sleep(time.Duration(rand.IntN(timeout)))
		}
	}()

	// go func() {
	// 	defer wg.Done()
	// 	for i := 0; i < 25000; i++ {
	// 		_, err = c.GetProduct(context.Background(), &product_v1.GetProductRequest{
	// 			Id: testP.Product.Id,
	// 		})
	// 		if err != nil {
	// 			fmt.Println("failed to get grpc:", err)
	// 		}
	// 		time.Sleep(time.Duration(rand.IntN(timeout)))
	// 	}
	// }()

	// go func() {
	// 	defer wg.Done()
	// 	for i := 0; i < 25000; i++ {
	// 		_, err = c.SearchProducts(context.Background(), &product_v1.SearchProductsRequest{
	// 			Limit:  rand.Uint64N(50),
	// 			Offset: 0,
	// 		})
	// 		if err != nil {
	// 			fmt.Println("failed to get products:", err)
	// 		}
	// 		time.Sleep(time.Duration(rand.IntN(timeout)))
	// 	}
	// }()

	// go func() {
	// 	for i := 0; i < 25000; i++ {
	// 		_, err = c.UpdateProduct(context.Background(), &product_v1.UpdateProductRequest{
	// 			Id:       testP.Product.Id,
	// 			Name:     b.Babble(),
	// 			Desc:     "test",
	// 			Category: "test",
	// 			Price:    1.0,
	// 		})
	// 		if err != nil {
	// 			fmt.Println("failed to update grpc:", err)
	// 		}
	// 		time.Sleep(time.Duration(rand.IntN(timeout)))
	// 	}
	// }()

	wg.Wait()

	fmt.Println("done")
}
