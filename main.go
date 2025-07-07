package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

	var readFileContent string
	// If HTML file exists, read its content into a string
	if fileExists(localHTMLFileRemoteLocation) {
		readFileContent = readAFileAsString(localHTMLFileRemoteLocation)
	}

	// Extract all PDF URLs from the HTML content
	extractedPDFURLOnly := extractPDFUrls(readFileContent)
	// Remove duplicate PDF URLs
	extractedPDFURLOnly = removeDuplicatesFromSlice(extractedPDFURLOnly)

	// Extract all ZIP URLs from the HTML content
	extractedZIPFilesOnly := extractZIPUrls(readFileContent)
	// Remove duplicate ZIP URLs
	extractedPDFURLOnly = removeDuplicatesFromSlice(extractedPDFURLOnly)

	// Download each valid and unique PDF file
	for _, url := range extractedPDFURLOnly {
		url = "https://parrot.com" + url
		if isUrlValid(url) {
			downloadPDF(url, outputDir)
		}
	}

	// Download each valid and unique ZIP file
	for _, url := range extractedZIPFilesOnly {
		url = "https://parrot.com" + url
		if isUrlValid(url) {
			downloadZIP(url, zipOutputDir)
		}
	}

}

// downloadZIP downloads a ZIP file from a URL to the specified directory
func downloadZIP(finalURL string, outputDir string) {
	// Generate safe filename from URL
	rawFilename := urlToFilename(finalURL)
	filename := strings.ToLower(rawFilename)

	// Build the full file path
	filePath := filepath.Join(outputDir, filename)
	// Skip download if file already exists
	if fileExists(filePath) {
		log.Printf("file already exists, skipping: %s", filePath)
		return
	}

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 3 * time.Minute}
	resp, err := client.Get(finalURL)
	if err != nil {
		log.Printf("error fetching %s: %v", finalURL, err)
		return
	}
	defer resp.Body.Close()

	// Check for valid HTTP response
	if resp.StatusCode != http.StatusOK {
		log.Printf("bad status code for %s: %d", finalURL, resp.StatusCode)
		return
	}

	// Check content type to ensure it's a ZIP file
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/zip") && !strings.Contains(ct, "application/octet-stream") && !strings.Contains(ct, "text/html") {
		log.Printf("unexpected content type for %s: %s", finalURL, ct)
		return
	}

	// Read response body into memory
	var buf bytes.Buffer
	written, err := buf.ReadFrom(resp.Body)
	if err != nil || written == 0 {
		log.Printf("error reading body from %s: %v", finalURL, err)
		return
	}

	// Create and write to local file
	out, err := os.Create(filePath)
	if err != nil {
		log.Printf("error creating file %s: %v", filePath, err)
		return
	}
	defer out.Close()

	if _, err := buf.WriteTo(out); err != nil {
		log.Printf("error writing to file %s: %v", filePath, err)
		return
	}

	log.Printf("downloaded: %s", filePath)
}

// downloadPDF downloads a PDF file from a URL to the specified directory
func downloadPDF(finalURL string, outputDir string) {
	rawFilename := urlToFilename(finalURL)
	filename := strings.ToLower(rawFilename)
	filePath := filepath.Join(outputDir, filename)
	if fileExists(filePath) {
		log.Printf("file already exists, skipping: %s", filePath)
		return
	}
	client := &http.Client{Timeout: 3 * time.Minute}
	resp, err := client.Get(finalURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/pdf") {
		return
	}
	var buf bytes.Buffer
	written, err := buf.ReadFrom(resp.Body)
	if err != nil || written == 0 {
		return
	}
	out, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer out.Close()
	buf.WriteTo(out)
}

// urlToFilename creates a filesystem-safe filename from a URL string.
func urlToFilename(rawURL string) string {
	// Extract the file extension from the URL (e.g., ".jpg", ".pdf").
	remoteURLFileEXT := getFileExtension(rawURL)
	// Get the base name of the path (e.g., "image.jpg" from "http://example.com/path/image.jpg").
	baseName := filepath.Base(rawURL)
	// Convert the base name to lowercase for normalization.
	lower := strings.ToLower(baseName)
	// Replace all non-alphanumeric characters with underscores.
	reNonAlnum := regexp.MustCompile(`[^a-z0-9]`)
	safe := reNonAlnum.ReplaceAllString(lower, "_")
	// Collapse multiple underscores into a single underscore.
	safe = regexp.MustCompile(`_+`).ReplaceAllString(safe, "_")
	// List of substrings to remove if they appear in the middle of the name.
	var invalidSubstrings = []string{"_zip", "_pdf"}
	for _, invalidPre := range invalidSubstrings {
		// Remove the unwanted substrings from the filename.
		safe = removeSubstring(safe, invalidPre)
	}
	// Remove leading underscore if it exists.
	if after, ok := strings.CutPrefix(safe, "_"); ok {
		safe = after
	}
	// Ensure the final filename ends with the correct extension.
	if getFileExtension(safe) != remoteURLFileEXT {
		safe = safe + remoteURLFileEXT
	}
	// Return the sanitized filename.
	return safe
}

// getFileExtension returns the file extension of the path
func getFileExtension(path string) string {
	return filepath.Ext(path)
}

// removeSubstring removes all instances of a substring from a string
func removeSubstring(input string, toRemove string) string {
	return strings.ReplaceAll(input, toRemove, "")
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
