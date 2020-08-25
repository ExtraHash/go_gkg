package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// GHKey is one ssh key info returned from the github API.
type GHKey struct {
	ID  int    `json:"id"`
	Key string `json:"key"`
}

func writeKeys(keyList []string) {
	fileName := "authorized_keys"
	file, err := os.Create(fileName)
	check(err)

	defer file.Close()

	for _, key := range keyList {
		file.WriteString(key + "\n")
	}
}

func fetchKeys(username string) []string {
	fmt.Println("Fetching keys for " + username)
	resp, err := http.Get("https://api.github.com/users/" + username + "/keys")
	check(err)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	check(err)

	keyRes := []GHKey{}
	json.Unmarshal(body, &keyRes)
	fmt.Println("Keys fetched successfully.")

	keyList := []string{}
	for _, key := range keyRes {
		keyList = append(keyList, key.Key)
	}

	return keyList
}

func readAuthKeyFile() []string {
	homedir, err := os.UserHomeDir()
	check(err)

	file, err := os.Open(homedir + "/.ssh/authorized_keys")
	check(err)

	defer file.Close()

	keyList := []string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		keyList = append(keyList, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return keyList
}

func deDupe(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func parseArgs() string {
	username := flag.String("username", "", "GitHub username for which to retrieve keys.")
	flag.Parse()
	return *username
}

func main() {
	username := parseArgs()

	ghKeyList := fetchKeys(username)
	fmt.Println(ghKeyList)
	localKeyList := readAuthKeyFile()
	fmt.Println(localKeyList)
	mergedKeyList := deDupe(append(ghKeyList, localKeyList...))
	fmt.Println(localKeyList)

	writeKeys(mergedKeyList)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
