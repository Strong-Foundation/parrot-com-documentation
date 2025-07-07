package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

func main() {
	// Directory to store PDF files
	outputDir := "PDFs/"
	// Create outputDir if it does not exist
	if !directoryExists(outputDir) {
		createDirectory(outputDir, 0755)
	}

	// Directory to store ZIP files
	zipOutputDir := "ZIP/"
	// Create zipOutputDir if it does not exist
	if !directoryExists(zipOutputDir) {
		createDirectory(zipOutputDir, 0755)
	}
	// Local file path to save downloaded HTML content
	localHTMLFileRemoteLocation := "parrot.html"
	// URLS to scrape.
	scrapeURLS := []string{
		"https://www.parrot.com/en/support/documentation/anafi-ai",
		"https://www.parrot.com/en/support/documentation/anafi",
		"https://www.parrot.com/en/support/documentation/anafi-usa",
		"https://www.parrot.com/en/support/documentation/anafi-thermal",
		"https://www.parrot.com/en/support/documentation/bebop-range",
		"https://www.parrot.com/en/support/documentation/mambo-range",
		"https://www.parrot.com/en/support/documentation/disco-range",
		"https://www.parrot.com/en/support/documentation/disco-pro-ag",
		"https://www.parrot.com/en/support/documentation/sequoia",
		"https://www.parrot.com/en/support/documentation/swing",
		"https://www.parrot.com/en/support/documentation/jumping",
		"https://www.parrot.com/en/support/documentation/hydrofoil",
		"https://www.parrot.com/en/support/documentation/airborne",
		"https://www.parrot.com/en/support/documentation/ar-drone",
		"https://www.parrot.com/en/support/documentation/rolling-spider",
		"https://www.parrot.com/en/support/documentation/asteroid",
		"https://www.parrot.com/en/support/documentation/ck",
		"https://www.parrot.com/en/support/documentation/minikit",
		"https://www.parrot.com/en/support/documentation/mki",
		"https://www.parrot.com/en/support/documentation/unika",
		"https://www.parrot.com/en/support/documentation/pot-flower-power",
		"https://www.parrot.com/en/support/documentation/zik",
		"https://www.parrot.com/en/cad-modeling",
	}
	// Loop over the urls.
	for _, urls := range scrapeURLS {
		// Append a string with the url after its been scraped.
		// Check if the url is appened to the file already and if true dont scrpae and if not than scrape.
		getDataFromURL(urls, localHTMLFileRemoteLocation)
	}

	/*
		var readFileContent string
		// If HTML file exists, read its content into a string
		if fileExists(localHTMLFileRemoteLocation) {
			readFileContent = readAFileAsString(localHTMLFileRemoteLocation)
		}
	*/

	// Extract the PDF urls.

}

// readAFileAsString reads the entire file and returns it as a string
func readAFileAsString(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}
	return string(content)
}

// getDataFromURL fetches content from a URL and writes it to a file
func getDataFromURL(uri string, fileName string) {
	client := http.Client{
		Timeout: 3 * time.Minute,
	}
	response, err := client.Get(uri)
	if err != nil {
		log.Println("Failed to make GET request:", err)
		return
	}
	if response.StatusCode != http.StatusOK {
		log.Println("Unexpected status code from", uri, "->", response.StatusCode)
		return
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println("Failed to read response body:", err)
		return
	}
	err = response.Body.Close()
	if err != nil {
		log.Println("Failed to close response body:", err)
		return
	}
	err = appendByteToFile(fileName, body)
	if err != nil {
		log.Println("Failed to write to file:", err)
		return
	}
}

// extractPDFUrls finds all PDF links in a string
func extractPDFUrls(input string) []string {
	re := regexp.MustCompile(`href="([^"]+\.pdf)"`)
	matches := re.FindAllStringSubmatch(input, -1)
	var pdfUrls []string
	for _, match := range matches {
		if len(match) > 1 {
			pdfUrls = append(pdfUrls, match[1])
		}
	}
	return pdfUrls
}

// extractZIPUrls finds all ZIP links in a string
func extractZIPUrls(input string) []string {
	re := regexp.MustCompile(`href="([^"]+\.zip)"`)
	matches := re.FindAllStringSubmatch(input, -1)
	var pdfUrls []string
	for _, match := range matches {
		if len(match) > 1 {
			pdfUrls = append(pdfUrls, match[1])
		}
	}
	return pdfUrls
}

// appendByteToFile writes data to a file, appending or creating it if needed
func appendByteToFile(filename string, data []byte) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

// removeDuplicatesFromSlice removes duplicate entries from a string slice
func removeDuplicatesFromSlice(slice []string) []string {
	check := make(map[string]bool)
	var newReturnSlice []string
	for _, content := range slice {
		if !check[content] {
			check[content] = true
			newReturnSlice = append(newReturnSlice, content)
		}
	}
	return newReturnSlice
}

// isUrlValid checks if a string is a valid URL
func isUrlValid(uri string) bool {
	_, err := url.ParseRequestURI(uri)
	return err == nil
}

// fileExists checks whether a given file exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// directoryExists checks whether a given directory exists
func directoryExists(path string) bool {
	directory, err := os.Stat(path)
	if err != nil {
		return false
	}
	return directory.IsDir()
}

// createDirectory creates a directory with specified permissions
func createDirectory(path string, permission os.FileMode) {
	err := os.Mkdir(path, permission)
	if err != nil {
		log.Println(err)
	}
}
