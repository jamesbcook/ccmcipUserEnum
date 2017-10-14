package main

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"flag"
)

//PhoneDirectory CiscoIP
type PhoneDirectory struct {
	Prompt           string
	SoftKeyItems     []SoftKeyItem    `xml:"SoftKeyItem"`
	DirectoryEntries []DirectoryEntry `xml:"DirectoryEntry"`
}

//SoftKeyItem for input
type SoftKeyItem struct {
	Name     string
	Position int
	URL      string
}

//DirectoryEntry names and telephone
type DirectoryEntry struct {
	Name      string
	Telephone string
}

func parseData(data []byte) *PhoneDirectory {
	pd := &PhoneDirectory{}
	err := xml.Unmarshal(data, pd)
	if err != nil {
		log.Fatal(err)
	}
	return pd
}

func main() {
	fileName := flag.String("f", "names.txt", "file to write discovered names to")
	url := flag.String("u", "https://localhost:8443/", "https://<url>:8443/xmldirectorylist.jsp")
	flag.Parse()
	file, err := os.Create(*fileName)
	if err != nil {
		log.Fatal(err)
	}
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(*url)
	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	pd := parseData(res)
	total := strings.SplitAfter(pd.Prompt, " ")
	fmt.Println("Number of records", total[5])
	running := true
	x := 0

	for {
		if !running {
			break
		}
		if len(pd.SoftKeyItems) == 4 {
			break
		}
		for _, user := range pd.DirectoryEntries {
			formatName := fmt.Sprintf("%s\n", user.Name)
			_, err := file.WriteString(formatName)
			if err != nil {
				log.Fatal(err)
			}
			x++
		}
		fmt.Println("Number of names ", x+1)

		for _, item := range pd.SoftKeyItems {
			if item.Name != "Next" {
				continue
			}
			if item.URL == "" {
				running = false
			}
			//fmt.Println(item.URL)
			resp, err := client.Get(item.URL)
			if err != nil {
				log.Fatal(err)
			}
			res, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			pd = parseData(res)
			break
		}
	}
	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}
}
