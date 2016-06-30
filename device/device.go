package device

import (
	"fmt"

	"golang.org/x/exp/io/spi"
)

const (
	speed = 1000000
	bpw   = 8
	delay = 0
	mode  = spi.Mode0

	startBit = 1

	adcBits   = 10                       // number of bits of resolution in the ADC
	maxADCVal = (2 << (adcBits - 1)) - 1 // maximum value of 10-bit ADC
)

type Device struct {
	device *spi.Device
}

func Open(channel uint8) (*Device, error) {
	spidev, err := spi.Open(&spi.Devfs{
		Dev:      fmt.Sprintf("/dev/spidev0.%d", channel),
		Mode:     mode,
		MaxSpeed: speed,
	})
	if err != nil {
		return nil, err
	}
	return &Device{spidev}, nil
}

func (d *Device) Close() error {
	return d.device.Close()
}

// ReadSensor returns the analog value of the given channel of the convertor,
// normalized for the number of bits of resolution. This results in a value
// in the range [0, 1].
func (d *Device) ReadSensor(channel uint) (float64, error) {
	val, err := getRawSensorValue(d.device, channel)
	if err != nil {
		return 0, err
	}
	return maxADCVal/float64(val) - 1, nil
}

func getRawSensorValue(dev *spi.Device, chanNum uint) (uint16, error) {
	outbound := [3]byte{startBit, byte(8+chanNum) << 4, 0}
	inbound := make([]byte, 3)

	var err error
	err = dev.Tx(outbound[:], inbound)
	if err != nil {
		return 0, err
	}
	return uint16(inbound[1]&0x03)<<8 | uint16(inbound[2]), nil
}
