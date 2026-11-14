package main

import (
	"fmt"
	"log"

	"github.com/MJKhaani/ESP32-Marauder/internal/app"
	"github.com/MJKhaani/ESP32-Marauder/internal/serialtransport"
)

func main() {
	fmt.Println("Marauder Studio CLI Core")
	appInstance, err := app.NewApp(serialtransport.NewSystemPortEnumerator(), "firmware/catalog/targets.json", "marauder_studio.db")
	if err != nil {
		log.Fatalf("Initialization failed: %v", err)
	}

	ports, err := appInstance.ScanPorts()
	if err != nil {
		log.Fatalf("Failed to scan ports: %v", err)
	}

	fmt.Printf("Discovered %d candidate ports\n", len(ports))
	for _, p := range ports {
		fmt.Printf(" - %s [%s:%s] %s (%s)\n", p.Port.Path, p.Port.VID, p.Port.PID, p.Port.Product, p.State)
	}
}
