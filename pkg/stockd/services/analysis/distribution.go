package analysis

import (
	"stock/pkg/models"
	"stock/pkg/utils"
)

func Distribution(values []float64, numBins int) []*models.DistBucket {
	if len(values) == 0 {
		return nil
	}
	vmin, vmax := values[0], values[0]
	for _, v := range values {
		if v < vmin {
			vmin = v
		}
		if v > vmax {
			vmax = v
		}
	}
	if vmin == vmax {
		return []*models.DistBucket{{Lower: vmin, Upper: vmax, Count: len(values), Pct: 100.0}}
	}
	width := (vmax - vmin) / float64(numBins)
	out := make([]*models.DistBucket, 0, numBins)
	for i := 0; i < numBins; i++ {
		low := vmin + float64(i)*width
		high := vmin + float64(i+1)*width
		count := 0
		if i == numBins-1 {
			for _, v := range values {
				if low <= v && v <= high {
					count++
				}
			}
		} else {
			for _, v := range values {
				if low <= v && v < high {
					count++
				}
			}
		}
		pct := utils.RoundTo(float64(count)/float64(len(values))*100.0, 1)
		out = append(out, &models.DistBucket{Lower: low, Upper: high, Count: count, Pct: pct})
	}
	return out
}
