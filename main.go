package main

import (
	"errors"
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
	spiChannel  = 0     // which SPI bus channel to use
	sampleCount = 5     // number of samples per measurement
	rref        = 10000 // reference resistor resistance in Ohms

	// constants for Steinhart-hart equation, iGrill probe, hand-measured
	c1 = 0.761793296025725e-3
	c2 = 2.114881554906883e-4
	c3 = 1.0244107975830052e-7

	// iGrill constants from forum post
	// c1 = 0.7739251279e-3
	// c2 = 2.088025997e-4
	// c3 = 1.154400438e-7

	tempChannelID0 = 0
	tempChannelID1 = 1
	tempChannelID2 = 2
)

var tempChannelIDs = []uint{tempChannelID0, tempChannelID1, tempChannelID2}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	sigch := make(chan os.Signal, 2)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
	go handleSignals(sigch, ctx, cancel)

	dev, err := device.Open(spiChannel)
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	tempch := make(chan tempReading, 5*len(tempChannelIDs))
	go readTemps(ctx, dev, tempch)
	reportReadings(tempch)
}

var errProbeDisconnected = errors.New("probe is not connected")

func getTempF(dev *device.Device, chanNum uint) (float64, error) {
	vals := make([]float64, sampleCount)
	for i := range vals {
		adcval, err := dev.ReadSensor(chanNum)
		if err != nil {
			return 0, err
		}
		if adcval == 0 {
			return 0, errProbeDisconnected
		}
		vals[i] = adcval
	}

	// calculate resistance of thermistor rt as a voltage divider
	rt := rref / mean(vals)
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

type tempReading struct {
	channel     uint
	temperature float64
	timestamp   time.Time
}

func reportReadings(tempch <-chan tempReading) {
	for reading := range tempch {
		log.Printf("value for chan %d is: %.2f", reading.channel, reading.temperature)
	}
}

func readTemps(ctx context.Context, dev *device.Device, tempch chan<- tempReading) {
	for {
		select {
		case <-ctx.Done():
			close(tempch)
			return
		case <-time.After(2 * time.Second):
			for _, chanID := range tempChannelIDs {
				val, err := getTempF(dev, chanID)
				if err != nil {
					if err == errProbeDisconnected {
						log.Printf("probe for channel %d is disconnected", chanID)
						continue
					}
					log.Fatal("reading value: ", err)
				}
				tempch <- tempReading{
					timestamp:   time.Now().UTC(),
					channel:     chanID,
					temperature: val,
				}
			}
		}
	}
}
