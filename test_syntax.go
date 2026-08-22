package main

import "fmt"

type SystemService struct {
	isOpen bool
}

// IsOpen returns whether the system service is open
func (s *SystemService) IsOpen() bool {
	return s.isOpen
}

func main() {
	fmt.Println("Syntax is valid")
}
