package main

import (
	"fmt"
	"os"
    "path/filepath"
    "strconv"
    "syscall"
)

func fileGetAbsolutePath(path string) (string, os.FileInfo) {
    ret, err := filepath.Abs(path)
    if err != nil {
        LoggerError.Fatalf("Invalid file: %v", err)
    }

    f, err := os.Lstat(ret)
    if err != nil {
        LoggerError.Fatalf("File stats failed: %v", err)
    }

    return ret, f
}

func checkIfFileExists(path string) bool {
    f, err := os.Stat(path);

    // check if path exists
    if err != nil {
        return false
    }

    return ( ! f.IsDir() ) && f.Mode().IsRegular()
}

func checkIfDirectoryExists(path string) bool {
    f, err := os.Stat(path);

    // check if path exists
    if err != nil {
        return false
    }

    return f.IsDir()
}

func checkIfFileExistsAndOwnedByRoot(path string) bool {
    f, err := os.Stat(path);

    // check if path exists
    if err != nil {
        return false
    }

    // check if it is not a file
    if ! f.Mode().IsRegular() {
        return false
    }

	uidS := fmt.Sprint(f.Sys().(*syscall.Stat_t).Uid)
	uid, err := strconv.Atoi(uidS)
	if err != nil {
		return false
	}

    if uid != 0 {
        return false
    }

    return true
}

func checkIfFileIsValid(f os.FileInfo, path string) bool {
    if f.IsDir() {
        return false
    }

    if f.Mode().IsRegular() {
        if f.Mode().Perm() & 0022 == 0 {
            return true
        } else {
            LoggerInfo.Printf("Ignoring file with wrong modes (not xx22) %s\n", path)
        }
    } else {
        LoggerInfo.Printf("Ignoring non regular file %s\n", path)
    }

    return false
}
