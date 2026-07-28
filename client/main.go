package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/fatih/color"
)

func main() {
	colorGreen := color.New(color.FgGreen)
	colorWhite := color.New(color.FgWhite)

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	productChan := make(chan string, 1)
	orderChan := make(chan string, 1)

	go func() {
		defer wg.Done()
		runClientAndCollect(filepath.Join(currentDir, "product"), "product_client.go", "PRODUCTS", productChan)
	}()

	go func() {
		defer wg.Done()
		runClientAndCollect(filepath.Join(currentDir, "order"), "ordermgt_client.go", "ORDERS", orderChan)
	}()

	go func() {
		wg.Wait()
		close(productChan)
		close(orderChan)
	}()

	var productOutput, orderOutput string
	received := 0

	for received < 2 {
		select {
		case output, ok := <-productChan:
			if ok {
				productOutput = output
				received++
			}
		case output, ok := <-orderChan:
			if ok {
				orderOutput = output
				received++
			}
		}
	}

	if productOutput != "" {
		colorGreen.Println("INFORMATION ABOUT PRODUCTS")
		colorWhite.Println(productOutput)
	}

	if orderOutput != "" {
		colorGreen.Println("INFORMATION ABOUT ORDERS")
		colorWhite.Println(orderOutput)
	}
}

func runClientAndCollect(dir string, fileName string, label string, ch chan<- string) {
	cmd := exec.Command("go", "run", fileName)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		ch <- fmt.Sprintf("Error: %v\n%s", err, string(output))
		return
	}
	ch <- string(output)
}
