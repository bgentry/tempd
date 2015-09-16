package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"golang.org/x/net/context"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	sigch := make(chan os.Signal, 2)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
	go handleSignals(sigch, ctx, cancel)

	log.Println("hello world!")
	embd.InitGPIO()
	defer embd.CloseGPIO()

	pin, err := embd.NewDigitalPin("GPIO_5")
	if err != nil {
		log.Fatal("opening pin:", err)
	}
	defer resetPin(pin)

	if err = pin.SetDirection(embd.Out); err != nil {
		log.Fatal("setting pin direction:", err)
	}

	nextValHigh := true
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(500 * time.Millisecond):
			if nextValHigh {
				pin.Write(embd.High)
			} else {
				pin.Write(embd.Low)
			}
			nextValHigh = !nextValHigh
		}
	}
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

func resetPin(pin embd.DigitalPin) {
	if err := pin.SetDirection(embd.In); err != nil {
		log.Fatal("resetting pin:", err)
	}
	pin.Close()
}
