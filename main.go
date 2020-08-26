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

var homedir string = getHomeDir()

func getHomeDir() string {
	homedir, err := os.UserHomeDir()
	check(err)
	return homedir
}

// GHKey is one ssh key info returned from the github API.
type GHKey struct {
	ID  int    `json:"id"`
	Key string `json:"key"`
}

func writeKeys(keyList []string) {
	filename := homedir + "/.ssh/authorized_keys"

	fmt.Println("Writing merged key list to " + filename)

	file, err := os.Create(filename)
	check(err)

	defer file.Close()

	for _, key := range keyList {
		file.WriteString(key + "\n")
	}

	fmt.Println("Written successfully. Enjoy your day!")
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
	fmt.Println("Checking local keyfile.")
	filename := homedir + "/.ssh/authorized_keys"

	if !fileExists(filename) {
		fmt.Println("Authorized key file does not exist at " + filename + ", will be created.")
		return []string{}
	}

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

	fmt.Println("Local keys retrieved.")
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

	if username == "" {
		log.Fatal("GitHub username is required, e.g., ./gkg -username=MyUserName")
	}

	ghKeyList := fetchKeys(username)
	localKeyList := readAuthKeyFile()
	mergedKeyList := deDupe(append(ghKeyList, localKeyList...))

	writeKeys(mergedKeyList)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
