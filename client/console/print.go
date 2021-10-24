package console

import "os"

// Fatalln Print and exit
func Fatalln(v ...interface{}) {
	Error.Println(v)
	os.Exit(1)
}

// Fatalf Print and exit
func Fatalf(format string, v ...interface{}) {
	Error.Printf(format, v)
	os.Exit(1)
}
