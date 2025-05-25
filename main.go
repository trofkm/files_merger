package main

import (
	"bytes"
	"flag"
	"fmt"
	"golang.org/x/sync/errgroup"
	"os"
	"path/filepath"
	"strings"
)

var (
	extensions    = flag.String("extensions", "go", "file extensions to be parsed comma separated")
	output        = flag.String("output", "output.txt", "output file name")
	ignore        = flag.String("ignore", ".git,.idea", "ignore dirs with these names comma separated")
	commentSymbol = flag.String("comment", "//", "comment symbol which used to write file name")

	Usage = func() {
		fmt.Printf("%s - utility to merge files with their names and contents\n", os.Args[0])
		fmt.Printf("Usage: %s [flags] path\n", os.Args[0])
		flag.PrintDefaults()
	}
)

// cache to prevent scanning the same file (it is possible if user pass multiple dirs and their paths somehow intersect)
var pathCache = make(map[string]bool)

func main() {
	flag.Usage = Usage
	flag.Parse()

	paths := flag.Args()
	if len(paths) == 0 {
		fmt.Printf("No paths provided\n")
		Usage()
		os.Exit(1)
	}

	exts, err := prepareExtensions(*extensions)
	if err != nil {
		return
	}

	ignDirs, err := prepareIgnoredDirs(*ignore)
	if err != nil {
		return
	}

	file, err := os.OpenFile(*output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Could not create output file: %s", err)
		os.Exit(1)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	outputCh := make(chan []byte, 4)

	receiveAndWrite := func() error {
		for data := range outputCh {
			_, err2 := file.Write(data)
			if err2 != nil {
				return err2
			}
		}
		return nil
	}

	wg := errgroup.Group{}
	wg.Go(receiveAndWrite)

	for _, path := range paths {
		if err := parse(path, exts, ignDirs, *commentSymbol, outputCh); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}

	close(outputCh)

	if err := wg.Wait(); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}

// prepareExtensions trim spaces and add '.' to the extensions
func prepareExtensions(extensions string) ([]string, error) {
	extSplitted := strings.Split(extensions, ",")
	if len(extSplitted) == 0 {
		return nil, fmt.Errorf("extensions must be a non-empty comma-separated string")
	}

	// remove duplicates
	extSplitted = removeDuplicates(extSplitted)

	for i, e := range extSplitted {
		extSplitted[i] = fmt.Sprintf(".%s", strings.TrimSpace(e))
	}
	return extSplitted, nil
}

// prepareIgnoredDirs trim spaces from the dir list and split it by comma
func prepareIgnoredDirs(ignoredDirs string) ([]string, error) {
	ignoredDirsSplitted := strings.Split(ignoredDirs, ",")
	if len(ignoredDirsSplitted) == 0 {
		return nil, fmt.Errorf("ignoredDirs must be a non-empty comma-separated string")
	}

	// remove duplicates
	ignoredDirsSplitted = removeDuplicates(ignoredDirsSplitted)

	for i, d := range ignoredDirsSplitted {
		ignoredDirsSplitted[i] = strings.TrimSpace(d)
	}
	return ignoredDirsSplitted, nil
}

// checkExt checks if the extension is in the allowed list
func checkExt(name string, allowedExts []string) bool {
	ext := filepath.Ext(name)
	for _, e := range allowedExts {
		if ext == e {
			return true
		}
	}
	return false
}
func isIgnoredDir(path string, dirEntry os.DirEntry, ignoredDirs []string) bool {
	if !dirEntry.IsDir() {
		return false
	}

	for _, d := range ignoredDirs {
		if dirEntry.Name() == d {
			return true
		}
	}
	return false
}

// parse all files in the path with the given extensions and write their contents to the output channel
func parse(path string, extensions []string, ignoredDirs []string, commentSymbol string, output chan<- []byte) error {
	buf := bytes.Buffer{}

	err := filepath.WalkDir(path, func(path string, entry os.DirEntry, err error) error {

		// check if entry is not in ignored dir
		if isIgnoredDir(path, entry, ignoredDirs) {
			return filepath.SkipDir
		}

		if err != nil {
			return err
		}
		if !entry.IsDir() && checkExt(entry.Name(), extensions) {
			if pathCache[path] {
				return nil
			}
			// file name
			_, err := fmt.Fprintf(&buf, "%s %s\n", commentSymbol, path)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			// the actual file content
			_, err = fmt.Fprintf(&buf, "%s\n", string(data))
			if err != nil {
				return err
			}
			// TODO: I tried to do this in a separate goroutine without any additional copy,
			//  but looks like it's not working, so bytes.Buffer looks useless here
			btsCopy := make([]byte, len(buf.Bytes()))
			copy(btsCopy, buf.Bytes())
			output <- btsCopy
			buf.Reset()

			pathCache[path] = true
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func removeDuplicates[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	var list []T
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
