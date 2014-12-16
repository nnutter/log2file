package main

import (
	"bufio"
	"fmt"
	"gopkg.in/fsnotify.v1"
	"log"
	"os"
)

// log2file reads from standard input and write to the file, reopening its file
// handle if the file is renamed or deleted.
func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: log2file <filename>")
		os.Exit(1)
	}

	logFileName := os.Args[1]

	openLogFile := func(logFileName string) *os.File {
		mode := os.O_CREATE | os.O_APPEND | os.O_WRONLY
		perm := os.FileMode(0660)

		// file has to exist in order to watch it
		logFile, err := os.OpenFile(logFileName, mode, perm)
		if err != nil {
			log.Fatal(err)
		}
		return logFile
	}
	logFile := openLogFile(logFileName)

	writer := make(chan string)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				isRemove := (ev.Op&fsnotify.Remove == fsnotify.Remove)
				isRename := (ev.Op&fsnotify.Rename == fsnotify.Rename)
				if isRemove || isRename {
					logFile.Close()
					logFile = openLogFile(logFileName)
				}
			case err := <-watcher.Errors:
				log.Fatal(err)
			case line := <-writer:
				_, err := fmt.Fprintln(logFile, line)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

	err = watcher.Add(logFileName)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		writer <- line
	}

	watcher.Close()
}
