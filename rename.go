package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Inspector struct {
	target  string
	replace string

	output chan string
	mutex  sync.Mutex

	wg sync.WaitGroup
}

func hasChildren(target string) bool {

	res, _ := os.ReadDir(target)

	return len(res) > 0
}

func filter[T any](origin []T, f func(T) bool) (ret []T) {
	for _, el := range origin {
		if f(el) {
			ret = append(ret, el)
		}
	}
	return
}

func atomicMessage[T any](c chan T, msg T, m *sync.Mutex) {
	m.Lock()
	defer m.Unlock()
	c <- msg
}

func (i *Inspector) inspect(root string, index int) {

	thisDir, _ := os.ReadDir(root)

	folders := filter(thisDir, func(de fs.DirEntry) bool {
		return de.IsDir() && hasChildren(filepath.Join(root, de.Name()))
	})

	for _, folder := range folders {
		i.wg.Add(1)
		// i.waiters++
		go i.inspect(filepath.Join(root, folder.Name()), index+1)
	}

	for _, entry := range thisDir {
		if strings.Contains(entry.Name(), i.target) {
			// fmt.Println("Aquiring lock...")
			atomicMessage(i.output, filepath.Join(root, entry.Name()), &i.mutex)
			// fmt.Println("Released Lock!")
		}
	}

	i.wg.Done()

	if index == 0 {
		i.wg.Wait()
		close(i.output)
	}

}

func (i *Inspector) consumer() {

	var acc []string

	for k := range i.output {
		fmt.Println("Found: ", k)
		acc = append(acc, k)
	}

	// fmt.Println("==== Done collecting ====")

	for _, el := range acc {

		err := os.Rename(el, strings.ReplaceAll(el, i.target, i.replace))
		for err != nil {
			err = os.Rename(strings.Replace(el, i.target, i.replace, 1), strings.ReplaceAll(el, i.target, i.replace))
		}

	}

}

func sanitize() (args []string) {

	args = os.Args[1:]

	return
}

func main() {

	args := os.Args[1:]

	fmt.Println("root: ", args[0], "\ntrgt: ", args[1])

	root := args[0]

	inspector := Inspector{
		target:  args[1],
		replace: args[2],

		output: make(chan string),
	}

	inspector.wg.Add(1)
	go inspector.inspect(root, 0)

	inspector.consumer()

}
