package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	config "github.com/Unknwon/goconfig"
	termbox "github.com/nsf/termbox-go"
)

const appName string = "mcmober"
const localpath string = ".minecraft/mods"

// 构建目录URL
func makeDirURL(host string) (url string) {
	url = "http://" + host + "/mods"
	return
}

// 构建文件URL
func makeFileURL(host, file string) (url string) {
	url = makeDirURL(host)
	url += "/" + file
	return
}

// 截取从A到B之间的子串, 如果不存在A或者B, 返回 ""
func between(str, starting, ending string) string {
	s := strings.Index(str, starting)
	if s < 0 {
		return ""
	}
	s += len(starting)
	e := strings.Index(str[s:], ending)
	if e < 0 {
		return ""
	}
	return str[s : s+e]
}

// 获取包名, 不包括版本号
func getPkgName(str string) string {
	idx := strings.Index(str, "-")
	if idx < 0 {
		return str
	}
	return str[:idx]
}

// 获取标准名, 去除[]的前缀
func getStdName(str string) string {
	idx := strings.Index(str, "]")
	if idx < 0 {
		return str
	}
	return str[idx+1:]
}

// 从服务器上获取文件列表
func getList(host string) (pkgs []string, err error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println()
			fmt.Println("Error Occurred,", err)
		}
	}()

	url := makeDirURL(host)

	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var str string = string(data)
	// fmt.Println(str)

	strlist := strings.Split(str, "</tr>")
	strlist = strlist[1 : len(strlist)-1]
	for _, row := range strlist {
		tds := strings.Split(row, "</td>")
		if len(tds) < 3 {
			continue
		}
		name := between(tds[0], "<tt>", "</tt>")
		pkgs = append(pkgs, name)
	}

	return
}

// 根据服务器文件列表检查本地列表, 只保留不存在或者不一致的部分
func checkList(pkgs []string) (npkgs []string, err error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println()
			fmt.Println("Error Occurred,", err)
		}
	}()

	err = os.MkdirAll(localpath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	dir, err := ioutil.ReadDir(localpath)
	if err != nil {
		panic(err)
	}
	mmap := make(map[string]string)
	for _, file := range dir {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		stdName := getStdName(fileName)
		if stdName == "" {
			continue
		}
		if stdName != fileName {
			opath := localpath + "/" + fileName
			npath := localpath + "/" + stdName
			e := os.Rename(opath, npath)
			if e != nil {
				continue
			}
		}
		pkgName := getPkgName(stdName)
		mmap[pkgName] = stdName
	}

	for _, pkg := range pkgs {
		pkgName := getPkgName(pkg)
		fileName, isExist := mmap[pkgName]
		if !isExist || fileName != pkg {
			npkgs = append(npkgs, pkg)
			fmt.Println(pkg)
		}
	}

	return
}

// 下载服务器文件
func downloadFile(url string) (err error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println()
			fmt.Println("Error Occurred,", err)
		}
	}()

	idx := strings.LastIndex(url, "/")
	fileName := url
	if idx != -1 {
		fileName = fileName[idx+1 : len(fileName)]
	}
	fmt.Println("Start Download File", fileName)
	fileName = localpath + "/" + fileName

	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	fsize, err := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 32)
	if err != nil {
		fmt.Println(err)
	}

	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	buf := make([]byte, 4096)
	var csize int64 = 0
	for {
		n, err := res.Body.Read(buf)
		if err != nil {
			if err != io.EOF {
				os.Remove(fileName)
				panic(err)
			} else {
				f.Write(buf[:n])
				break
			}
		}
		csize += int64(n)
		f.Write(buf[:n])
		fmt.Printf("%s\t\t%d%%, %d/%d\r", fileName, 100*csize/fsize, csize, fsize)
	}
	fmt.Printf("%s\t\t%d%%, %d/%d\n", fileName, 100, fsize, fsize)
	fmt.Println("Download Finish", fileName)

	return
}

// 暂停直到任意键被按下
func pause(str string) {
	fmt.Println(str)
Loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			break Loop
		}
	}
}

func main() {
	termbox.Init()
	defer termbox.Close()

	confName := appName + ".ini"
	cfg, _ := config.LoadConfigFile(confName)
	host, _ := cfg.GetValue("General", "host")
	if host == "" {
		fmt.Println("无法找到服务器")
		fmt.Println()
		pause("请按任意键继续...")
		return
	}

	// const host string = "134.175.247.23:8080"
	fmt.Println("正在获取mods列表")
	pkgs, _ := getList(host)
	pkgs, _ = checkList(pkgs)
	fmt.Println()

	fmt.Println("开始同步mods")
	fmt.Println()
	for _, pkg := range pkgs {
		url := makeFileURL(host, pkg)
		downloadFile(url)
		fmt.Println()
	}

	fmt.Println("同步完成")
	fmt.Println()

	pause("请按任意键继续...")
}
