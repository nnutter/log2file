package main

import (
	"bufio"
	"fmt"
	"gopkg.in/fsnotify.v1"
	"log"
	"os"
)

type watchedFile struct {
	Name    string
	Errors  chan error
	Events  chan fsnotify.Event
	handle  *os.File
	watcher *fsnotify.Watcher
}

// log2file reads from standard input and write to the file, reopening its file
// handle if the file is renamed or deleted.
func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: log2file <filename>")
		os.Exit(1)
	}

	logFileName := os.Args[1]

	writer := make(chan string)

	wf, err := NewWatchedFile(logFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer wf.Close()

	go func() {
		for {
			select {
			case ev := <-wf.Events:
				isRemove := (ev.Op&fsnotify.Remove == fsnotify.Remove)
				isRename := (ev.Op&fsnotify.Rename == fsnotify.Rename)
				if isRemove || isRename {
					wf, err = wf.Reopen()
					if err != nil {
						log.Fatal(err)
					}
				}
			case err := <-wf.Errors:
				log.Fatal(err)
			case line := <-writer:
				err = wf.Println(line)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		writer <- line
	}
}

func NewWatchedFile(name string) (*watchedFile, error) {
	mode := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	perm := os.FileMode(0660)

	handle, err := os.OpenFile(name, mode, perm)
	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		handle.Close()
		return nil, err
	}

	err = watcher.Add(name)
	if err != nil {
		handle.Close()
		watcher.Close()
		return nil, err
	}

	return &watchedFile{
		Name:    name,
		Errors:  watcher.Errors,
		Events:  watcher.Events,
		handle:  handle,
		watcher: watcher,
	}, nil
}

func (wf *watchedFile) Close() error {
	hErr := wf.handle.Close()
	wErr := wf.watcher.Close()

	if hErr != nil {
		return hErr
	}

	if wErr != nil {
		return wErr
	}

	return nil
}

func (wf *watchedFile) Reopen() (*watchedFile, error) {
	err := wf.Close()
	if err != nil {
		return nil, err
	}

	return NewWatchedFile(wf.Name)
}

func (wf *watchedFile) Println(line string) error {
	_, err := fmt.Fprintln(wf.handle, line)
	if err != nil {
		return err
	}
	return nil
}
