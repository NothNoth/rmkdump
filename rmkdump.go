package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

const rmkdumpVersion = "1.0"
const remarkableURL = "http://10.11.99.1/documents/"
const remarkableDownloadURL = "http://10.11.99.1/download/"
const defaultBackupFolder = "./backup/"
const indexFile = ".index.json"

type Document struct {
	Bookmarked     bool
	ID             string
	ModifiedClient string
	Parent         string
	Type           string
	Version        int
	VissibleName   string
}

type IndexEntry struct {
	ID             string
	ModifiedClient string
}

func updateIndex(idx map[string]string) {
	data, err := json.Marshal(&idx)
	if err != nil {
		return
	}

	ioutil.WriteFile(indexFile, data, 0644)
}

func loadIndex(indexFilePath string) (idx map[string]string) {

	fmt.Printf("> Loading index file from %s...\n", indexFilePath)
	data, err := ioutil.ReadFile(indexFilePath)
	if err != nil {
		idx = make(map[string]string)
		fmt.Println(">> No index file found, creating a new one")
		return idx
	}

	err = json.Unmarshal(data, &idx)
	if err != nil {
		os.Remove(indexFilePath)
		idx = make(map[string]string)
		fmt.Printf(">> Corrupted index file found, creating a new one")
		return idx
	}
	fmt.Printf(">> Loaded %d entries from previous backup\n", len(idx))
	return
}

func (d Document) String() string {
	return fmt.Sprintf("#%s / Date %s", d.ID, d.ModifiedClient)
}

func cleanupFileName(name string) string {
	name = strings.Replace(name, "/", "-", -1)
	name = strings.Replace(name, "#", "-", -1)

	if (strings.HasSuffix(name, ".pdf") == false) && (strings.HasSuffix(name, ".epub") == false) {
		name += ".pdf"
	}

	return name
}

func downloadID(ID string, saveFolder string, saveName string) error {
	resp, err := http.Get(remarkableDownloadURL + ID + "/placeholder")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	os.MkdirAll(saveFolder, os.ModePerm)

	err = ioutil.WriteFile(path.Join(saveFolder, cleanupFileName(saveName)), body, 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func dumpFromRoot(backupFolder string, idx map[string]string, currentPath string, rootID string) {
	var documentsList []Document

	fmt.Println("> Saving all from folder " + currentPath)

	resp, err := http.Get(remarkableURL + rootID)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal(body, &documentsList)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, d := range documentsList {
		if d.Type == "CollectionType" {
			dumpFromRoot(backupFolder, idx, path.Join(currentPath, d.VissibleName), d.ID)
		} else {
			fmt.Printf(">> %s| %s: ", d.ID, path.Join(currentPath, d.VissibleName))

			//Check if this needs to be synced
			date, found := idx[d.ID]
			if (found == false) || (date != d.ModifiedClient) {
				fmt.Println("Needs to be synced")
				err := downloadID(d.ID, path.Join(backupFolder, currentPath), d.VissibleName)
				if err != nil {
					fmt.Println(err)
				} else {
					idx[d.ID] = d.ModifiedClient
					updateIndex(idx)
				}
			} else {
				fmt.Println("Up to date")
			}
		}
	}
}

func main() {
	var backupFolder = defaultBackupFolder

	fmt.Printf("Remakable Tablet backup utility v%s\n", rmkdumpVersion)

	if len(os.Args) == 2 {
		backupFolder = os.Args[1]
	}

	idx := loadIndex(path.Join(backupFolder, indexFile))
	dumpFromRoot(backupFolder, idx, "", "")
}
