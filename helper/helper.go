package helper

// CalculateBatasSKS calculates maximum SKS quota allowed for a student based on IPK
func CalculateBatasSKS(ipk float64) int {
	if ipk >= 3.00 {
		return 24
	}
	if ipk >= 2.50 {
		return 21
	}
	return 18
}
