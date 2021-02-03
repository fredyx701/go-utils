package utils

// Fibonacci  start n = 0
func Fibonacci(n int) int {
	if n >= 2 {
		return Fibonacci(n-1) + Fibonacci(n-2)
	}
	if n == 1 || n == 0 {
		return 1
	}
	return 0
}
