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
type sniffer struct {
	target  string
	replace string

	output chan string
	mutex  sync.Mutex

	wg sync.WaitGroup
}

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

// General purpose message sending syncronization
func atomicMessage[T any](c chan T, msg T, m *sync.Mutex) {
	m.Lock()
	defer m.Unlock()
	c <- msg
}

func (i *sniffer) sniff(root string, index int) {

	thisDir, _ := os.ReadDir(root)

	folders := filter(thisDir, func(de fs.DirEntry) bool {
		return de.IsDir() && hasChildren(filepath.Join(root, de.Name()))
	})

	for _, folder := range folders {
		i.wg.Add(1)
		// i.waiters++
		go i.sniff(filepath.Join(root, folder.Name()), index+1)
	}

	for _, entry := range thisDir {
		if strings.Contains(entry.Name(), i.target) {
			atomicMessage(i.output, filepath.Join(root, entry.Name()), &i.mutex)
		}
	}

	i.wg.Done()

	if index == 0 {
		// fmt.Println("Waiting...")
		i.wg.Wait()
		close(i.output)
	}

}

func (i *sniffer) modify() {

	var acc []string

	for k := range i.output {
		// fmt.Println("Found: ", k)
		acc = append(acc, k)
	}

	var count int = 0

	for _, el := range acc {

		err := os.Rename(el, strings.ReplaceAll(el, i.target, i.replace))
		for err != nil {
			err = os.Rename(strings.Replace(el, i.target, i.replace, 1), strings.ReplaceAll(el, i.target, i.replace))

			if err != nil && strings.ContainsAny(err.Error(), "file exists") {
				err = nil
			}
		}
		count++

	}

	fmt.Printf("Found and renamed %d/%d items\n", count, len(acc))

}

func buildSniffer() (i sniffer, root string) {

	// fmt.Println(len(os.Args), " -> ", os.Args)

	args := os.Args[1:]

	if len(args) < 3 {
		fmt.Println("Bad arguments -> rename <root-folder> <target-string> <replace-string>")
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

	i = sniffer{
		target:  args[1],
		replace: args[2],

		output: make(chan string),
	}

	return
}

func main() {

	sniffer, root := buildSniffer()

	sniffer.wg.Add(1)
	go sniffer.sniff(root, 0)
	sniffer.modify()

}
