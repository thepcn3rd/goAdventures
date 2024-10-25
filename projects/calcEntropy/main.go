package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
)

// Use the GOPATH for development and then transition over to the prep script
// go env -w GOPATH="/home/thepcn3rd/go/workspaces/calcEntropy"

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o calcEntropy.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o calcEntropy.exe -ldflags "-w -s" main.go

/*
References:

https://www.ibm.com/docs/en/qsip/7.5?topic=content-analyzing-files-embedded-malicious-activity
"...entropy is used as an indicator of the variability of bits per byte. Because each character in a data unit consists of 1 byte, the entropy value indicates the variation of the characters and the compressibility of the data unit. Variations in the entropy values in the file might indicate that suspect content is hidden in files. For example, the high entropy values might be an indication that the data is stored encrypted and compressed and the lower values might indicate that at runtime the payload is decrypted and stored in different sections. "


http://www.forensickb.com/2013/03/file-entropy-explained.html
"The equation used by Shannon has a resulting value of something between zero (0) and eight (8). The closer the number is to zero, the more orderly or non-random the data is. The closer the data is to the value of eight, the more random or non-uniform the data is."


*/

// Function to calculate the entropy of a byte slice
func calculateEntropy(reader io.Reader) (float64, error) {
	freq := make(map[byte]float64)
	var count float64
	buf := make([]byte, 1024)

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		count += float64(n)
		for _, b := range buf[:n] {
			freq[b]++
		}
	}

	var entropy float64
	for _, f := range freq {
		p := f / count
		entropy -= p * math.Log2(p)
	}
	return entropy, nil
}

// Function to calculate the entropy of a byte slice
func calculateEntropyChunk(data []byte) float64 {
	// Initialize a map to store the frequency of each byte
	frequency := make(map[byte]int)
	for _, b := range data {
		frequency[b]++
	}

	// Calculate the entropy
	var entropy float64
	dataLen := float64(len(data))
	for _, count := range frequency {
		probability := float64(count) / dataLen
		entropy -= probability * math.Log2(probability)
	}

	return entropy
}

func main() {
	var colorReset = "\033[0m"
	var colorGreen = "\033[32m"

	filePtr := flag.String("f", "", "File to read and calculate entropy")
	flag.Parse()

	file, err := os.Open(*filePtr)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	entropy, err := calculateEntropy(file)
	if err != nil {
		fmt.Println("Error calculating entropy", err)
		os.Exit(1)
	}

	fmt.Printf("%sEntropy: %.4f bits/byte%s\n\n", colorGreen, entropy, colorReset)

	// Calculate the chunks of a file...............................................
	chunkSize := 256
	fmt.Printf("%sChunk size is set to: %d\n%s", colorGreen, chunkSize, colorReset)
	fileChunk, err := os.Open(*filePtr)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create a buffer to read 64 bytes at a time
	buffer := make([]byte, chunkSize)
	chunkIndex := 0
	for {
		bytesRead, err := fileChunk.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				break // End of file reached
			}
			log.Fatalf("Failed to read file: %v", err)
		}

		if bytesRead > 0 {
			// Calculate and print the entropy of the chunk
			chunk := buffer[:bytesRead]
			entropyChunk := calculateEntropyChunk(chunk)
			// Display the chunk if the entropy is higher than the base entropy...
			if entropyChunk > entropy {
				//if entropyChunk != entropy {
				fmt.Printf("\n%sEntropy of chunk %d: %f%s\n", colorGreen, chunkIndex, entropyChunk, colorReset)

				// Output the hex of the chunk
				var hexStr string
				//var hexStrArray []string
				var asciiStr string
				//var asciiStrArray []string
				var increment int
				for index, b := range chunk {
					hexStr += fmt.Sprintf("%02x ", b)
					if b >= 32 && b <= 126 { // printable ASCII range
						asciiStr += fmt.Sprintf("%c ", b)
					} else {
						asciiStr += ". " // fmt.Printf(".") // non-printable characters are replaced with a dot
					}
					if index%32 == 31 {
						fmt.Printf("%03d - %s | %s\n", index, hexStr, asciiStr)
						hexStr = ""
						//fmt.Printf("ASCII: %s\n", asciiStr)
						asciiStr = ""
					}
					increment = index
				}
				if increment != 255 {
					fmt.Printf("%03d - %s | %s\n", increment, hexStr, asciiStr)
				}

			}
			chunkIndex++
		}

		if bytesRead < len(buffer) {
			break // Last chunk read
		}
	}
	fmt.Printf("\n%sTotal Chunks: %d%s\n\n", colorGreen, chunkIndex, colorReset)

}
