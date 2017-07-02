package max30105

func (d *MAX30105Driver) SetFIFOAverage(numSamples byte) error {
	return d.bitMask(FIFOConfig, SampleAverageMask, numSamples)
}

func (d *MAX30105Driver) ClearFIFO() error {
	err := d.bus.WriteByteToReg(d.address, FIFOWriterPtr, 0)
	if err != nil {
		return err
	}
	err = d.bus.WriteByteToReg(d.address, FIFOOverflow, 0)
	if err != nil {
		return err
	}
	err = d.bus.WriteByteToReg(d.address, FIFOReadPtr, 0)
	if err != nil {
		return err
	}
	return nil
}

func (d *MAX30105Driver) bitMask(reg byte, mask byte, value byte) error {
	original, err := d.bus.ReadByteFromReg(d.address, reg)
	if err != nil {
		return err
	}

	original = original & mask
	return d.bus.WriteByteToReg(d.address, reg, original|value)
}

func (d *MAX30105Driver) EnableFIFORollover() error {
	return d.bitMask(FIFOConfig, RolloverMask, RolloverEnable)
}

func (d *MAX30105Driver) DisableFIFORollover() error {
	return d.bitMask(FIFOConfig, RolloverMask, RolloverDisable)
}

func (d *MAX30105Driver) SetLEDMode(mode byte) error {
	switch mode {
	case ModeMultiLED:
		d.activeLEDs = 3
	case ModeRedIROnly:
		d.activeLEDs = 2
	case ModeRedOnly:
		d.activeLEDs = 1
	}

	return d.bitMask(ModeConfig, ModeMask, mode)
}

func (d *MAX30105Driver) SetADCRange(adcRange byte) error {
	return d.bitMask(ParticleConfig, ADCRangeMask, adcRange)
}

func (d *MAX30105Driver) SetSampleRate(sampleRate byte) error {
	return d.bitMask(ParticleConfig, SampleRateMask, sampleRate)
}

func (d *MAX30105Driver) SetPulseWidth(pulseWidth byte) error {
	return d.bitMask(ParticleConfig, PulseWidthMask, pulseWidth)
}

func (d *MAX30105Driver) SetRedAmplitude(amplitude byte) error {
	return d.bus.WriteByteToReg(d.address, LED1PulseAmp, amplitude)
}

func (d *MAX30105Driver) SetIRAmplitude(amplitude byte) error {
	return d.bus.WriteByteToReg(d.address, LED2PulseAmp, amplitude)
}

func (d *MAX30105Driver) SetGreenAmplitude(amplitude byte) error {
	return d.bus.WriteByteToReg(d.address, LED3PulseAmp, amplitude)
}

func (d *MAX30105Driver) SetProximityAmplitude(amplitude byte) error {
	return d.bus.WriteByteToReg(d.address, LEDProxAmp, amplitude)
}

func (d *MAX30105Driver) SetProximityThreshold(threshold byte) error {
	return d.bus.WriteByteToReg(d.address, ProxIntThresh, threshold)
}

func (d *MAX30105Driver) EnableSlot(slotNumber byte, device byte) error {
	switch slotNumber {
	case 1:
		return d.bitMask(MultiLEDConfig1, Slot1Mask, device)
	case 2:
		return d.bitMask(MultiLEDConfig1, Slot2Mask, device<<4)
	case 3:
		return d.bitMask(MultiLEDConfig2, Slot3Mask, device)
	case 4:
		return d.bitMask(MultiLEDConfig2, Slot4Mask, device<<4)
	default:
		return ErrInvalidParameter
	}
}
