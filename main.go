package main

import (
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"golang.org/x/net/context"
)

const (
	channel  = 0
	speed    = 1000000
	bpw      = 8
	delay    = 0
	startBit = 1
	mode     = embd.SPIMode0

	sampleCount = 5

	maxADCVal = 1023  // maximum value of 10-bit ADC
	rref      = 10000 // reference resistor resistance in Ohms

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

	log.Println("hello world!")
	if err := embd.InitSPI(); err != nil {
		log.Fatal("initializing SPI:", err)
		panic(err)
	}
	defer embd.CloseSPI()
	log.Println("SPI initialized.")

	bus := embd.NewSPIBus(mode, channel, speed, bpw, delay)
	defer bus.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
			val, err := getTempF(bus, 0)
			if err != nil {
				log.Fatal("reading value: ", err)
			}
			log.Printf("value for chan %d is: %.2f", channel, val)
		}
	}
}

func getTempF(bus embd.SPIBus, chanNum int) (float64, error) {
	vals := make([]uint16, sampleCount)
	total := 0.0
	for i := range vals {
		adcval, err := getRawSensorValue(bus, chanNum)
		if err != nil {
			return 0, err
		}
		vals[i] = adcval
		total += float64(adcval)
	}

	meanVal := total / sampleCount

	// resistance of thermistor rt
	rt := rref / ((float64(maxADCVal) / float64(meanVal)) - 1)
	log.Printf("rt for chan %d is: %.2f", chanNum, rt)

	tempk := getTempKFromRt(rt)
	log.Printf("tempk for chan %d is: %.2f", chanNum, tempk)
	return convertTempKToF(tempk), nil
}

func getTempKFromRt(rt float64) float64 {
	return 1 / (c1 + c2*math.Log(rt) + c3*(math.Pow(math.Log(rt), 3)))
}

func convertTempKToF(tempk float64) float64 {
	return (tempk-273.15)*(9.0/5) + 32
}

// getRawSensorValue returns the analog value at the given channel of the convertor.
func getRawSensorValue(bus embd.SPIBus, chanNum int) (uint16, error) {
	data := [3]byte{startBit, byte(8+chanNum) << 4, 0}

	var err error
	err = bus.TransferAndRecieveData(data[:])
	if err != nil {
		return 0, err
	}
	return uint16(data[1]&0x03)<<8 | uint16(data[2]), nil
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
