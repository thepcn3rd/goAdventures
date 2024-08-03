package main

// To cross compile windows
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o passwordGenerator.bin -ldflags "-w -s" main.go

/* Future:

- Add custom word to add in the mix...
- Create the criteria for the length of the password, attempt to retry the creation if the length is not met

*/

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
)

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func patternDefinition() {

	fmt.Println("\nPattern definition:")
	fmt.Println("\ts - Special Character i.e. (!@#$%+-=)")
	fmt.Println("\ti - Positive Integer")
	fmt.Println("\tA - Upper-case letter")
	fmt.Println("\ta - Lower-case letter")
	fmt.Println("\tw - Word")
	fmt.Println("\tW - Word with l33t lettering i.e. (H3ll0 or G00dby3)")
	fmt.Println("\tC - Custom String (i.e. Summer2023)")
	fmt.Println("\tZ - Custom String with l33t")
}

func selectWord(leet bool, wFile string) string {
	var stringValue string
	var scanner *bufio.Scanner
	if wFile == "None" {
		stringValue = dictionaryLoad()
		//fmt.Println(stringValue)
		// Create a scanner to read lines from the input string
		scanner = bufio.NewScanner(strings.NewReader(stringValue))
	} else {
		f, err := os.Open(wFile)
		if err != nil {
			fmt.Println("Unable to open file...")
			log.Fatal(err)
		}
		defer f.Close()
		scanner = bufio.NewScanner(f)
	}

	// Count the lines
	lineCount := 0

	// Read and process each line
	for scanner.Scan() {
		//line := scanner.Text()
		//fmt.Println("Read line:", line)
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error:", err)
	}

	maxN := lineCount - 1
	minN := 0
	randomNum := rand.Intn(maxN - minN)

	// Reset the Line Count
	lineCount = 0
	// Restablish the scanner after it was read
	if wFile == "None" {
		scanner = bufio.NewScanner(strings.NewReader(stringValue))
	} else {
		f, err := os.Open(wFile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		scanner = bufio.NewScanner(f)
	}
	var randomWord string
	for scanner.Scan() {
		if lineCount == randomNum {
			randomWord = scanner.Text()
			//fmt.Println("Random Word A:", randomWordA)
		}
		lineCount++
	}

	if leet == true {
		randomWord = makeLeet(randomWord)
	}

	return randomWord
}

func makeLeet(rWord string) string {
	rWord = strings.Replace(rWord, "s", "$", -1)
	rWord = strings.Replace(rWord, "o", "0", -1)
	rWord = strings.Replace(rWord, "i", "1", -1)
	rWord = strings.Replace(rWord, "e", "3", -1)
	rWord = strings.Replace(rWord, "a", "@", -1)
	rWord = strings.Replace(rWord, "d", "#", -1)
	rWord = strings.Replace(rWord, "b", "8", -1)
	rWord = strings.Replace(rWord, "h", "4", -1)
	rWord = strings.Replace(rWord, "v", "^", -1)
	return rWord
}

func selectRandomInteger() string {
	numbers := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	n := rand.Int() % len(numbers)
	return numbers[n]
}

func selectSpecialChar() string {
	specialChar := []string{"!", "@", "#", "$", "%", "+", "-", "="}
	s := rand.Int() % len(specialChar)
	return specialChar[s]
}

func selectLetter(lower bool) string {
	const lowerCase = "abcdefghijklmnopqrstuvwxyz"
	const upperCase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if lower == true {
		l := rand.Int() % len(lowerCase)
		return string(lowerCase[l])
	} else {
		u := rand.Int() % len(upperCase)
		return string(upperCase[u])
	}
}

func customString(leet bool) string {
	var customStr string
	fmt.Print("\nEnter custom string: ")
	fmt.Scanf("%s", &customStr)
	if leet == true {
		customStr = makeLeet(customStr)
	}
	return customStr
}

/*

To create the dictionary below you can read in a wordlist with this go function...
package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	// Pulled the opensource version of an English dictionary
	filename := "sample3.txt" // Change this to the path of your file

	// Read the entire file into a byte array
	byteArray, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	for _, value := range byteArray {
		fmt.Printf("%d, ", value)
	}
}

*/

func dictionaryLoad() string {
	dict := []byte{97, 97, 114, 100, 45, 118, 97, 114, 107, 13, 10, 97, 97, 114, 100, 45, 119, 111, 108, 102, 13, 10, 97, 97, 114, 111, 110, 105, 99, 10, 97, 97, 114, 111, 110, 39, 115, 10, 97, 98, 97, 99, 97, 13, 10, 97, 98, 97, 99, 105, 110, 97, 116, 101, 13, 10, 97, 98, 97, 99, 105, 110, 97, 116, 105, 111, 110, 13, 10, 97, 98, 97, 99, 105, 115, 99, 117, 115, 13, 10, 97, 98, 97, 99, 105, 115, 116, 13, 10, 97, 98, 97, 99, 107, 13, 10, 97, 98, 97, 99, 107, 13, 10, 97, 98, 97, 99, 116, 105, 110, 97, 108, 13, 10, 97, 98, 97, 99, 116, 105, 111, 110, 13, 10, 97, 98, 97, 99, 116, 111, 114, 13, 10, 97, 98, 97, 99, 117, 108, 117, 115, 13, 10, 97, 98, 97, 99, 117, 115, 13, 10, 97, 98, 97, 100, 97, 13, 10, 97, 98, 97, 100, 100, 111, 110, 13, 10, 97, 98, 97, 102, 116, 13, 10, 97, 98, 97, 102, 116, 13, 10, 97, 98, 97, 105, 115, 97, 110, 99, 101, 13, 10, 97, 98, 97, 105, 115, 101, 114, 13, 10, 97, 98, 97, 105, 115, 116, 13, 10, 97, 98, 97, 108, 105, 101, 110, 97, 116, 101, 13, 10, 97, 98, 97, 108, 105, 101, 110, 97, 116, 105, 111, 110, 13, 10, 97, 98, 97, 108, 111, 110, 101, 13, 10, 97, 98, 97, 110, 100, 13, 10, 97, 98, 97, 110, 100, 111, 110, 13, 10, 97, 98, 97, 110, 100, 111, 110, 13, 10, 97, 98, 97, 110, 100, 111, 110, 13, 10, 97, 98, 97, 110, 100, 111, 110, 101, 100, 13, 10, 97, 98, 97, 110, 100, 111, 110, 101, 100, 108, 121, 13, 10, 97, 98, 97, 110, 100, 111, 110, 101, 101, 13, 10, 97, 98, 97, 110, 100, 111, 110, 101, 114, 13, 10, 97, 98, 97, 110, 100, 111, 110, 109, 101, 110, 116, 13, 10, 97, 98, 97, 110, 100, 117, 109, 13, 10, 97, 98, 97, 110, 101, 116, 13, 10, 97, 98, 97, 110, 103, 97, 13, 10, 97, 98, 97, 110, 110, 97, 116, 105, 111, 110, 10, 97, 98, 97, 114, 116, 105, 99, 117, 108, 97, 116, 105, 111, 110, 13, 10, 97, 98, 97, 115, 101, 13, 10, 97, 98, 97, 115, 101, 100, 13, 10, 97, 98, 97, 115, 101, 100, 108, 121, 13, 10, 97, 98, 97, 115, 101, 109, 101, 110, 116, 13, 10, 97, 98, 97, 115, 101, 114, 13, 10, 97, 98, 97, 115, 104, 13, 10, 97, 98, 97, 115, 104, 101, 100, 108, 121, 13, 10, 97, 98, 97, 115, 104, 109, 101, 110, 116, 13, 10, 97, 98, 97, 115, 105, 97, 13, 10, 97, 98, 97, 115, 115, 105, 10, 97, 98, 97, 116, 97, 98, 108, 101, 13, 10, 97, 98, 97, 116, 101, 13, 10, 97, 98, 97, 116, 101, 13, 10, 97, 98, 97, 116, 101, 13, 10, 97, 98, 97, 116, 101, 109, 101, 110, 116, 13, 10, 97, 98, 97, 116, 101, 114, 13, 10, 97, 98, 97, 116, 105, 115, 10, 97, 98, 97, 116, 105, 115, 101, 100, 13, 10, 97, 98, 97, 116, 111, 114, 13, 10, 97, 98, 97, 116, 116, 111, 105, 114, 13, 10, 97, 98, 97, 116, 117, 114, 101, 13, 10, 97, 98, 97, 116, 118, 111, 105, 120, 13, 10, 97, 98, 97, 119, 101, 100, 13, 10, 97, 98, 97, 120, 105, 97, 108, 10, 97, 98, 97, 121, 13, 10, 97, 98, 98, 97, 13, 10, 97, 98, 98, 97, 99, 121, 13, 10, 97, 98, 98, 97, 116, 105, 97, 108, 13, 10, 97, 98, 98, 97, 116, 105, 99, 97, 108, 13, 10, 97, 98, 98, 101, 13, 10, 97, 98, 98, 101, 115, 115, 13, 10, 97, 98, 98, 101, 121, 13, 10, 97, 98, 98, 111, 116, 13, 10, 97, 98, 98, 111, 116, 115, 104, 105, 112, 13, 10, 97, 98, 98, 114, 101, 118, 105, 97, 116, 101, 13, 10, 97, 98, 98, 114, 101, 118, 105, 97, 116, 101, 13, 10, 97, 98, 98, 114, 101, 118, 105, 97, 116, 101, 13, 10, 97, 98, 98, 114, 101, 118, 105, 97, 116, 101, 100, 13, 10, 97, 98, 98, 114, 101, 118, 105, 97, 116, 105, 111, 110, 13, 10, 97, 98, 98, 114, 101, 118, 105, 97, 116, 111, 114, 13, 10, 97, 98, 98, 114, 101, 118, 105, 97, 116, 111, 114, 121, 13, 10, 97, 98, 98, 114, 101, 118, 105, 97, 116, 117, 114, 101, 13, 10, 97, 98, 100, 97, 108, 13, 10, 97, 98, 100, 101, 114, 105, 97, 110, 13, 10, 97, 98, 100, 101, 114, 105, 116, 101, 13, 10, 97, 98, 100, 101, 115, 116, 13, 10, 97, 98, 100, 105, 99, 97, 98, 108, 101, 13, 10, 97, 98, 100, 105, 99, 97, 110, 116, 13, 10, 97, 98, 100, 105, 99, 97, 110, 116, 13, 10, 97, 98, 100, 105, 99, 97, 116, 101, 13, 10, 97, 98, 100, 105, 99, 97, 116, 101, 13, 10, 97, 98, 100, 105, 99, 97, 116, 105, 111, 110, 13, 10, 97, 98, 100, 105, 99, 97, 116, 105, 118, 101, 13, 10, 97, 98, 100, 105, 99, 97, 116, 111, 114, 13, 10, 97, 98, 100, 105, 116, 105, 118, 101, 13, 10, 97, 98, 100, 105, 116, 111, 114, 121, 13, 10, 97, 98, 100, 111, 109, 101, 110, 13, 10, 97, 98, 100, 111, 109, 105, 110, 97, 108, 13, 10, 97, 98, 100, 111, 109, 105, 110, 97, 108, 13, 10, 97, 98, 100, 111, 109, 105, 110, 97, 108, 101, 115, 13, 10}
	return string(dict)
}

func main() {
	var patternSelected string
	var wordlistFile string
	patternSelected = "Aswia" // Default
	patternPtr := flag.String("pattern", "Aswia", "Pattern to create passwords")
	lengthPtr := flag.String("length", "8", "Minimum length of the password")
	wordlistPtr := flag.String("wordlist", "None", "Reads a custom wordlist (Uses default wordlist contained)")
	flag.Parse()

	fmt.Printf("\nMinimum length of the password is: %s\n", *lengthPtr)

	if !isFlagPassed("wordlist") {
		wordlistFile = "None"
	} else {
		wordlistFile = *wordlistPtr
	}

	if isFlagPassed("pattern") {
		if len(*patternPtr) <= 2 {
			fmt.Println("Length of the pattern must be 3 characters or longer.  (i.e. siw)")
			patternDefinition()
			fmt.Println("")
			flag.Usage()
			fmt.Println("")
			os.Exit(0)
		} else {
			patternSelected = *patternPtr
			patternDefinition()
		}
	} else {
		patternDefinition()
	}
	// Left for troubleshooting...
	//fmt.Printf("Default pattern to create a password: %s\n", patternSelected)
	//fmt.Printf("\nWord selected from the list: %s\n", selectWord())
	//fmt.Printf("Upper-case Letter selected from the list: %s\n", selectLetter(false))
	//fmt.Printf("Lower-case Letter selected from the list: %s\n", selectLetter(true))
	//fmt.Printf("Random integer selected: %s\n", selectRandomInteger())
	//fmt.Printf("Special character selected: %s\n\n", selectSpecialChar())

	outputString := ""
	for _, p := range patternSelected {
		//fmt.Println(string(p))
		patternItem := string(p)
		switch patternItem {
		case "A":
			outputString = outputString + selectLetter(false)
		case "a":
			outputString = outputString + selectLetter(true)
		case "i":
			outputString = outputString + selectRandomInteger()
		case "s":
			outputString = outputString + selectSpecialChar()
		case "w":
			outputString = outputString + selectWord(false, wordlistFile)
		case "W":
			outputString = outputString + selectWord(true, wordlistFile)
		case "C":
			outputString = outputString + customString(false)
		case "Z":
			outputString = outputString + customString(true)
		default:
			//donothing
		}

	}

	fmt.Printf("\nConstructed Password:\n%s\n\n", outputString)

}
