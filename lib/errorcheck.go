package lib

import (
	"fmt"
	"os"
)

func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func CheckErr(msg interface{}) {
	if msg != nil {
		fmt.Fprintln(os.Stderr, "Error:", msg)
		os.Exit(1)
	}
}
