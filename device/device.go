package device

import (
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
)

const (
	speed = 1000000
	bpw   = 8
	delay = 0
	mode  = embd.SPIMode0

	startBit = 1

	adcBits   = 10                       // number of bits of resolution in the ADC
	maxADCVal = (2 << (adcBits - 1)) - 1 // maximum value of 10-bit ADC
)

func Init() error {
	return embd.InitSPI()
}

func Close() error {
	return embd.CloseSPI()
}

type Device struct {
	bus embd.SPIBus
}

func New(channel uint8) *Device {
	return &Device{embd.NewSPIBus(mode, channel, speed, bpw, delay)}
}

func (d *Device) Close() error {
	return d.bus.Close()
}

// ReadSensor returns the analog value of the given channel of the convertor,
// normalized for the number of bits of resolution. This results in a value
// in the range [0, 1].
func (d *Device) ReadSensor(channel uint) (float64, error) {
	val, err := getRawSensorValue(d.bus, channel)
	if err != nil {
		return 0, err
	}
	return maxADCVal/float64(val) - 1, nil
}

func getRawSensorValue(bus embd.SPIBus, chanNum uint) (uint16, error) {
	data := [3]byte{startBit, byte(8+chanNum) << 4, 0}

	var err error
	err = bus.TransferAndRecieveData(data[:])
	if err != nil {
		return 0, err
	}
	return uint16(data[1]&0x03)<<8 | uint16(data[2]), nil
}
