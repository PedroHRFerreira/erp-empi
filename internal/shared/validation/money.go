package validation

func CalculateMarkupCents(costCents int64, markupPercent float64) int64 {
	if costCents <= 0 {
		return 0
	}
	return costCents + int64(float64(costCents)*(markupPercent/100))
}
