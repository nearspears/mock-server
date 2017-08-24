package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"bufio"
	"io"
	"io/ioutil"
	"fmt"
	"time"
	"strconv"
	"regexp"
	"os/exec"
	"path/filepath"
)

type handle struct {
	name string
}

func (this *handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	remote, err := url.Parse(getProp(this.name, "proxy_url"))
	if err != nil {
		panic(err)
	}

	delay, err := strconv.ParseInt(getProp(this.name, "proxy_delay"), 10, 64)
	if err == nil {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	var paths = readList()
	for i := range paths {
		if match(paths[i], r.URL.Path) {
			if origin := r.Header.Get("Origin"); origin != "" {
				r.Header.Set("Access-Control-Allow-Origin", origin)
				r.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
				r.Header.Set("Access-Control-Allow-Headers",
					"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			}
			if r.Method == "OPTIONS" {
				return
			}
			w.Header().Set("Content-Type", "application/json")
			cookie := getProp(this.name, "cookie")
			if cookie != "" {
				w.Header().Add("Set-Cookie", cookie)
			}

			io.WriteString(w, readFile(this.name, paths[i]))
			return;
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ServeHTTP(w, r)
}

func match(regex string, source string) bool {
	if regex == "" {
		return false
	}
	re, err := regexp.Compile("\\{(.+?)\\}")
	if err != nil {
		panic(err)
	}
	pattern := re.ReplaceAllString(regex, "(.*)")

	re, err = regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}
	return re.MatchString(source)
}

func readList() []string {
	f, err := os.Open(getCurrentDir() + "/filter.properties")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	m := make([]string, 10)
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
		m = append(m, strings.TrimSpace(strings.Replace(line, "\n", "", -1)))
		if err != nil || io.EOF == err {
			break
		}
	}
	return m
}

func readFile(name string, filePath string) string {
	f, err := os.Open(getCurrentDir() + "/" + getProp(name, "response_dir") + filePath + ".json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fd, err := ioutil.ReadAll(f)
	fmt.Println(string(fd))//TODO
	return string(fd)
}

func readProperties() map[string]string {
	f, err := os.Open(getCurrentDir() + "/config.properties")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	m := make(map[string]string)
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
		if strings.Contains(line, "=") {
			entry := strings.Split(line, "=")
			m[entry[0]] = strings.TrimSpace(strings.Replace(entry[1], "\n", "", -1))
		}
		if err != nil || io.EOF == err {
			break
		}
	}
	return m
}

func getProp(name string, key string) string {
	if name == "default" {
		return properties[key]
	} else {
		return properties[key + "." + name]
	}
}

func getCurrentDir() string {
	file, _ := exec.LookPath(os.Args[0])
	path := filepath.Dir(file)
	return path
}

func listenPort(name string, port string) {
	h := &handle{name: name}
	err := http.ListenAndServe(":" + port, h)
	if err != nil {
		log.Fatalln("ListenAndServe " + port + ":", err)
		intValue, _ :=  strconv.Atoi(port)
		channel <- intValue
	}
}

var properties map[string]string
var channel chan int = make(chan int)

func startServer() {
	//被代理的服务器host和port

	properties = readProperties()
	for k, v := range properties {
		if strings.Contains(k, "port") {
			name := "default"
			if strings.Contains(k, ".") {
				name = strings.Split(k, ".")[1]
			}
			go listenPort(name, v)
		}
	}

	for k, _ := range properties {
		if strings.Contains(k, "port") {
			fmt.Printf("server %s quit", <-channel)
		}
	}
}

func main() {
	startServer()
}