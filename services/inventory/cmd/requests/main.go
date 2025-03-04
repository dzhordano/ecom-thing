package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	inventory_v1 "github.com/dzhordano/ecom-thing/services/inventory/pkg/api/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	conn, err := grpc.NewClient("localhost:50001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Println("failed to close connection:", err)
		}
	}()

	c := inventory_v1.NewInventoryServiceClient(conn)

	// t, ccl := context.WithTimeout(context.Background(), 2*time.Second)
	// defer ccl()

	_, err = c.SetItem(context.Background(), &inventory_v1.SetItemRequest{
		Item: &inventory_v1.ItemOP{
			ProductId: "00000000-0000-0000-0000-000000000000",
			Quantity:  1,
		},
		OperationType: inventory_v1.OperationType_OPERATION_TYPE_ADD,
	})
	if err != nil {
		panic(err)
	}

	timeout := int(10 * time.Millisecond)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {

		defer wg.Done()
		for i := 0; i < 25000; i++ {
			_, err = c.SetItem(context.Background(), &inventory_v1.SetItemRequest{
				Item: &inventory_v1.ItemOP{
					ProductId: "00000000-0000-0000-0000-000000000000",
					Quantity:  1,
				},
				OperationType: inventory_v1.OperationType_OPERATION_TYPE_ADD,
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

	item, err := c.GetItem(context.Background(), &inventory_v1.GetItemRequest{
		Id: "00000000-0000-0000-0000-000000000000",
	})
	if err != nil {
		fmt.Println("failed to get item:", err)
	}

	fmt.Println("ITEM", item)

	wg.Wait()

	fmt.Println("done")
}
