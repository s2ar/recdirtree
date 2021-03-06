package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	// "io"
	"path/filepath"

	"github.com/pkg/errors"
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

// вызов go run main.go --path=/home/s2ar/Pictures/TEST --step=2
func main() {

	flag.StringVar(&path, "path", path, "директория для сканирования")
	flag.IntVar(&step, "step", 10, "кол-во успешных отправок за один запуск")

	flag.Parse()

	if _, err := os.Stat(path); err == nil || os.IsExist(err) {
	} else {
		handleError(errors.Wrap(err, "Invalid directory path"))
	}
	if _, err := os.Stat(ignoredFile); err == nil || os.IsExist(err) {
		err := os.Remove(ignoredFile)
		handleError(errors.Wrap(err, "File not deleted"))
	}
	// перед проходом по директории, восстановим мапу с файла
	uploadedFiles, err := readSuccessedFromFile()
	handleError(errors.Wrap(err, "Var uploadedFiles not created"))

	out := os.Stdout

	pathAbs, err := filepath.Abs(path)
	handleError(errors.Wrap(err, "No absolute path representation created"))
	recDirTree(out, pathAbs, 0, uploadedFiles, &limit, step)
}

func recDirTree(out io.Writer, path string, lvl int, uploadedFiles map[string]string, limit *int, step int) {

	var size int64
	var pathList []string
	var err error
	var availableExt = []string{".png", ".JPG", ".jpg"}

	lvl++
	hPath, err := os.Open(path)
	handleError(errors.Wrapf(err, "Did not open directory %s", path))
	fInfo, err := hPath.Stat()
	handleError(errors.Wrapf(err, "Did not create FileInfo structure %s", path))
	if fInfo.IsDir() {
		pathList, err = hPath.Readdirnames(1000)
		handleError(errors.Wrapf(err, "Did not read directory %s", path))

		sort.Strings(pathList)
	}

	for index := range pathList {

		absolutePath := filepath.Join(path, pathList[index])
		hPathFile, err := os.Open(absolutePath)
		handleError(errors.Wrapf(err, "Did not open file %s", absolutePath))
		fInfoFile, err := hPathFile.Stat()
		handleError(errors.Wrapf(err, "Did not create FileInfo structure %s", absolutePath))

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
	handleError(errors.Wrapf(err, "Did not open file %s", ignoredPath))
	defer f.Close()

	datawriter := bufio.NewWriter(f)
	_, err = datawriter.WriteString(ignoredPath + "\n")
	handleError(errors.Wrapf(err, "Did not write file %s", ignoredPath))
	datawriter.Flush()
}

func saveSuccessedToFile(key string, successedPath string) {
	f, err := os.OpenFile(successedFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	handleError(errors.Wrapf(err, "Did not open file %s", successedPath))
	defer f.Close()

	row := fmt.Sprintf(pattern, key, successedPath)
	datawriter := bufio.NewWriter(f)
	_, err = datawriter.WriteString(row + "\n")
	handleError(errors.Wrapf(err, "Did not write file %s", successedPath))
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

func handleError(e error) {
	if e != nil {
		log.Fatalf("%+v", e)
	}
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
