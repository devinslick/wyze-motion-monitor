package main

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Payload struct {
	CameraName string `json:"cameraName"`
	JPGPath    string `json:"jpgPath"`
	MP4Path    string `json:"mp4Path"`
}

func main() {
	// Check if the required command-line arguments are provided
	if len(os.Args) < 3 {
		log.Fatal("Usage: ./program <cameraName> <webhookURL>")
	}

	cameraName := os.Args[1]
	webhookURL := os.Args[2]

	// Load the last sent payload from the saved file
	lastPayload, err := loadLastPayload()
	if err != nil {
		log.Printf("Error while loading last payload: %v", err)
	}

	// Create a channel to receive termination signals
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt, syscall.SIGTERM)

	// Create a wait group to wait for goroutines to finish
	var wg sync.WaitGroup

	// Start monitoring image files
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorFiles("/media/mmc/alarm", "*.jpg", cameraName, webhookURL, lastPayload)
	}()

	// Start monitoring video files
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorFiles("/media/mmc/record", "*.mp4", cameraName, webhookURL, lastPayload)
	}()

	// Wait for termination signal
	<-terminate

	// Signal termination to goroutines
	wg.Wait()
}

// Monitor files in the specified folder with the given extension
func monitorFiles(folderPath, extension, cameraName, webhookURL string, lastPayload *Payload) {
	var lastModified time.Time

	for {
		files, err := filepath.Glob(filepath.Join(folderPath, "*", extension))
		if err != nil {
			log.Printf("Error while searching for files: %v", err)
		}

		if len(files) > 0 {
			latestFile := files[len(files)-1]
			fileInfo, err := os.Stat(latestFile)
			if err != nil {
				log.Printf("Error while getting file information: %v", err)
				continue
			}

			if fileInfo.ModTime().After(lastModified) {
				lastModified = fileInfo.ModTime()

				// Send JSON payload to webhook only if it's different from the last sent payload
				payload := Payload{
					CameraName: cameraName,
					JPGPath:    latestFile,
					MP4Path:    strings.Replace(latestFile, "/alarm/", "/record/", 1),
				}

				if !isDuplicatePayload(payload, lastPayload) {
					sendPayload(webhookURL, payload)
					lastPayload = &payload
					saveLastPayload(lastPayload)
				}
			}
		}

		time.Sleep(1 * time.Second) // Sleep for 1 second before checking again
	}
}

// Check if the payload is a duplicate of the last sent payload
func isDuplicatePayload(payload Payload, lastPayload *Payload) bool {
	return lastPayload != nil &&
		payload.CameraName == lastPayload.CameraName &&
		payload.JPGPath == lastPayload.JPGPath &&
		payload.MP4Path == lastPayload.MP4Path
}

// Load the last sent payload from the saved file
func loadLastPayload() (*Payload, error) {
	data, err := ioutil.ReadFile("last_payload.json")
	if err != nil {
		return nil, err
	}

	payload := &Payload{}
	err = json.Unmarshal(data, payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// Save the last sent payload to the file
func saveLastPayload(payload *Payload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("last_payload.json", data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Send JSON payload to the specified webhook URL
func sendPayload(webhookURL string, payload Payload) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error while marshaling JSON payload: %v", err)
		return
	}

	transport := &http.Transport{
	    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        }
        client := &http.Client{Transport: transport}

        // Use the custom HTTP client for sending the request
        resp, err := client.Post(webhookURL, "application/json", strings.NewReader(string(jsonPayload)))

	if err != nil {
		log.Printf("Error while sending payload to webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Webhook request failed with status: %s", resp.Status)
	}
}
