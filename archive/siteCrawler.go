package main

// go website spider/crawler

// To cross compile for linux
// GOOS=linux GOARCH=amd64 go build -o siteCrawler.bin -ldflags "-w -s" siteCrawler.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o siteCrawler.exe -ldflags "-w -s" siteCrawler.go

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

func checkError(reason string, err error, condition string) {
	if err != nil {
		fmt.Printf("%s...\n", reason)
		fmt.Printf("%s", err)
		if condition == "exit" {
			os.Exit(0)
		} else {
			return
		}
	}
}

func prepareString(osDetected string, input string) string {
	var output string
	if osDetected == "windows" {
		output = strings.Replace(input, "\r\n", "", -1)
	} else {
		output = strings.Replace(input, "\n", "", -1)
	}
	return output
}

func requestSite(site string) {
	var client http.Client
	req, err := http.NewRequest("GET", site, nil)
	checkError("Unable to Request Site", err, "exit")
	resp, err := client.Do(req)
	checkError("No Response from Requested Site", err, "exit")
	respBody, _ := io.ReadAll(resp.Body)

	// Read the site line by line
	//lines := strings.Split(string(respBody), "\n")
	//for _, line := range lines {
	//	fmt.Print(line)
	//}

	// Save the site to a file for analysis
	f, err := os.Create("siteOutput.txt")
	checkError("Unable to create file to save request", err, "exit")
	defer f.Close()
	f.Write(respBody)
	f.Close()
	fmt.Print("Saved crawled site to siteOutput.txt\n")

}

func parseSiteLinks() []string {
	var siteList []string
	data, err := os.ReadFile("siteOutput.txt")
	checkError("Unable to Read File", err, "exit")
	lines := strings.Split(string(data), "\n")
	m := regexp.MustCompile(`(http:\/\/|https:\/\/|\/\/)(?:[A-Za-z0-9\.\/\-\?=%_&;])+`)
	var urlTXT string
	for _, line := range lines {
		//Output original line
		//fmt.Print("x" + line + "\n")
		links := len(strings.Split(line, "href"))
		if links > 1 {
			contentsLine := strings.Split(line, "href")
			for _, content := range contentsLine {
				urlTXT = m.FindString(content)
				if urlTXT != "" {
					//fmt.Println(urlTXT)
					siteList = append(siteList, urlTXT)
				}
			}
			//fmt.Println("%i \n", links)
			//fmt.Print("XXX\n")
		}
		urlTXT = m.FindString(line)
		if urlTXT != "" {
			//fmt.Println(urlTXT)
			siteList = append(siteList, urlTXT)
		}
		//fmt.Print("\n")
	}
	return siteList
}

func parseSiteHREF() []string {
	var hrefList []string
	data, err := os.ReadFile("siteOutput.txt")
	checkError("Unable to Read File", err, "exit")
	lines := strings.Split(string(data), "\n")
	m := regexp.MustCompile(`(?:<a href=|<script.+?src=)\".*\">[^\x3c]*`)
	var urlTXT string
	for _, line := range lines {
		//Output original line
		//fmt.Print("x" + line + "\n")
		links := len(strings.Split(line, "/a>"))
		if links > 1 {
			contentsLine := strings.Split(line, "/a>")
			for _, content := range contentsLine {
				urlTXT = m.FindString(content)
				if urlTXT != "" {
					//fmt.Println(urlTXT)
					hrefList = append(hrefList, urlTXT)
				}
			}
			//fmt.Println("%i \n", links)
			//fmt.Print("XXX\n")
		}
		urlTXT = m.FindString(line)
		if urlTXT != "" {
			//fmt.Println(urlTXT)
			hrefList = append(hrefList, urlTXT)
		}
		//fmt.Print("\n")
	}
	return hrefList
}

func removeDuplicates(s []string) []string {
	var newSlice []string
	for _, url := range s {
		found := false
		for _, value := range newSlice {
			if value == url {
				found = true
			}
		}
		if found == false {
			newSlice = append(newSlice, url)
		}
	}
	sort.Strings(newSlice)
	return newSlice
}

func main() {
	osDetected := runtime.GOOS
	fmt.Print("<-- Toby spider / crawler -->\n")
	fmt.Print("Website to crawl: ")
	inputReader := bufio.NewReader(os.Stdin)
	siteSelected, err := inputReader.ReadString('\n')
	checkError("Reading the website to crawl", err, "continue")
	siteSelected = prepareString(osDetected, siteSelected)
	requestSite(siteSelected)
	siteList := parseSiteLinks()
	siteList = removeDuplicates(siteList)

	////////////////////////////////////////////////////////////////////////////
	// Save the site list to a file for analysis
	f, err := os.Create("siteList.txt")
	checkError("Unable to create file to save request", err, "exit")
	defer f.Close()
	//f.Write(respBody)
	for _, url := range siteList {
		//fmt.Printf("%s\n", url)
		f.WriteString(url + "\n")
	}
	f.Close()
	fmt.Print("Saved site list to siteList.txt\n")

	hrefList := parseSiteHREF()
	hrefList = removeDuplicates(hrefList)
	//////////////////////////////////////////////////////////////////////////
	// Save the href list to a file for analysis
	f, err = os.Create("hrefList.txt")
	checkError("Unable to create file to save request", err, "exit")
	defer f.Close()
	//f.Write(respBody)
	for _, url := range hrefList {
		//fmt.Printf("%s\n", url)
		f.WriteString(url + "\n")
	}
	f.Close()
	fmt.Print("Saved href list to hrefList.txt\n")
}
