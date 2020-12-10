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

	//fmt.Println(prettyPrint(uploadedFiles))

	//os.Exit(1)

	uploadedFiles, err := readSuccessedFromFile()
	if err != nil {
		check(err)
	}

	out := os.Stdout
	//fmt.Printf("%T\n", out)

	printFiles := true

	// перед проходом по директории, восстановим мапу с файла

	err = dirTree(out, path, printFiles, uploadedFiles, &limit, step)
	if err != nil {
		check(err)
	}
}

func dirTree(out io.Writer, path string, printFiles bool, uploadedFiles map[string]string, limit *int, step int) error {
	pathAbs, _ := filepath.Abs(path)
	recDirTree(out, pathAbs, printFiles, 0, "", uploadedFiles, limit, step)
	return nil
}

func recDirTree(out io.Writer, path string, printFiles bool, lvl int, sep string, uploadedFiles map[string]string, limit *int, step int) {
	//pathAbs, _ := filepath.Abs(path)

	var sepNew string // под уровень
	//var sepRow string // под итоговую строку
	//var sepCur string
	var size int64
	//var sizeStr string
	var pathList []string
	var pathListResult []string
	var pathListErr error
	var availableExt = []string{".png", ".JPG", ".jpg"}

	lvl++
	//hPath, _ := os.Open(pathAbs)
	hPath, _ := os.Open(path)
	fInfo, _ := hPath.Stat()
	if fInfo.IsDir() {
		pathList, pathListErr = hPath.Readdirnames(1000)
		if pathListErr != nil {
			check(pathListErr)
		}
		sort.Strings(pathList)

		// показать / не показать файлы
		if !printFiles {
			// скроем файлы
			for indexTmp := range pathList {
				hPathFileTmp, _ := os.Open(filepath.Join(path, pathList[indexTmp]))
				fInfoFileTmp, _ := hPathFileTmp.Stat()
				if fInfoFileTmp.IsDir() {
					pathListResult = append(pathListResult, pathList[indexTmp])
				}
			}
			pathList = pathListResult
		}
	}

	lenList := len(pathList)
	for index := range pathList {
		//sepCur = ""
		if lvl > 1 {
			//sepCur = sep
		}

		if lenList == index+1 {
			//sepRow = sepCur + "└───"
		} else {
			//sepRow = sepCur + "├───"
		}

		if lenList == index+1 {
			sepNew = sep + "\t"
		} else {
			sepNew = sep + "│\t"
		}

		absolutePath := filepath.Join(path, pathList[index])
		hPathFile, _ := os.Open(absolutePath)
		fInfoFile, _ := hPathFile.Stat()

		//sizeStr = ""
		if !fInfoFile.IsDir() {
			size = fInfoFile.Size()
			if size > 0 {
				//sizeStr = " (" + strconv.FormatInt(size, 10) + "b)"

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
			recDirTree(out, filepath.Join(path, pathList[index]), printFiles, lvl, sepNew, uploadedFiles, limit, step)
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
