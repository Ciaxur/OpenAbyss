package utils

import (
	"log"
	"os"
	"regexp"
)

func IsErrorSIGINT(err error) bool {
	if err != nil {
		return err.Error() == "^C"
	}
	return false
}

func HandleErr(err error, msg string) bool {
	if err != nil {
		if IsErrorSIGINT(err) {
			log.Fatalln("Unhandled SIGINT")
			os.Exit(1)
		}
		errDescIdx := regexp.MustCompile(`(?:desc = )\w.*$`).FindStringIndex(err.Error())
		if len(errDescIdx) > 0 {
			log.Println(msg+":", err.Error()[errDescIdx[0]+7:errDescIdx[1]])
		} else {
			log.Println(msg+": ", err)
		}
		return false
	}
	return true
}
