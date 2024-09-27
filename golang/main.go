package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

const url = "https://storage.googleapis.com/panels-api/data/20240916/media-1a-i-p~s"

// data struct to get only dhd stuff
type Data struct {
	Data map[string]struct {
		Dhd string `json:"dhd"`
	} `json:"data"`
}

func main() {
	response, error := fetchData(url)
	if error != nil {
		fmt.Printf("Failed to fetch data: %v\n", error)
		return
	}

	jsonData, error := parseData(response)
	if error != nil {
		fmt.Printf("Failed to parse data: %v\n", error)
		return
	}

	// get the total number of images to download
	imageCount := countImages(jsonData)
	fmt.Printf("Total images to download: %d\n", imageCount)

	// ensure that the downloads directory exists in the same directory as the executable
	downloadDir := "downloads"

	if err := createDirectory(downloadDir); err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		return
	}

	bar := progressbar.Default(int64(imageCount))
	fileIndex := 1

	// iterate over the data with dhd prop and download each image
	for _, subproperty := range jsonData.Data {
		if subproperty.Dhd != "" {
			imageUrl := subproperty.Dhd

			// download the image and save it
			err := downloadImage(imageUrl, downloadDir, fileIndex)
			if err != nil {
				fmt.Printf("Error downloading image: %v\n", err)
				continue
			}

			bar.Add(1)
			fileIndex++
		}
	}

	fmt.Printf("%d/%d images downloaded successfully\n", fileIndex, imageCount)
}

func fetchData(url string) (*http.Response, error) {
	response, error := http.Get(url)

	if error != nil {
		return nil, fmt.Errorf("failed to fetch JSON file: %v", error)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JSON file: %s", response.Status)
	}

	return response, nil
}

func parseData(response *http.Response) (Data, error) {
	var jsonData Data
	error := json.NewDecoder(response.Body).Decode(&jsonData)

	if error != nil {
		return Data{}, fmt.Errorf("failed to parse JSON: %v", error)
	}

	return jsonData, nil
}

// countImages returns the total number of images to download
func countImages(jsonData Data) int {
	count := 0
	for _, subproperty := range jsonData.Data {
		if subproperty.Dhd != "" {
			count++
		}
	}

	return count
}

func createDirectory(dir string) error {
	if _, error := os.Stat(dir); os.IsNotExist(error) {
		error = os.Mkdir(dir, os.ModePerm)
		if error != nil {
			return error
		}
		fmt.Printf("Created directory: %s\n", dir)
	}
	return nil
}

// downloadImage downloads an image from a url and saves it to the download directory
func downloadImage(imageUrl, downloadDir string, fileIndex int) error {
	// fetch the image
	response, error := http.Get(imageUrl)
	if error != nil {
		return fmt.Errorf("failed to download image: %v", error)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download image: %s", response.Status)
	}

	// remove all query parameters from the image url
	imageUrlWithoutParams := strings.Split(imageUrl, "?")[0]

	// create the image file
	fileExtension := filepath.Ext(imageUrlWithoutParams)
	filePath := filepath.Join(downloadDir, fmt.Sprintf("%d%s", fileIndex, fileExtension))
	file, error := os.Create(filePath)
	if error != nil {
		return fmt.Errorf("failed to create file: %v", error)
	}
	defer file.Close()

	// write the image data to file
	_, error = io.Copy(file, response.Body)
	if error != nil {
		return fmt.Errorf("failed to save image: %v", error)
	}

	return nil
}
