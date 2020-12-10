package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	// "io"
	"path/filepath"
)

var (
	ignoredFile   = "ignored_file.txt"
	successedFile = "successed_file.txt"
	pattern       = "%s#~^^^~#%s"
	uploadedFiles = make(map[string]string)
	key           string
	path          string
	found         bool
	limit         int
	step          int
)

func main() {

	flag.StringVar(&path, "path", path, "директория для сканирования")
	flag.IntVar(&step, "step", 10, "кол-во успешных отправок за один запуск")

	flag.Parse()

	if _, err := os.Stat(path); err == nil || os.IsExist(err) {
	} else {
		check(err)
	}

	if _, err := os.Stat(ignoredFile); err == nil || os.IsExist(err) {
		err := os.Remove(ignoredFile)
		check(err)
	}

	uploadedFiles, err := readSuccessedFromFile()
	check(err)

	out := os.Stdout

	// перед проходом по директории, восстановим мапу с файла
	err = dirTree(out, path, uploadedFiles, &limit, step)
	check(err)

}

func dirTree(out io.Writer, path string, uploadedFiles map[string]string, limit *int, step int) error {
	pathAbs, err := filepath.Abs(path)
	check(err)
	recDirTree(out, pathAbs, 0, uploadedFiles, limit, step)
	return nil
}

func recDirTree(out io.Writer, path string, lvl int, uploadedFiles map[string]string, limit *int, step int) {

	var size int64
	var pathList []string
	var pathListErr error
	var availableExt = []string{".png", ".JPG", ".jpg"}

	lvl++
	hPath, _ := os.Open(path)
	fInfo, _ := hPath.Stat()
	if fInfo.IsDir() {
		pathList, pathListErr = hPath.Readdirnames(1000)
		check(pathListErr)

		sort.Strings(pathList)
	}

	for index := range pathList {

		absolutePath := filepath.Join(path, pathList[index])
		hPathFile, _ := os.Open(absolutePath)
		fInfoFile, _ := hPathFile.Stat()

		if !fInfoFile.IsDir() {
			size = fInfoFile.Size()
			if size > 0 {

				// проверяем расширение файла
				if isContains(availableExt, filepath.Ext(pathList[index])) {

					key = pathList[index] + "," + strconv.FormatInt(size, 10)
					_, found = uploadedFiles[key]
					if !found {
						// пользовательская логика
						// после успеха запись в мапу и сохранение в файл
						if true {
							fmt.Println(absolutePath)
							uploadedFiles[key] = absolutePath
							saveSuccessedToFile(key, absolutePath)
							//fmt.Fprintln(out, absolutePath)
							*limit++

							if *limit >= step {
								fmt.Println("Limit reached")
								os.Exit(0)
							}
						}

					}
					//fmt.Println(uploadedFiles)

				} else {
					// фиксируем те которые не прошли
					saveIgnoredExtToFile(absolutePath)
				}
			}
		}
		//fmt.Fprintln(out, sepRow+pathList[index]+sizeStr)
		if fInfo.IsDir() {
			recDirTree(out, filepath.Join(path, pathList[index]), lvl, uploadedFiles, limit, step)
		}
	}

}

func isContains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func saveIgnoredExtToFile(ignoredPath string) {

	f, err := os.OpenFile(ignoredFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	check(err)
	defer f.Close()

	datawriter := bufio.NewWriter(f)
	_, _ = datawriter.WriteString(ignoredPath + "\n")
	datawriter.Flush()
}

func saveSuccessedToFile(key string, successedPath string) {
	f, err := os.OpenFile(successedFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	check(err)
	defer f.Close()

	row := fmt.Sprintf(pattern, key, successedPath)
	datawriter := bufio.NewWriter(f)
	_, _ = datawriter.WriteString(row + "\n")
	datawriter.Flush()
}

func readSuccessedFromFile() (map[string]string, error) {
	lines := make(map[string]string)
	var s []string

	if _, err := os.Stat(successedFile); err == nil || os.IsExist(err) {
		file, err := ioutil.ReadFile(successedFile)
		if err != nil {
			return lines, err
		}
		buf := bytes.NewBuffer(file)
		for {
			line, err := buf.ReadString('\n')
			if len(line) == 0 {
				if err != nil {
					if err == io.EOF {
						break
					}
					return lines, err
				}
			}
			s = strings.Split(line, fmt.Sprintf(pattern, "", ""))
			if len(s) != 2 {
				continue
			}

			lines[s[0]] = s[1]
			if err != nil && err != io.EOF {
				return lines, err
			}
		}
	}
	return lines, nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
