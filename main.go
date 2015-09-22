package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bgentry/tempd/device"
	"github.com/bgentry/tempd/temperature"
	"golang.org/x/net/context"
)

const (
	spiChannel = 0

	sampleCount = 5

	rref = 10000 // reference resistor resistance in Ohms

	sensorChannel = 9

	// constants for Steinhart-hart equation, iGrill probe, hand-measured
	c1 = 0.761793296025725e-3
	c2 = 2.114881554906883e-4
	c3 = 1.0244107975830052e-7

	// iGrill constants from forum post
	// c1 = 0.7739251279e-3
	// c2 = 2.088025997e-4
	// c3 = 1.154400438e-7
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	sigch := make(chan os.Signal, 2)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
	go handleSignals(sigch, ctx, cancel)

	if err := device.Init(); err != nil {
		log.Fatal("initializing device:", err)
	}
	defer device.Close()

	dev := device.New(spiChannel)
	defer dev.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
			val, err := getTempF(dev, 0)
			if err != nil {
				log.Fatal("reading value: ", err)
			}
			log.Printf("value for chan %d is: %.2f", sensorChannel, val)
		}
	}
}

func getTempF(dev *device.Device, chanNum int) (float64, error) {
	vals := make([]float64, sampleCount)
	for i := range vals {
		adcval, err := dev.ReadSensor(chanNum)
		if err != nil {
			return 0, err
		}
		vals[i] = adcval
	}

	// resistance of thermistor rt
	rt := rref / (mean(vals) - 1)
	log.Printf("rt for chan %d is: %.2f", chanNum, rt)

	tempk := temperature.SteinhartTemp(c1, c2, c3, rt)
	log.Printf("tempk for chan %d is: %.2f", chanNum, tempk)
	return temperature.KToF(tempk), nil
}

func handleSignals(sigch <-chan os.Signal, ctx context.Context, cancel context.CancelFunc) {
	select {
	case <-ctx.Done():
	case sig := <-sigch:
		switch sig {
		case os.Interrupt:
			log.Println("SIGINT")
		case syscall.SIGTERM:
			log.Println("SIGTERM")
		}
		cancel()
	}
}

func mean(vals []float64) float64 {
	total := 0.0
	for _, val := range vals {
		total += float64(val)
	}
	return total / float64(len(vals))
}
