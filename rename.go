package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

/**
*	Simple struct to help our goal
*
 */
type searcher struct {
	target  string
	replace string

	output chan string

	wg sync.WaitGroup
}

const BUFFER_SIZE int = 64

/**
*	Basic helper function, checks if a give folder/file has children
 */
func hasChildren(target string) bool {
	res, _ := os.ReadDir(target)
	return len(res) > 0
}

/**
*	General purpose function, filters an array,
*		returning the elements selected by func f
*
 */
func filter[T any](origin []T, f func(T) bool) (ret []T) {
	for _, el := range origin {
		if f(el) {
			ret = append(ret, el)
		}
	}
	return
}

func (s *searcher) search(root string, index int) {

	thisDir, _ := os.ReadDir(root)

	folders := filter(thisDir, func(de fs.DirEntry) bool {
		return de.IsDir() && hasChildren(filepath.Join(root, de.Name()))
	})

	for _, folder := range folders {
		s.wg.Add(1)
		go s.search(filepath.Join(root, folder.Name()), index+1)
	}

	if index == 0 && strings.Contains(root, s.target) {
		s.output <- root
	}

	for _, entry := range thisDir {

		if strings.Contains(entry.Name(), s.target) {
			s.output <- filepath.Join(root, entry.Name())
		}
	}

	s.wg.Done()

	if index == 0 {

		// fmt.Println("Waiting...")
		s.wg.Wait()
		close(s.output)
	}

}

func (s *searcher) modify() {

	var acc []string

	for k := range s.output {
		acc = append(acc, k)
	}

	// var count int = 0

	for _, el := range acc {
		currPath := el
		finalPath := strings.ReplaceAll(el, s.target, s.replace)
		err := os.Rename(currPath, finalPath)
		for err != nil {
			currPath = strings.Replace(currPath, s.target, s.replace, 1)
			err = os.Rename(currPath, finalPath)
			// fmt.Println(err)

			if err != nil && strings.Contains(err.Error(), "file exists") {
				err = nil
			}
		}

	}

	fmt.Printf("Done\n")

}

func buildSearcher() (s searcher, root string) {

	// fmt.Println(len(os.Args), " -> ", os.Args)

	args := os.Args[1:]

	if len(args) < 3 {
		fmt.Println("Bad arguments, please see: rename <root-folder> <target-string> <replace-string>")
		os.Exit(-1)
	}

	if len(args) > 3 {
		fmt.Printf("Too many arguments, ignoring: ")
		for _, arg := range args[3:] {
			fmt.Printf("<%s> ", arg)
		}
		fmt.Println()
	}

	root = args[0]

	s = searcher{
		target:  args[1],
		replace: args[2],

		output: make(chan string),
	}

	return
}

func main() {

	searcher, root := buildSearcher()

	searcher.wg.Add(1)
	go searcher.search(root, 0)
	searcher.modify()

	// test()

}
