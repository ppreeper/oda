package main

import (
	"archive/zip"
	"io"
	"os"
)

func zipDir() {
	zipArchive, _ := os.Create("myproject.zip")
	defer zipArchive.Close()
	writer := zip.NewWriter(zipArchive)

	f1, _ := os.Open("pom.xml")
	defer f1.Close()
	w1, _ := writer.Create("pom.xml")
	io.Copy(w1, f1)

	f2, _ := os.Open("SampleApplication.java")
	defer f2.Close()
	w2, _ := writer.Create("src/main/java/SampleApplication.java")
	io.Copy(w2, f2)

	writer.Close()
}
