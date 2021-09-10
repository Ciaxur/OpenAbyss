package utils

import "os"

/**
 * Helper function that checks if a file exists
 *  given a file name returning the state of
 *  the existance of the file
 */
func FileExists(filename string) bool {
	if fstat, err := os.Stat(filename); err == nil && !fstat.IsDir() {
		return true
	}
	return false
}

/**
 * Helper function that checks if a dir exists
 *  given a path returning the state of
 *  the existance of the dir
 */
func DirExists(pathname string) bool {
	if fstat, err := os.Stat(pathname); err == nil && fstat.IsDir() {
		return true
	}
	return false
}

/**
 * Helper function that checks if a given path exists
 *  given a path name returning the state of
 *  the existance of the path
 */
func PathExists(pathname string) bool {
	if _, err := os.Stat(pathname); err == nil {
		return true
	}
	return false
}
