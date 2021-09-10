package utils

import (
	"log"
	"os"
)

func IsErrorSIGINT(err error) bool {
	if err != nil {
		return err.Error() == "^C"
	}
	return false
}

func HandleErr(err error, msg string) {
	if err != nil {
		if IsErrorSIGINT(err) {
			log.Fatalln("Unhandled SIGINT")
			os.Exit(1)
		}
		log.Panicln(msg, err)
	}
}
