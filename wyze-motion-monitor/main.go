package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Payload represents the JSON payload sent in the webhook
type Payload struct {
	CameraName string `json:"cameraName"`
	JPGPath    string `json:"jpgPath"`
}

// Function to send the payload as a webhook
func sendPayload(webhookURL string, payload Payload) error {
	if webhookURL == "" {
		// Webhook URL not provided, skip sending the payload
		return nil
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	// Use the custom HTTP client for sending the request
	resp, err := client.Post(webhookURL, "application/json", strings.NewReader(string(jsonPayload)))

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	return nil
}

// Function to copy a file
func copyFile(srcPath, destPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

// Function to monitor the latest folder for new JPG files
func monitorJPGFiles(folderPath, cameraName, webhookURL string) {
	lastJPGPath := readLastJPGPath()

	for {
		// Get the list of directories in the specified folder
		dirList, err := getDirectories(folderPath)
		if err != nil {
			log.Printf("Error while getting directory list: %v", err)
			continue
		}

		if len(dirList) > 0 {
			// Sort the directories in descending order based on their names
			sortDirectoriesDescending(dirList)

			latestDir := dirList[0]
			dirPath := filepath.Join(folderPath, latestDir)

			// Get the list of JPG files in the latest directory
			files, err := filepath.Glob(filepath.Join(dirPath, "*.jpg"))
			if err != nil {
				log.Printf("Error while searching for JPG files: %v", err)
				continue
			}

			if len(files) > 0 {
				// Sort the files to ensure they are in ascending order
				sortFiles(files)

				// Find the latest JPG file
				latestJPGPath := files[len(files)-1]

				if latestJPGPath != lastJPGPath {
					lastJPGPath = latestJPGPath

					// Copy the latest JPG file to /media/mmc/wz_mini/www/latest.jpg
					err = copyFile(latestJPGPath, "/media/mmc/wz_mini/www/latest.jpg")
					if err != nil {
						log.Printf("Error while copying JPG file: %v", err)
					} else {
						log.Println("Copied latest JPG file to /media/mmc/wz_mini/www/latest.jpg")
					}

					// Send JSON payload to webhook if webhook URL is provided
					if webhookURL != "" {
						payload := Payload{
							CameraName: cameraName,
							JPGPath:    latestJPGPath,
						}

						err := sendPayload(webhookURL, payload)
						if err != nil {
							log.Printf("Error while sending payload to webhook: %v", err)
						} else {
							log.Printf("Sent webhook for JPG file: %s", latestJPGPath)
						}
					}

					// Save the last JPG path to a file
					err := saveLastJPGPath(lastJPGPath)
					if err != nil {
						log.Printf("Error while saving last JPG path: %v", err)
					}
				}
			}
		}

		time.Sleep(1 * time.Second) // Sleep for 1 second before checking again
	}
}

// Function to read the last JPG path from a file
func readLastJPGPath() string {
	filePath := "/media/mmc/wz_mini/www/last_jpg_path.txt"
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		log.Printf("Error while reading last JPG path file: %v", err)
		return ""
	}
	return string(data)
}

// Function to save the last JPG path to a file
func saveLastJPGPath(path string) error {
	filePath := "/media/mmc/wz_mini/www/last_jpg_path.txt"
	err := ioutil.WriteFile(filePath, []byte(path), 0644)
	if err != nil {
		return err
	}
	return nil
}

// Function to get the list of directories in the specified folder
func getDirectories(folderPath string) ([]string, error) {
	var dirList []string

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error while walking directory: %v", err)
			return nil
		}

		if info.IsDir() && path != folderPath {
			dirList = append(dirList, info.Name())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return dirList, nil
}

// Function to sort the directories in descending order based on their names
func sortDirectoriesDescending(dirList []string) {
	sort.Slice(dirList, func(i, j int) bool {
		return dirList[i] > dirList[j]
	})
}

// Function to sort the files in ascending order
func sortFiles(files []string) {
	sort.Strings(files)
}

func main() {
	// Retrieve command-line arguments
	cameraName := os.Args[1]
	var webhookURL string
	if len(os.Args) > 2 {
		webhookURL = os.Args[2]
	}

	// Specify the folder path for JPG monitoring
	jpgFolderPath := "/media/mmc/alarm"

	log.Printf("Starting JPG monitoring for Camera: %s", cameraName)

	// Start monitoring JPG files
	go monitorJPGFiles(jpgFolderPath, cameraName, webhookURL)

	// Wait indefinitely
	select {}
}
