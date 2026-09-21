package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"

	bluetooth "tinygo.org/x/bluetooth"
)

const appleID = 0x004C
const iBeaconType = 0x02

func scanIBeacon(data []byte) (uuid string, major uint16, minor uint16, txPower int8, ok bool) {
	if len(data) < 23 {
		return
	}

	if data[0] != iBeaconType {
		return
	}

	uuidBytes := data[2:18]
	uuid = fmt.Sprintf("%X-%X-%X-%X-%X",
		uuidBytes[0:4],
		uuidBytes[4:6],
		uuidBytes[6:8],
		uuidBytes[8:10],
		uuidBytes[10:16],
	)

	major = binary.BigEndian.Uint16(data[18:20])
	minor = binary.BigEndian.Uint16(data[20:22])
	txPower = int8(data[22])
	ok = true
	return
}

func main() {
	adapter := bluetooth.DefaultAdapter
	must("enable BLE stack", adapter.Enable())

	fmt.Println("Scanning for nearby iBeacons...")

	seen := make(map[string]bool)

	err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		for _, elem := range device.AdvertisementPayload.ManufacturerData() {
			if elem.CompanyID != appleID {
				continue
			}

			if uuid, major, minor, tx, ok := scanIBeacon(elem.Data); ok {
				addr := device.Address.String()

				if seen[addr] {
					continue
				}
				seen[addr] = true

				n := 2.0
				distance := math.Pow(10, float64(int16(tx)-device.RSSI)/(10*n))
				fmt.Printf("UUID: %s\n", uuid)
				fmt.Printf("Major Value: %d\n", major)
				fmt.Printf("Minor Value: %d\n", minor)
				fmt.Printf("TX Power Value: %d dBm\n", tx)
				fmt.Printf("RSSI (Received Signal Strength Indication): %d dBm\n", device.RSSI)
				fmt.Printf("Approximate Distance: %.2f meters\n", distance)
				fmt.Printf("MAC Address: %s\n", addr)
			}
		}
	})

	must("scan", err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}

func must(action string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed with error message: %s %v\n", action, err)
		os.Exit(1)
	}
}
