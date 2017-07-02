package max30105

import (
	"bytes"
	"encoding/binary"
	"time"
)

func (d *MAX30105Driver) GetWritePointer() (byte, error) {
	return d.bus.ReadByteFromReg(d.address, FIFOWriterPtr)
}

func (d *MAX30105Driver) GetReadPointer() (byte, error) {
	return d.bus.ReadByteFromReg(d.address, FIFOReadPtr)
}

func (d *MAX30105Driver) ReadTemperature() (float64, error) {
	d.bus.WriteByteToReg(d.address, DieTempConfig, 0x01)

	time.Sleep(time.Millisecond * 200)
	value, err := d.bus.ReadByteFromReg(d.address, DieTempConfig)
	if err != nil {
		return 0, err
	}

	if (value & 0x01) != 0 {
		return 0, ErrReadTimeout
	}

	tempInt, err := d.bus.ReadByteFromReg(d.address, DieTempInt)
	if err != nil {
		return 0, err
	}

	tempFrac, err := d.bus.ReadByteFromReg(d.address, DieTempFrac)
	if err != nil {
		return 0, err
	}

	return float64(tempInt) + (float64(tempFrac) * 0.0625), nil
}

func (d *MAX30105Driver) ReadSamples() ([]Sample, error) {
	readPointer, err := d.GetReadPointer()
	if err != nil {
		return nil, err
	}

	writePointer, err := d.GetWritePointer()
	if err != nil {
		return nil, err
	}

	if readPointer == writePointer {
		return []Sample{}, nil
	}

	numSamples := writePointer - readPointer
	if writePointer < readPointer {
		numSamples += 32
	}

	d.bus.WriteByte(d.address, FIFOData)

	data, err := d.bus.ReadBytes(d.address, int(numSamples)*d.activeLEDs*3)
	if err != nil {
		return nil, err
	}

	rd := bytes.NewReader(data)

	var samples []Sample

	for rd.Len() > 0 {
		var red, ir, green int

		red, err = readNumber(rd)
		if err != nil {
			return nil, err
		}

		if d.activeLEDs > 1 {
			ir, err = readNumber(rd)
			if err != nil {
				return nil, err
			}
		}

		if d.activeLEDs > 2 {
			green, err = readNumber(rd)
			if err != nil {
				return nil, err
			}
		}

		samples = append(samples, Sample{
			Red:   red,
			IR:    ir,
			Green: green,
		})
	}

	return samples, nil
}

func readNumber(rd *bytes.Reader) (int, error) {
	numberBytes := make([]byte, 3)
	_, err := rd.Read(numberBytes)
	if err != nil {
		return 0, err
	}

	numberRd := bytes.NewReader([]byte{0x00, numberBytes[0],
		numberBytes[1], numberBytes[2]})

	var number uint32
	err = binary.Read(numberRd, binary.BigEndian, &number)
	if err != nil {
		return 0, err
	}

	number &= 0x3FFFF

	return int(number), nil
}
