// +build ignore

package max30105

import "sort"

const sampleFrequency = 25
const bufferSize = sampleFrequency * 4
const ma4Size = 4

//uch_spo2_table is approximated as  -45.060*ratioAverage* ratioAverage + 30.354 *ratioAverage + 94.845 ;
var uchSpo2Table = []int{95, 95, 95, 96, 96, 96, 97, 97, 97, 97, 97, 98, 98, 98, 98, 98, 99, 99, 99, 99,
	99, 99, 99, 99, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100,
	100, 100, 100, 100, 99, 99, 99, 99, 99, 99, 99, 99, 98, 98, 98, 98, 98, 98, 97, 97,
	97, 97, 96, 96, 96, 96, 95, 95, 95, 94, 94, 94, 93, 93, 93, 92, 92, 92, 91, 91,
	90, 90, 89, 89, 89, 88, 88, 87, 87, 86, 86, 85, 85, 84, 84, 83, 82, 82, 81, 81,
	80, 80, 79, 78, 78, 77, 76, 76, 75, 74, 74, 73, 72, 72, 71, 70, 69, 69, 68, 67,
	66, 66, 65, 64, 63, 62, 62, 61, 60, 59, 58, 57, 56, 56, 55, 54, 53, 52, 51, 50,
	49, 48, 47, 46, 45, 44, 43, 42, 41, 40, 39, 38, 37, 36, 35, 34, 33, 31, 30, 29,
	28, 27, 26, 25, 23, 22, 21, 20, 19, 17, 16, 15, 14, 12, 11, 10, 9, 7, 6, 5,
	3, 2, 1}

var anX = make([]int, 100)
var anY = make([]int, 100)

func maximHeartRateAndOxygenSaturation(punIRBuffer, punRedBuffer []int,
	pnSpo2 *int, pchSpo2Valid *bool, pnHeartRate, pchHRValid *int) {
	var unIRMean, nTh1, nNpks, nExactIRValleyLocsCount, nMiddleIdx,
		nPeakIntervalSum, nYAc, nXAc, nYDcMax, nXDcMax, nYDcMaxIdx, nXDcMaxIdx,
		nRatioAverage, nNume, nDenom int
	anIRValleyLocs := make([]int, 15)
	var anRatio []int

	// Calculates DC mean and subtract DC from IR.
	for _, v := range punIRBuffer {
		unIRMean += v
	}
	unIRMean = unIRMean / len(punIRBuffer)

	// Remove DC and invert signal so that we can use peak detector as valley
	// detector.
	for k, v := range punIRBuffer {
		anX[k] = -1 * (v - unIRMean)
	}

	// 4 point moving average.
	for k := 0; k < bufferSize-ma4Size; k++ {
		anX[k] = (anX[k] + anX[k+1] + anX[k+2] + anX[k+3]) / 4
	}

	// Calculate threshold.
	for k := 0; k < bufferSize; k++ {
		nTh1 += anX[k]
	}

	nTh1 = nTh1 / bufferSize

	if nTh1 < 30 {
		nTh1 = 30
	} else if nTh1 > 60 {
		nTh1 = 60
	}

	maximFindPeaks(anIRValleyLocs, &nNpks, anX, bufferSize, nTh1, 4, 15)
	nPeakIntervalSum = 0
	if nNpks >= 2 {
		for k := 1; k < nNpks; k++ {
			nPeakIntervalSum += anIRValleyLocs[k] - anIRValleyLocs[k-1]
		}
		// Since we flipped signal, we use peak detector as valley detector.
		nPeakIntervalSum = nPeakIntervalSum / (nNpks - 1)
		*pnHeartRate = (sampleFrequency * 60) / nPeakIntervalSum
		*pchHRValid = 1
	} else {
		// Not enough peaks to make a calculation.
		*pnHeartRate = -999
		*pchHRValid = 0
	}

	// Load raw value again for SPO2 calculation : RED(=y) and IR(=X)
	for k := 0; k < len(punIRBuffer); k++ {
		anX[k] = punIRBuffer[k]
		anY[k] = punRedBuffer[k]
	}

	// Find precise min near anIRValleyLocs.
	nExactIRValleyLocsCount = nNpks

	// Using exact_ir_valley_locs , find ir-red DC andir-red AC for SPO2
	// calibration an_ratio.
	// Finding AC/DC maximum of raw.

	for k := 0; k < nExactIRValleyLocsCount; k++ {
		// Do not use SPO2 because the valley loc is out of range.
		if anIRValleyLocs[k] > bufferSize {
			*pnSpo2 = -999
			*pchSpo2Valid = false
			return
		}
	}

	// Find max between two valley locations.
	// Use an_ratio between AC component of IR, red and DC component of
	// IR and red for SPO2.
	for k := 0; k < nExactIRValleyLocsCount-1; k++ {
		nYDcMax = -2147483647
		nXDcMax = -2147483647
		if (anIRValleyLocs[k+1] - anIRValleyLocs[k]) > 3 {
			for i := anIRValleyLocs[k]; i < anIRValleyLocs[k+1]; i++ {
				if anX[i] > nXDcMax {
					nXDcMax = anX[i]
					nXDcMaxIdx = i
				}
				if anY[i] > nYDcMax {
					nYDcMax = anY[i]
					nYDcMaxIdx = i
				}
			}

			nYAc = (anY[anIRValleyLocs[k+1]] - anY[anIRValleyLocs[k]]) * (nYDcMaxIdx - anIRValleyLocs[k])
			nYAc = anY[anIRValleyLocs[k]] + nYAc/(anIRValleyLocs[k+1]-anIRValleyLocs[k])
			nYAc = anY[nYDcMaxIdx] - nYAc
			nXAc = (anX[anIRValleyLocs[k+1]] - anX[anIRValleyLocs[k]]) * (nXDcMaxIdx - anIRValleyLocs[k])
			nXAc = anX[anIRValleyLocs[k]] + nXAc/(anIRValleyLocs[k+1]-anIRValleyLocs[k])
			nXAc = anX[nYDcMaxIdx] - nXAc
			nNume = (nYAc * nXDcMax) >> 7
			nDenom = (nXAc * nYDcMax) >> 7

			if (nDenom > 0) && (len(anRatio) < 5) && (nNume != 0) {
				anRatio = append(anRatio, (nNume*100)/nDenom)
			}
		}
	}

	sort.Ints(anRatio)

	nMiddleIdx = len(anRatio) / 2

	if nMiddleIdx > 1 {
		nRatioAverage = (anRatio[nMiddleIdx-1] + anRatio[nMiddleIdx]) / 2
	} else {
		nRatioAverage = anRatio[nMiddleIdx]
	}

	if (nRatioAverage > 2) && (nRatioAverage < 184) {
		*pnSpo2 = uchSpo2Table[nRatioAverage]
		*pchSpo2Valid = true
	} else {
		*pnSpo2 = -999
		*pchSpo2Valid = false
	}
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maximFindPeaks(pnLocs []int, nNpks *int, pnX []int, nSize int,
	nMinHeight int, nMinDistance int, nMaxNum int) {
	maximPeaksAboveMinHeight(pnLocs, nNpks, pnX, nSize, nMinHeight)
	maximRemoveClosePeaks(pnLocs, nNpks, pnX, nMinDistance)
	*nNpks = min(*nNpks, nMaxNum)
}

// Returns nNpks
func maximPeaksAboveMinHeight(pnLocs []int, nNpks *int, pnX []int, nSize int, nMinHeight int) {
	i := 1
	var nWidth int

	for i < nSize-1 {
		// Find left edge of potential peaks.
		if (pnX[i] > nMinHeight) && (pnX[i] > pnX[i-1]) {
			nWidth = 1
			// Find flat peaks.
			for i+nWidth < nSize && pnX[i] == pnX[i+nWidth] {
				nWidth++
			}
			// Find right edge of peaks.
			if pnX[i] > pnX[i+nWidth] && *nNpks < 15 {
				pnLocs[*nNpks] = i
				(*nNpks)++
				// For flat peaks, peak location is left edge
				i += nWidth + 1
			} else {
				i += nWidth
			}
		} else {
			i++
		}
	}
}

func maximRemoveClosePeaks(pnLocs []int, pnNpks *int, pnX []int,
	nMinDistance int) {
	var nOldNpks, nDist int

	sort.Slice(pnLocs, func(i int, j int) bool {
		return pnX[pnLocs[i]] > pnX[pnLocs[j]]
	})

	for i := -1; i < *pnNpks; i++ {
		nOldNpks = *pnNpks
		*pnNpks = i + 1
		for j := i + 1; j < nOldNpks; j++ {
			if i == -1 {
				nDist = pnLocs[j] + 1
			} else {
				nDist = pnLocs[j] - pnLocs[i]
			}

			if (nDist > nMinDistance) || (nDist < -nMinDistance) {
				pnLocs[*pnNpks] = pnLocs[j]
				(*pnNpks)++
			}
		}
	}

	// Resort (sic, restore?) indices to ascending order.
	sort.Ints(pnX)
}
