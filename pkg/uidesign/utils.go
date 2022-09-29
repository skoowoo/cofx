package uidesign

func max(nums ...int) int {
	var max int
	for _, n := range nums {
		if n > max {
			max = n
		}
	}
	return max
}
