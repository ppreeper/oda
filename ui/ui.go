package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

func AreYouSure(prompt string) bool {
	var confirm1, confirm2 bool
	huh.NewConfirm().
		Title(fmt.Sprintf("Are you sure you want to %s?", prompt)).
		Value(&confirm1).
		Run()
	if !confirm1 {
		return false
	}
	huh.NewConfirm().
		Title(fmt.Sprintf("Are you really sure you want to %s?", prompt)).
		Value(&confirm2).
		Run()
	if !confirm1 || !confirm2 {
		return false
	}
	return true
}
