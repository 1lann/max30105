package max30105

import (
	"errors"
	"time"

	"github.com/kidoman/embd"
)

// Sample represents a sample reading from the sensor.
type Sample struct {
	Red   int
	IR    int
	Green int
}

// MAX30105Driver represents the initialised driver for a MAX30105 chip.
type MAX30105Driver struct {
	bus        embd.I2CBus
	address    byte
	activeLEDs int
}

// All the constants for all the registers
const (
	MAX30105Address = 0x57

	INT1Register = 0x00
	INT2Register = 0x01
	INTEnable1   = 0x02
	INTEnable2   = 0x03

	FIFOWriterPtr = 0x04
	FIFOOverflow  = 0x05
	FIFOReadPtr   = 0x06
	FIFOData      = 0x07

	FIFOConfig      = 0x08
	ModeConfig      = 0x09
	ParticleConfig  = 0x0A
	LED1PulseAmp    = 0x0C
	LED2PulseAmp    = 0x0D
	LED3PulseAmp    = 0x0E
	LEDProxAmp      = 0x10
	MultiLEDConfig1 = 0x11
	MultiLEDConfig2 = 0x12

	DieTempInt    = 0x1F
	DieTempFrac   = 0x20
	DieTempConfig = 0x21

	ProxIntThresh = 0x30

	RevisionID = 0xFE
	PartID     = 0xFF

	IntAFullMask    = 1 << 7
	IntAFullEnable  = 0x80
	IntAFullDisable = 0x00

	IntDataReadyMask    = 1 << 6
	IntDataReadyEnable  = 0x40
	IntDataReadyDisable = 0x00

	IntAlcOvfMask    = 1 << 5
	IntAlcOvfEnable  = 0x20
	IntAlcOvfDisable = 0x00

	IntProxIntMask    = 1 << 4
	IntProxIntEnable  = 0x10
	IntProxIntDisable = 0x00

	IntDieTempReadyMask    = 1 << 1
	IntDieTempReadyEnable  = 0x02
	IntDieTempReadyDisable = 0x00

	SampleAverageMask = 0xE0
	SampleAverage1    = 0x00
	SampleAverage2    = 0x20
	SampleAverage4    = 0x40
	SampleAverage8    = 0x60
	SampleAverage16   = 0x80
	SampleAverage32   = 0xA0

	RolloverMask    = 0xEF
	RolloverEnable  = 0x10
	RolloverDisable = 0x00

	AFullMask    = 0xF0
	ShutdownMask = 0x7F
	Shutdown     = 0x80
	Wakeup       = 0x00

	ResetMask = 0xBF
	Reset     = 0x40

	ModeMask      = 0xF8
	ModeRedOnly   = 0x02
	ModeRedIROnly = 0x03
	ModeMultiLED  = 0x07

	ADCRangeMask  = 0x9F
	ADCRange2048  = 0x00
	ADCRange4096  = 0x20
	ADCRange8192  = 0x40
	ADCRange16384 = 0x60

	SampleRateMask = 0xE3
	SampleRate50   = 0x00
	SampleRate100  = 0x04
	SampleRate200  = 0x08
	SampleRate400  = 0x0C
	SampleRate800  = 0x10
	SampleRate1000 = 0x14
	SampleRate1600 = 0x18
	SampleRate3200 = 0x1C

	PulseWidthMask = 0xFC
	PulseWidth69   = 0x00
	PulseWidth118  = 0x01
	PulseWidth215  = 0x02
	PulseWidth411  = 0x03

	Slot1Mask = 0xF8
	Slot2Mask = 0x8F
	Slot3Mask = 0xF8
	Slot4Mask = 0x8F

	SlotNone       = 0x00
	SlotRedLED     = 0x01
	SlotIRLED      = 0x02
	SlotGreenLED   = 0x03
	SlotNonePilot  = 0x04
	SlotRedPilot   = 0x05
	SlotIRPilot    = 0x06
	SlotGreenPilot = 0x07

	ExpectedPartID = 0x15
)

// Possible errors that can occur
var (
	ErrIncorrectPart    = errors.New("max30105: incorrect part")
	ErrInvalidParameter = errors.New("max30105: invalid parameter")
	ErrReadTimeout      = errors.New("max30105: read timeout")
)

// NewDriver returns a new MAX30105 driver with the provided options.
func NewDriver(bus embd.I2CBus) *MAX30105Driver {
	d := &MAX30105Driver{
		bus:     bus,
		address: MAX30105Address,
	}

	return d
}

func (d *MAX30105Driver) Setup() error {
	data, err := d.bus.ReadByteFromReg(d.address, PartID)
	if err != nil {
		return err
	}

	if data != ExpectedPartID {
		return ErrIncorrectPart
	}

	var errs []error
	errs = append(errs, d.bitMask(ModeConfig, ResetMask, Reset))
	time.Sleep(time.Millisecond * 200)
	errs = append(errs, d.SetFIFOAverage(SampleAverage8))
	errs = append(errs, d.EnableFIFORollover())
	errs = append(errs, d.SetLEDMode(ModeMultiLED))
	errs = append(errs, d.SetADCRange(ADCRange4096))
	errs = append(errs, d.SetSampleRate(SampleRate400))
	errs = append(errs, d.SetPulseWidth(PulseWidth411))
	errs = append(errs, d.SetRedAmplitude(0))
	errs = append(errs, d.SetIRAmplitude(0x1F))
	errs = append(errs, d.SetGreenAmplitude(0))
	errs = append(errs, d.SetProximityAmplitude(0x1F))
	errs = append(errs, d.EnableSlot(1, SlotRedLED))
	errs = append(errs, d.EnableSlot(2, SlotIRLED))
	errs = append(errs, d.EnableSlot(3, SlotGreenLED))
	errs = append(errs, d.ClearFIFO())

	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

// SPO2 \ln \left(-\frac{x}{2.05}+2.7\right)
