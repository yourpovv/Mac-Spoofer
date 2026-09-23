package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"mac-changer/internal/banner"
	"mac-changer/internal/elevate"
	"mac-changer/internal/wifi"
)

func main() {
	if !elevate.IsAdmin() {
		if err := elevate.RelaunchElevated(); err != nil {
			showFailure("get administrator rights", err)
			fmt.Println("Right-click mac-changer.exe -> Run as administrator, then click Yes.")
			os.Exit(1)
		}
		return
	}
	banner.SetTitle("MAC Changer")
	menuLoop()
}

func menuLoop() {
	reader := bufio.NewReader(os.Stdin)
	for {
		banner.Clear()
		banner.Show()
		showOptions()
		fmt.Printf("%s ", banner.Accent("Option:"))
		menuChoice, err := readLine(reader)
		if err != nil {
			showFailure("read your choice", err)
			return
		}

		switch menuChoice {
		case "4":
			os.Exit(0)
		case "1", "2", "3":
			banner.Clear()
			banner.Show()
		default:
			banner.Clear()
			banner.Show()
			fmt.Println(banner.Accent("Invalid option."))
			fmt.Println()
			fmt.Printf("%s ", banner.Accent("Press ENTER to go back:"))
			if _, err := reader.ReadString('\n'); err != nil {
				return
			}
			continue
		}

		switch menuChoice {
		case "1":
			spoofMac()
		case "2":
			restoreMac()
		case "3":
			checkMac()
		}
		fmt.Println()
		fmt.Printf("%s ", banner.Accent("Press ENTER to go back:"))
		if _, err := reader.ReadString('\n'); err != nil {
			return
		}
	}
}

func showOptions() {
	fmt.Println(banner.Accent("[1] Spoof Mac Address"))
	fmt.Println(banner.Accent("[2] Revert Spoof"))
	fmt.Println(banner.Accent("[3] Check Mac Address"))
	fmt.Println(banner.Accent("[4] Exit"))
	fmt.Println()
}

func spoofMac() {
	wifiAdapter, err := wifi.Detect()
	if err != nil {
		showFailure("find your Wi-Fi adapter", err)
		return
	}
	fmt.Printf("Adapter: %s (%s)\n", wifiAdapter.ConnectionName, wifiAdapter.DriverDesc)
	oldMac := wifi.CurrentMac(wifiAdapter.ConnectionName)
	fmt.Printf("Current: %s\n", oldMac)
	fmt.Println()

	spoofedMac, err := wifi.RandomMac()
	if err != nil {
		showFailure("generate a random MAC", err)
		return
	}

	if err := wifi.ApplySpoof(wifiAdapter, spoofedMac); err != nil {
		showFailure("spoof your MAC", err)
		return
	}
	fmt.Printf("Spoofed: %s -> %s\n", oldMac, wifi.CurrentMac(wifiAdapter.ConnectionName))
}

func restoreMac() {
	wifiAdapter, err := wifi.Detect()
	if err != nil {
		showFailure("find your Wi-Fi adapter", err)
		return
	}
	fmt.Printf("Adapter: %s (%s)\n", wifiAdapter.ConnectionName, wifiAdapter.DriverDesc)
	fmt.Printf("Current: %s\n", wifi.CurrentMac(wifiAdapter.ConnectionName))
	fmt.Println()

	if err := wifi.Revert(wifiAdapter); err != nil {
		showFailure("restore your MAC", err)
		return
	}
	fmt.Printf("Restored hardware MAC: %s\n", wifi.CurrentMac(wifiAdapter.ConnectionName))
}

func checkMac() {
	wifiAdapter, err := wifi.Detect()
	if err != nil {
		showFailure("find your Wi-Fi adapter", err)
		return
	}
	fmt.Printf("Adapter: %s (%s)\n", wifiAdapter.ConnectionName, wifiAdapter.DriverDesc)
	fmt.Printf("Current: %s\n", wifi.CurrentMac(wifiAdapter.ConnectionName))
	if spoofed, active := wifi.SpoofedMac(wifiAdapter); active {
		fmt.Printf("Spoof: %s (override active)\n", wifi.DisplayMac(spoofed))
	} else {
		fmt.Println("Spoof: off, using hardware MAC")
	}
}

func showFailure(action string, err error) {
	fmt.Printf("Couldn't %s: %s\n", action, err)
}

func readLine(reader *bufio.Reader) (string, error) {
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}
	return strings.TrimSpace(text), nil
}
