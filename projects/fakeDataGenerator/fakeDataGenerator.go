package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*
Purpose:
The purpose of this program is to generate fake SSNs, Credit Card Numbers, Birth Dates and other information of the like.  With the information
generated then use it to simulate an adversary exfiltrating the information out of a network.  Could also be used to encrypt the data as would ransomware
and then exfiltrate it.

Setup the Environment

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/fakeDataGenerator"

Make the directories - src
Copy the commonFunctions folder into the src directory so that it can be referenced

// To cross compile for linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o fakeDataGenerator.bin -ldflags "-w -s" main.go

// To cross compile windows
GOOS=windows GOARCH=amd64 go build -o fakeDataGenerator.exe -ldflags "-w -s" main.go

References:
ChatGPT was used to gather ideas - 11/2/2023
First and Last Names are from SecLists Top 1000 Male, Female and Family Names
Top 100 email domains - http://www.emaildiscussions.com/showthread.php?t=74831 (Did not check for accuracy...)

# Future task to create different styles of usernames/emails first.last, firstInitial.Last, first.LastInitial

# Future task is to add random addresses (maybe associate with phone numbers ??)
openaddresses.io used to gather random addresses from each state...
https://batch.openaddresses.io/data#map=0/0/0

Batch script to generate 250+ files... (Uses 3.5GB of disk space...)
mkdir information
mkdir information\archive2010
mkdir information\archive2000
mkdir information\archive
fakeDataGenerator.exe -q 10000 > information\list.csv
FOR /L %%X IN (0,1,3) DO fakeDataGenerator.exe -q 100000 > information\list_202%%X.csv
FOR /L %%X IN (0,1,9) DO fakeDataGenerator.exe -q 100000 > information\archive2010\list_201%%X.csv
FOR /L %%X IN (0,1,9) DO fakeDataGenerator.exe -q 100000 > information\archive2000\list_200%%X.csv
FOR /L %%X IN (0,1,250) DO fakeDataGenerator.exe -q 100000 > information\archive\list-%%X.csv

Create a config.json file or as many as you need to read in a specified config like the following.
Note: The display if toggled between "y" and "n" will output the column.  The order is the order in
	  in the columns of the csv that it will be printed

{
        "outputQuantity": 5,
        "originalFileOutput": "original.txt",
		"fileOutput": "output.txt",
        "creditCard": {
                "visaBinNumber": 12345,
                "display": "y",
                "order": 1
        },
        "expDateCard": {
                "display": "y",
                "order": 2,
                "minExpDateMonths": 1,
                "maxExpDateMonths": 36
        },
        "cvvCard": {
                "display": "y",
                "order": 3
        },
        "atmPin": {
                "display": "y",
                "order": 4,
                "lengthDigits": 4
        },
        "ssn": {
                "display": "y",
                "order": 5
        },
        "dob": {
                "display": "y",
                "order": 6,
                "minAgeYears": 18,
                "maxAgeYears": 102
        },
        "fullName": {
                "display": "y",
                "order": 7
        },
        "firstName": {
                "display": "y",
                "order": 8
        },
        "middleInitial": {
                "display": "y",
                "order": 9
        },
        "lastName": {
                "display": "y",
                "order": 10
        },
		"userName": {
                "display": "y",
				"firstname.lastname": "y",
                "firstinitial.lastname": "n",
                "firstname.lastinitial": "n",
                "delimeter": ".",
                "order": 11
        },
        "email": {
                "display": "y",
                "order": 12
        },
        "homePhone": {
                "display": "y",
                "order": 13
        },
        "cellPhone": {
                "display": "y",
                "order": 14
        },
        "workPhone": {
                "display": "y",
                "order": 15
        },
        "mothersMaidenName": {
                "display": "y",
                "order": 16
        },
		"password": {
                "display": "y",
                "order": 17,
				"length": 9
        }
}

*/

// Configuration file struct
type structConfig struct {
	OutputQuantity     int                    `json:"outputQuantity"`
	OriginalFileOutput string                 `json:"originalFileOutput"`
	FileOutput         string                 `json:"fileOutput"`
	CreditCard         structCreditCard       `json:"creditCard"`
	ExpDateCard        structExpDateCard      `json:"expDateCard"`
	CVVCard            structDataCommon       `json:"cvvCard"`
	ATMPin             structDataCommonLength `json:"atmPin"`
	SSN                structDataCommon       `json:"ssn"`
	DOB                structDOB              `json:"dob"`
	Fullname           structDataCommon       `json:"fullName"`
	FirstName          structDataCommon       `json:"firstName"`
	MiddleInitial      structDataCommon       `json:"middleInitial"`
	LastName           structDataCommon       `json:"lastName"`
	Email              structDataCommon       `json:"Email"`
	UserName           structUsername         `json:"userName"`
	HomePhone          structDataCommon       `json:"homePhone"`
	CellPhone          structDataCommon       `json:"cellPhone"`
	WorkPhone          structDataCommon       `json:"workPhone"`
	MothersMaidenName  structDataCommon       `json:"mothersMaidenName"`
	Password           structDataCommonLength `json:"password"`
}

type structCreditCard struct {
	VisaBinNumber int    `json:"visaBinNumber"`
	Display       string `json:"display"`
	Order         int    `json:"order"`
}

type structExpDateCard struct {
	Display          string `json:"display"`
	Order            int    `json:"order"`
	MinExpDateMonths int    `json:"minExpDateMonths"`
	MaxExpDateMonths int    `json:"maxExpDateMonths"`
}

type structDataCommon struct {
	Display string `json:"display"`
	Order   int    `json:"order"`
}

type structUsername struct {
	Display              string `json:"display"`
	FirstInitialLastName string `json:"firstinitial.lastname"`
	FirstNameLastInitial string `json:"firstname.lastinitial"`
	FirstNameLastName    string `json:"firstname.lastname"`
	Delimeter            string `json:"delimeter"`
	Order                int    `json:"order"`
}

type structDataCommonLength struct {
	Display string `json:"display"`
	Order   int    `json:"order"`
	Length  int    `json:"length"`
}

type structDOB struct {
	Display     string `json:"display"`
	Order       int    `json:"order"`
	MinAgeYears int    `json:"minAgeYears"`
	MaxAgeYears int    `json:"maxAgeYears"`
}

func SaveOutputFile(message string, fileName string) {
	outFile, _ := os.Create(fileName)
	//CheckError("Unable to create txt file", err, true)
	defer outFile.Close()
	w := bufio.NewWriter(outFile)
	n, err := w.WriteString(message)
	if n < 1 {
		CheckError("Unable to write to txt file", err, true)
	}
	outFile.Sync()
	w.Flush()
	outFile.Close()
}

// CheckError checks for errors
func CheckError(reasonString string, err error, exitBool bool) {
	if err != nil && exitBool == true {
		fmt.Printf("%s\n", reasonString)
		//fmt.Printf("%s\n\n", err)
		os.Exit(0)
	} else if err != nil && exitBool == false {
		fmt.Printf("%s\n", reasonString)
		//fmt.Printf("%s\n", err)
		return
	}
}

func reverseIntSlice(slice []int) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func luhnCheck(number string) (bool, int, int) {
	// This takes a number that does not have the check digit and calculates it
	//fmt.Println(number)
	digits := []int{}
	for _, digitStr := range number {
		digit, err := strconv.Atoi(string(digitStr))
		if err != nil {
			return false, -1, -1 // Invalid input
		}
		digits = append(digits, digit)
	}
	//fmt.Println(digits)
	reverseIntSlice(digits)
	//fmt.Println(digits)
	sumDigits := 0
	for i := 0; i < len(digits); i++ {
		if i%2 == 0 {
			sumDigits += digits[i]
			//fmt.Printf("Original: %d - Calculated: %d - Sum: %d\n", digits[i], digits[i], sumDigits)
		} else {
			if digits[i]*2 < 10 {
				sumDigits += digits[i] * 2
				//fmt.Printf("Original: %d - Calculated: %d - Sum: %d\n", digits[i], digits[i]*2, sumDigits)
			} else {
				sumDigits += (digits[i] * 2) - 9
				//fmt.Printf("Original: %d - Calculated: %d - Sum: %d\n", digits[i], (digits[i]*2)-9, sumDigits)
			}
		}
	}

	checkDigit := 10 - (sumDigits % 10)
	verification := sumDigits % 10
	//fmt.Printf("Checkdigit: %d\n", checkDigit)
	return true, checkDigit, verification
}

func generateFakeCCNumbers(binNumber int, brand string, quantity int) []int64 {
	// Visa Card
	// Begins with the number 4
	// Numbers 2 through 6 is the BIN
	// 16 digits long

	fakeCCSlice := []int64{}
	for len(fakeCCSlice) < quantity {
		// First number is a 4 ----------------------------------------------------------------------------------
		initialDigit := 4
		// Generate a random bin number if bin number is 0 ------------------------------------------------------
		var fakeBinNumber int
		if binNumber == 0 {
			fakeBinNumber = rand.Intn(99999)
		} else if binNumber < 100000 && binNumber >= 0 {
			fakeBinNumber = binNumber
		} else {
			fmt.Println("Unable to process the bin number provided...")
			os.Exit(0)
		}
		//formattedFakeBinNumber := fmt.Sprintf("%05d", fakeBinNumber)
		//fmt.Println(formattedFakeBinNumber)
		// Generate a random card number for the next 9 digits --------------------------------------------------
		fakeNumber := rand.Intn(1000000000)
		//formattedFakeNumber := fmt.Sprintf("%09d", fakeNumber)
		//fmt.Println(formattedFakeNumber)
		// Last digit is the Luhn Algorithm Checksum Digit
		fakeCCNumber_without_Checkdigit := (initialDigit * 100000000000000) + (fakeBinNumber * 1000000000) + fakeNumber
		//fmt.Println(fakeCCNumber_without_Checksum)
		validLuhn, checkDigit, _ := luhnCheck(strconv.FormatInt(int64(fakeCCNumber_without_Checkdigit), 10))
		//fmt.Println(luhnNumber)

		if validLuhn == true {
			fakeCCNumber := (fakeCCNumber_without_Checkdigit * 10) + checkDigit
			//fmt.Printf("Validating Fake Generated CC: %d\n", fakeCCNumber)
			validLuhnVerify, _, verificationVerify := luhnCheck(strconv.FormatInt(int64(fakeCCNumber), 10))
			if validLuhnVerify == true && verificationVerify == 0 {
				//fmt.Printf("Valid Fake Generated CC: %d\n\n", fakeCCNumber)
				fakeCCSlice = append(fakeCCSlice, int64(fakeCCNumber))
			}
		}

	}
	return fakeCCSlice
}

func generateFakeSSN(quantity int) []string {
	fakeSSNSlice := []string{}
	for len(fakeSSNSlice) < quantity {
		fakeSSNNumber := rand.Intn(999999999)
		fakeSSNSlice = append(fakeSSNSlice, fmt.Sprintf("%09d", fakeSSNNumber))
	}
	return fakeSSNSlice
}

func generateFakePhone(quantity int) []string {
	fakePhoneSlice := []string{}
	for len(fakePhoneSlice) < quantity {
		fakePhoneNumber := rand.Intn(9999999999)
		fakePhoneSlice = append(fakePhoneSlice, fmt.Sprintf("%010d", fakePhoneNumber))
	}
	return fakePhoneSlice
}

func generateFakeCVV(quantity int) []string {
	fakeCVVSlice := []string{}
	for len(fakeCVVSlice) < quantity {
		fakeCVVNumber := rand.Intn(999)
		fakeCVVSlice = append(fakeCVVSlice, fmt.Sprintf("%03d", fakeCVVNumber))
	}
	return fakeCVVSlice
}

func generateFakeATMPin(quantity int) []string {
	fakePinSlice := []string{}
	for len(fakePinSlice) < quantity {
		fakePinNumber := rand.Intn(9999)
		fakePinSlice = append(fakePinSlice, fmt.Sprintf("%04d", fakePinNumber))
	}
	return fakePinSlice
}

func generateFakeExpirationDate(quantity int, minMonths int, maxMonths int) []string {
	fakeExpirationDateSlice := []string{}
	for len(fakeExpirationDateSlice) < quantity {
		currentDate := time.Now()
		// Create a fake Expiration date between 1 month and 3 years
		minExpDate := currentDate.AddDate(0, minMonths, 0)
		maxExpDate := currentDate.AddDate(0, maxMonths, 0)
		hoursRange := int(maxExpDate.Sub(minExpDate).Hours() / 24)
		randomHour := rand.Intn(hoursRange)
		randomExpDate := currentDate.AddDate(0, 0, (randomHour))
		fakeExpirationDateSlice = append(fakeExpirationDateSlice, fmt.Sprintf("%s", randomExpDate.Format("01/2006")))
	}
	return fakeExpirationDateSlice
}

func generateFakeDOB(minAge int, maxAge int, quantity int) []string {
	fakeDOBSlice := []string{}
	for len(fakeDOBSlice) < quantity {
		// Calculate the minimum and maximum birth dates
		//fmt.Println(minAge)
		currentDate := time.Now()
		minBirthDate := currentDate.AddDate((maxAge * -1), 0, 0)
		maxBirthDate := currentDate.AddDate((minAge * -1), 0, 0)
		// Generate a random number based on the number of hours between the minAge and maxAge
		hoursRange := int(maxBirthDate.Sub(minBirthDate).Hours() / 24)
		//fmt.Println(hoursRange)
		randomHour := rand.Intn(hoursRange)
		randomDOB := currentDate.AddDate(0, 0, (randomHour * -1))
		//fakeDOBSlice = append(fakeDOBSlice, fmt.Sprintf("%s", randomDOB.Format("2006-01-02")))
		fakeDOBSlice = append(fakeDOBSlice, fmt.Sprintf("%s", randomDOB.Format("01/02/2006")))
	}
	return fakeDOBSlice
}

func createFirstNameSlice() []string {
	var firstNameSlice []string
	firstNameSlice = append(firstNameSlice, "James", "John", "Robert", "Michael", "William", "David", "Richard", "Charles", "Joseph", "Thomas", "Christopher", "Daniel", "Paul", "Mark", "Donald")
	firstNameSlice = append(firstNameSlice, "George", "Kenneth", "Steven", "Edward", "Brian", "Ronald", "Anthony", "Kevin", "Jason", "Matthew", "Gary", "Timothy", "Jose", "Larry")
	firstNameSlice = append(firstNameSlice, "Jeffrey", "Frank", "Scott", "Eric", "Stephen", "Andrew", "Raymond", "Gregory", "Joshua", "Jerry", "Dennis", "Walter", "Patrick", "Peter")
	firstNameSlice = append(firstNameSlice, "Harold", "Douglas", "Henry", "Carl", "Arthur", "Ryan", "Roger", "Joe", "Juan", "Jack", "Albert", "Jonathan", "Justin", "Terry")
	firstNameSlice = append(firstNameSlice, "Gerald", "Keith", "Samuel", "Willie", "Ralph", "Lawrence", "Nicholas", "Roy", "Benjamin", "Bruce", "Brandon", "Adam", "Harry", "Fred")
	firstNameSlice = append(firstNameSlice, "Wayne", "Billy", "Steve", "Louis", "Jeremy", "Aaron", "Randy", "Howard", "Eugene", "Carlos", "Russell", "Bobby", "Victor", "Martin")
	firstNameSlice = append(firstNameSlice, "Ernest", "Phillip", "Todd", "Jesse", "Craig", "Alan", "Shawn", "Clarence", "Sean", "Philip", "Chris", "Johnny", "Earl", "Jimmy")
	firstNameSlice = append(firstNameSlice, "Antonio", "Danny", "Bryan", "Tony", "Luis", "Mike", "Stanley", "Leonard", "Nathan", "Dale", "Manuel", "Rodney", "Curtis", "Norman")
	firstNameSlice = append(firstNameSlice, "Allen", "Marvin", "Vincent", "Glenn", "Jeffery", "Travis", "Jeff", "Chad", "Jacob", "Lee", "Melvin", "Alfred", "Kyle", "Francis")
	firstNameSlice = append(firstNameSlice, "Bradley", "Jesus", "Herbert", "Frederick", "Ray", "Joel", "Edwin", "Don", "Eddie", "Ricky", "Troy", "Randall", "Barry", "Alexander")
	firstNameSlice = append(firstNameSlice, "Bernard", "Mario", "Leroy", "Francisco", "Marcus", "Micheal", "Theodore", "Clifford", "Miguel", "Oscar", "Jay", "Jim", "Tom", "Calvin")
	firstNameSlice = append(firstNameSlice, "Alex", "Jon", "Ronnie", "Bill", "Lloyd", "Tommy", "Leon", "Derek", "Warren", "Darrell", "Jerome", "Floyd", "Leo", "Alvin")
	firstNameSlice = append(firstNameSlice, "Tim", "Wesley", "Gordon", "Dean", "Greg", "Jorge", "Dustin", "Pedro", "Derrick", "Dan", "Lewis", "Zachary", "Corey", "Herman")
	firstNameSlice = append(firstNameSlice, "Maurice", "Vernon", "Roberto", "Clyde", "Glen", "Hector", "Shane", "Ricardo", "Sam", "Rick", "Lester", "Brent", "Ramon", "Charlie")
	firstNameSlice = append(firstNameSlice, "Tyler", "Gilbert", "Gene", "Marc", "Reginald", "Ruben", "Brett", "Angel", "Nathaniel", "Rafael", "Leslie", "Edgar", "Milton", "Raul")
	firstNameSlice = append(firstNameSlice, "Ben", "Chester", "Cecil", "Duane", "Franklin", "Andre", "Elmer", "Brad", "Gabriel", "Ron", "Mitchell", "Roland", "Arnold", "Harvey")
	firstNameSlice = append(firstNameSlice, "Jared", "Adrian", "Karl", "Cory", "Claude", "Erik", "Darryl", "Jamie", "Neil", "Jessie", "Christian", "Javier", "Fernando", "Clinton")
	firstNameSlice = append(firstNameSlice, "Ted", "Mathew", "Tyrone", "Darren", "Lonnie", "Lance", "Cody", "Julio", "Kelly", "Kurt", "Allan", "Nelson", "Guy", "Clayton")
	firstNameSlice = append(firstNameSlice, "Hugh", "Max", "Dwayne", "Dwight", "Armando", "Felix", "Jimmie", "Everett", "Jordan", "Ian", "Wallace", "Ken", "Bob", "Jaime")
	firstNameSlice = append(firstNameSlice, "Casey", "Alfredo", "Alberto", "Dave", "Ivan", "Johnnie", "Sidney", "Byron", "Julian", "Isaac", "Morris", "Clifton", "Willard", "Daryl")
	firstNameSlice = append(firstNameSlice, "Ross", "Virgil", "Andy", "Marshall", "Salvador", "Perry", "Kirk", "Sergio", "Marion", "Tracy", "Seth", "Kent", "Terrance", "Rene")
	firstNameSlice = append(firstNameSlice, "Eduardo", "Terrence", "Enrique", "Freddie", "Wade", "Austin", "Stuart", "Fredrick", "Arturo", "Alejandro", "Jackie", "Joey", "Nick", "Luther")
	firstNameSlice = append(firstNameSlice, "Wendell", "Jeremiah", "Evan", "Julius", "Dana", "Donnie", "Otis", "Shannon", "Trevor", "Oliver", "Luke", "Homer", "Gerard", "Doug")
	firstNameSlice = append(firstNameSlice, "Kenny", "Hubert", "Angelo", "Shaun", "Lyle", "Matt", "Lynn", "Alfonso", "Orlando", "Rex", "Carlton", "Ernesto", "Cameron", "Neal")
	firstNameSlice = append(firstNameSlice, "Pablo", "Lorenzo", "Omar", "Wilbur", "Blake", "Grant", "Horace", "Roderick", "Kerry", "Abraham", "Willis", "Rickey", "Jean", "Ira")
	firstNameSlice = append(firstNameSlice, "Andres", "Cesar", "Johnathan", "Malcolm", "Rudolph", "Damon", "Kelvin", "Rudy", "Preston", "Alton", "Archie", "Marco", "Wm", "Pete")
	firstNameSlice = append(firstNameSlice, "Randolph", "Garry", "Geoffrey", "Jonathon", "Felipe", "Bennie", "Gerardo", "Ed", "Dominic", "Robin", "Loren", "Delbert", "Colin", "Guillermo")
	firstNameSlice = append(firstNameSlice, "Earnest", "Lucas", "Benny", "Noel", "Spencer", "Rodolfo", "Myron", "Edmund", "Garrett", "Salvatore", "Cedric", "Lowell", "Gregg", "Sherman")
	firstNameSlice = append(firstNameSlice, "Wilson", "Devin", "Sylvester", "Kim", "Roosevelt", "Israel", "Jermaine", "Forrest", "Wilbert", "Leland", "Simon", "Guadalupe", "Clark", "Irving")
	firstNameSlice = append(firstNameSlice, "Carroll", "Bryant", "Owen", "Rufus", "Woodrow", "Sammy", "Kristopher", "Mack", "Levi", "Marcos", "Gustavo", "Jake", "Lionel", "Marty")
	firstNameSlice = append(firstNameSlice, "Taylor", "Ellis", "Dallas", "Gilberto", "Clint", "Nicolas", "Laurence", "Ismael", "Orville", "Drew", "Jody", "Ervin", "Dewey", "Al")
	firstNameSlice = append(firstNameSlice, "Wilfred", "Josh", "Hugo", "Ignacio", "Caleb", "Tomas", "Sheldon", "Erick", "Frankie", "Stewart", "Doyle", "Darrel", "Rogelio", "Terence")
	firstNameSlice = append(firstNameSlice, "Santiago", "Alonzo", "Elias", "Bert", "Elbert", "Ramiro", "Conrad", "Pat", "Noah", "Grady", "Phil", "Cornelius", "Lamar", "Rolando")
	firstNameSlice = append(firstNameSlice, "Clay", "Percy", "Dexter", "Bradford", "Merle", "Darin", "Amos", "Terrell", "Moses", "Irvin", "Saul", "Roman", "Darnell", "Randal")
	firstNameSlice = append(firstNameSlice, "Tommie", "Timmy", "Darrin", "Winston", "Brendan", "Toby", "Van", "Abel", "Dominick", "Boyd", "Courtney", "Jan", "Emilio", "Elijah")
	firstNameSlice = append(firstNameSlice, "Cary", "Domingo", "Santos", "Aubrey", "Emmett", "Marlon", "Emanuel", "Jerald", "Edmond", "Emil", "Dewayne", "Will", "Otto", "Teddy")
	firstNameSlice = append(firstNameSlice, "Reynaldo", "Bret", "Morgan", "Jess", "Trent", "Humberto", "Emmanuel", "Stephan", "Louie", "Vicente", "Lamont", "Stacy", "Garland", "Miles")
	firstNameSlice = append(firstNameSlice, "Micah", "Efrain", "Billie", "Logan", "Heath", "Rodger", "Harley", "Demetrius", "Ethan", "Eldon", "Rocky", "Pierre", "Junior", "Freddy")
	firstNameSlice = append(firstNameSlice, "Eli", "Bryce", "Antoine", "Robbie", "Kendall", "Royce", "Sterling", "Mickey", "Chase", "Grover", "Elton", "Cleveland", "Dylan", "Chuck")
	firstNameSlice = append(firstNameSlice, "Damian", "Reuben", "Stan", "August", "Leonardo", "Jasper", "Russel", "Erwin", "Benito", "Hans", "Monte", "Blaine", "Ernie", "Curt")
	firstNameSlice = append(firstNameSlice, "Quentin", "Agustin", "Murray", "Jamal", "Devon", "Adolfo", "Harrison", "Tyson", "Burton", "Brady", "Elliott", "Wilfredo", "Bart", "Jarrod")
	firstNameSlice = append(firstNameSlice, "Vance", "Denis", "Damien", "Joaquin", "Harlan", "Desmond", "Elliot", "Darwin", "Ashley", "Gregorio", "Buddy", "Xavier", "Kermit", "Roscoe")
	firstNameSlice = append(firstNameSlice, "Esteban", "Anton", "Solomon", "Scotty", "Norbert", "Elvin", "Williams", "Nolan", "Carey", "Rod", "Quinton", "Hal", "Brain", "Rob")
	firstNameSlice = append(firstNameSlice, "Elwood", "Kendrick", "Darius", "Moises", "Son", "Marlin", "Fidel", "Thaddeus", "Cliff", "Marcel", "Ali", "Jackson", "Raphael", "Bryon")
	firstNameSlice = append(firstNameSlice, "Armand", "Alvaro", "Jeffry", "Dane", "Joesph", "Thurman", "Ned", "Sammie", "Rusty", "Michel", "Monty", "Rory", "Fabian", "Reggie")
	firstNameSlice = append(firstNameSlice, "Mason", "Graham", "Kris", "Isaiah", "Vaughn", "Gus", "Avery", "Loyd", "Diego", "Alexis", "Adolph", "Norris", "Millard", "Rocco")
	firstNameSlice = append(firstNameSlice, "Gonzalo", "Derick", "Rodrigo", "Gerry", "Stacey", "Carmen", "Wiley", "Rigoberto", "Alphonso", "Ty", "Shelby", "Rickie", "Noe", "Vern")
	firstNameSlice = append(firstNameSlice, "Bobbie", "Reed", "Jefferson", "Elvis", "Bernardo", "Mauricio", "Hiram", "Donovan", "Basil", "Riley", "Ollie", "Nickolas", "Maynard", "Scot")
	firstNameSlice = append(firstNameSlice, "Vince", "Quincy", "Eddy", "Sebastian", "Federico", "Ulysses", "Heriberto", "Donnell", "Cole", "Denny", "Davis", "Gavin", "Emery", "Ward")
	firstNameSlice = append(firstNameSlice, "Romeo", "Jayson", "Dion", "Dante", "Clement", "Coy", "Odell", "Maxwell", "Jarvis", "Bruno", "Issac", "Mary", "Dudley", "Brock")
	firstNameSlice = append(firstNameSlice, "Sanford", "Colby", "Carmelo", "Barney", "Nestor", "Hollis", "Stefan", "Donny", "Art", "Linwood", "Beau", "Weldon", "Galen", "Isidro")
	firstNameSlice = append(firstNameSlice, "Truman", "Delmar", "Johnathon", "Silas", "Frederic", "Dick", "Kirby", "Irwin", "Cruz", "Merlin", "Merrill", "Charley", "Marcelino", "Lane")
	firstNameSlice = append(firstNameSlice, "Harris", "Cleo", "Carlo", "Trenton", "Kurtis", "Hunter", "Aurelio", "Winfred", "Vito", "Collin", "Denver", "Carter", "Leonel", "Emory")
	firstNameSlice = append(firstNameSlice, "Pasquale", "Mohammad", "Mariano", "Danial", "Blair", "Landon", "Dirk", "Branden", "Adan", "Numbers", "Clair", "Buford", "German", "Bernie")
	firstNameSlice = append(firstNameSlice, "Wilmer", "Joan", "Emerson", "Zachery", "Fletcher", "Jacques", "Errol", "Dalton", "Monroe", "Josue", "Dominique", "Edwardo", "Booker", "Wilford")
	firstNameSlice = append(firstNameSlice, "Sonny", "Shelton", "Carson", "Theron", "Raymundo", "Daren", "Tristan", "Houston", "Robby", "Lincoln", "Jame", "Genaro", "Gale", "Bennett")
	firstNameSlice = append(firstNameSlice, "Octavio", "Cornell", "Laverne", "Hung", "Arron", "Antony", "Herschel", "Alva", "Giovanni", "Garth", "Cyrus", "Cyril", "Ronny", "Stevie")
	firstNameSlice = append(firstNameSlice, "Lon", "Freeman", "Erin", "Duncan", "Kennith", "Carmine", "Augustine", "Young", "Erich", "Chadwick", "Wilburn", "Russ", "Reid", "Myles")
	firstNameSlice = append(firstNameSlice, "Anderson", "Morton", "Jonas", "Forest", "Mitchel", "Mervin", "Zane", "Rich", "Jamel", "Lazaro", "Alphonse", "Randell", "Major", "Johnie")
	firstNameSlice = append(firstNameSlice, "Jarrett", "Brooks", "Ariel", "Abdul", "Dusty", "Luciano", "Lindsey", "Tracey", "Seymour", "Scottie", "Eugenio", "Mohammed", "Sandy", "Valentin")
	firstNameSlice = append(firstNameSlice, "Chance", "Arnulfo", "Lucien", "Ferdinand", "Thad", "Ezra", "Sydney", "Aldo", "Rubin", "Royal", "Mitch", "Earle", "Abe", "Wyatt")
	firstNameSlice = append(firstNameSlice, "Marquis", "Lanny", "Kareem", "Jamar", "Boris", "Isiah", "Emile", "Elmo", "Aron", "Leopoldo", "Everette", "Josef", "Gail", "Eloy")
	firstNameSlice = append(firstNameSlice, "Dorian", "Rodrick", "Reinaldo", "Lucio", "Jerrod", "Weston", "Hershel", "Barton", "Parker", "Lemuel", "Lavern", "Burt", "Jules", "Gil")
	firstNameSlice = append(firstNameSlice, "Eliseo", "Ahmad", "Nigel", "Efren", "Antwan", "Alden", "Margarito", "Coleman", "Refugio", "Dino", "Osvaldo", "Les", "Deandre", "Normand")
	firstNameSlice = append(firstNameSlice, "Kieth", "Ivory", "Andrea", "Trey", "Norberto", "Napoleon", "Jerold", "Fritz", "Rosendo", "Milford", "Sang", "Deon", "Christoper", "Alfonzo")
	firstNameSlice = append(firstNameSlice, "Lyman", "Josiah", "Brant", "Wilton", "Rico", "Jamaal", "Dewitt", "Carol", "Brenton", "Yong", "Olin", "Foster", "Faustino", "Claudio")
	firstNameSlice = append(firstNameSlice, "Judson", "Gino", "Edgardo", "Berry", "Alec", "Tanner", "Jarred", "Donn", "Trinidad", "Tad", "Shirley", "Prince", "Porfirio", "Odis")
	firstNameSlice = append(firstNameSlice, "Maria", "Lenard", "Chauncey", "Chang", "Tod", "Mel", "Marcelo", "Kory", "Augustus", "Keven", "Hilario", "Bud", "Sal", "Rosario")
	firstNameSlice = append(firstNameSlice, "Orval", "Mauro", "Dannie", "Zachariah", "Olen", "Anibal", "Milo", "Jed", "Frances", "Thanh", "Dillon", "Amado", "Newton", "Connie")
	firstNameSlice = append(firstNameSlice, "Lenny", "Tory", "Richie", "Lupe", "Horacio", "Brice", "Mohamed", "Delmer", "Dario", "Reyes", "Dee", "Mac", "Jonah", "Jerrold")
	firstNameSlice = append(firstNameSlice, "Robt", "Hank", "Sung", "Rupert", "Rolland", "Kenton", "Damion", "Chi", "Antone", "Waldo", "Fredric", "Bradly", "Quinn", "Kip")
	firstNameSlice = append(firstNameSlice, "Mary", "Patricia", "Linda", "Barbara", "Elizabeth", "Jennifer", "Maria", "Susan", "Margaret", "Dorothy", "Lisa", "Nancy", "Karen", "Betty", "Helen")
	firstNameSlice = append(firstNameSlice, "Sandra", "Donna", "Carol", "Ruth", "Sharon", "Michelle", "Laura", "Sarah", "Kimberly", "Deborah", "Jessica", "Shirley", "Cynthia", "Angela")
	firstNameSlice = append(firstNameSlice, "Melissa", "Brenda", "Amy", "Anna", "Rebecca", "Virginia", "Kathleen", "Pamela", "Martha", "Debra", "Amanda", "Stephanie", "Carolyn", "Christine")
	firstNameSlice = append(firstNameSlice, "Marie", "Janet", "Catherine", "Frances", "Ann", "Joyce", "Diane", "Alice", "Julie", "Heather", "Teresa", "Doris", "Gloria", "Evelyn")
	firstNameSlice = append(firstNameSlice, "Jean", "Cheryl", "Mildred", "Katherine", "Joan", "Ashley", "Judith", "Rose", "Janice", "Kelly", "Nicole", "Judy", "Christina", "Kathy")
	firstNameSlice = append(firstNameSlice, "Theresa", "Beverly", "Denise", "Tammy", "Irene", "Jane", "Lori", "Rachel", "Marilyn", "Andrea", "Kathryn", "Louise", "Sara", "Anne")
	firstNameSlice = append(firstNameSlice, "Jacqueline", "Wanda", "Bonnie", "Julia", "Ruby", "Lois", "Tina", "Phyllis", "Norma", "Paula", "Diana", "Annie", "Lillian", "Emily")
	firstNameSlice = append(firstNameSlice, "Robin", "Peggy", "Crystal", "Gladys", "Rita", "Dawn", "Connie", "Florence", "Tracy", "Edna", "Tiffany", "Carmen", "Rosa", "Cindy")
	firstNameSlice = append(firstNameSlice, "Grace", "Wendy", "Victoria", "Edith", "Kim", "Sherry", "Sylvia", "Josephine", "Thelma", "Shannon", "Sheila", "Ethel", "Ellen", "Elaine")
	firstNameSlice = append(firstNameSlice, "Marjorie", "Carrie", "Charlotte", "Monica", "Esther", "Pauline", "Emma", "Juanita", "Anita", "Rhonda", "Hazel", "Amber", "Eva", "Debbie")
	firstNameSlice = append(firstNameSlice, "April", "Leslie", "Clara", "Lucille", "Jamie", "Joanne", "Eleanor", "Valerie", "Danielle", "Megan", "Alicia", "Suzanne", "Michele", "Gail")
	firstNameSlice = append(firstNameSlice, "Bertha", "Darlene", "Veronica", "Jill", "Erin", "Geraldine", "Lauren", "Cathy", "Joann", "Lorraine", "Lynn", "Sally", "Regina", "Erica")
	firstNameSlice = append(firstNameSlice, "Beatrice", "Dolores", "Bernice", "Audrey", "Yvonne", "Annette", "June", "Samantha", "Marion", "Dana", "Stacy", "Ana", "Renee", "Ida")
	firstNameSlice = append(firstNameSlice, "Vivian", "Roberta", "Holly", "Brittany", "Melanie", "Loretta", "Yolanda", "Jeanette", "Laurie", "Katie", "Kristen", "Vanessa", "Alma", "Sue")
	firstNameSlice = append(firstNameSlice, "Elsie", "Beth", "Jeanne", "Vicki", "Carla", "Tara", "Rosemary", "Eileen", "Terri", "Gertrude", "Lucy", "Tonya", "Ella", "Stacey")
	firstNameSlice = append(firstNameSlice, "Wilma", "Gina", "Kristin", "Jessie", "Natalie", "Agnes", "Vera", "Willie", "Charlene", "Bessie", "Delores", "Melinda", "Pearl", "Arlene")
	firstNameSlice = append(firstNameSlice, "Maureen", "Colleen", "Allison", "Tamara", "Joy", "Georgia", "Constance", "Lillie", "Claudia", "Jackie", "Marcia", "Tanya", "Nellie", "Minnie")
	firstNameSlice = append(firstNameSlice, "Marlene", "Heidi", "Glenda", "Lydia", "Viola", "Courtney", "Marian", "Stella", "Caroline", "Dora", "Jo", "Vickie", "Mattie", "Terry")
	firstNameSlice = append(firstNameSlice, "Maxine", "Irma", "Mabel", "Marsha", "Myrtle", "Lena", "Christy", "Deanna", "Patsy", "Hilda", "Gwendolyn", "Jennie", "Nora", "Margie")
	firstNameSlice = append(firstNameSlice, "Nina", "Cassandra", "Leah", "Penny", "Kay", "Priscilla", "Naomi", "Carole", "Brandy", "Olga", "Billie", "Dianne", "Tracey", "Leona")
	firstNameSlice = append(firstNameSlice, "Jenny", "Felicia", "Sonia", "Miriam", "Velma", "Becky", "Bobbie", "Violet", "Kristina", "Toni", "Misty", "Mae", "Shelly", "Daisy")
	firstNameSlice = append(firstNameSlice, "Ramona", "Sherri", "Erika", "Katrina", "Claire", "Lindsey", "Lindsay", "Geneva", "Guadalupe", "Belinda", "Margarita", "Sheryl", "Cora", "Faye")
	firstNameSlice = append(firstNameSlice, "Ada", "Natasha", "Sabrina", "Isabel", "Marguerite", "Hattie", "Harriet", "Molly", "Cecilia", "Kristi", "Brandi", "Blanche", "Sandy", "Rosie")
	firstNameSlice = append(firstNameSlice, "Joanna", "Iris", "Eunice", "Angie", "Inez", "Lynda", "Madeline", "Amelia", "Alberta", "Genevieve", "Monique", "Jodi", "Janie", "Maggie")
	firstNameSlice = append(firstNameSlice, "Kayla", "Sonya", "Jan", "Lee", "Kristine", "Candace", "Fannie", "Maryann", "Opal", "Alison", "Yvette", "Melody", "Luz", "Susie")
	firstNameSlice = append(firstNameSlice, "Olivia", "Flora", "Shelley", "Kristy", "Mamie", "Lula", "Lola", "Verna", "Beulah", "Antoinette", "Candice", "Juana", "Jeannette", "Pam")
	firstNameSlice = append(firstNameSlice, "Kelli", "Hannah", "Whitney", "Bridget", "Karla", "Celia", "Latoya", "Patty", "Shelia", "Gayle", "Della", "Vicky", "Lynne", "Sheri")
	firstNameSlice = append(firstNameSlice, "Marianne", "Kara", "Jacquelyn", "Erma", "Blanca", "Myra", "Leticia", "Pat", "Krista", "Roxanne", "Angelica", "Johnnie", "Robyn", "Francis")
	firstNameSlice = append(firstNameSlice, "Adrienne", "Rosalie", "Alexandra", "Brooke", "Bethany", "Sadie", "Bernadette", "Traci", "Jody", "Kendra", "Jasmine", "Nichole", "Rachael", "Chelsea")
	firstNameSlice = append(firstNameSlice, "Mable", "Ernestine", "Muriel", "Marcella", "Elena", "Krystal", "Angelina", "Nadine", "Kari", "Estelle", "Dianna", "Paulette", "Lora", "Mona")
	firstNameSlice = append(firstNameSlice, "Doreen", "Rosemarie", "Angel", "Desiree", "Antonia", "Hope", "Ginger", "Janis", "Betsy", "Christie", "Freda", "Mercedes", "Meredith", "Lynette")
	firstNameSlice = append(firstNameSlice, "Teri", "Cristina", "Eula", "Leigh", "Meghan", "Sophia", "Eloise", "Rochelle", "Gretchen", "Cecelia", "Raquel", "Henrietta", "Alyssa", "Jana")
	firstNameSlice = append(firstNameSlice, "Kelley", "Gwen", "Kerry", "Jenna", "Tricia", "Laverne", "Olive", "Alexis", "Tasha", "Silvia", "Elvira", "Casey", "Delia", "Sophie")
	firstNameSlice = append(firstNameSlice, "Kate", "Patti", "Lorena", "Kellie", "Sonja", "Lila", "Lana", "Darla", "May", "Mindy", "Essie", "Mandy", "Lorene", "Elsa")
	firstNameSlice = append(firstNameSlice, "Josefina", "Jeannie", "Miranda", "Dixie", "Lucia", "Marta", "Faith", "Lela", "Johanna", "Shari", "Camille", "Tami", "Shawna", "Elisa")
	firstNameSlice = append(firstNameSlice, "Ebony", "Melba", "Ora", "Nettie", "Tabitha", "Ollie", "Jaime", "Winifred", "Kristie", "Marina", "Alisha", "Aimee", "Rena", "Myrna")
	firstNameSlice = append(firstNameSlice, "Marla", "Tammie", "Latasha", "Bonita", "Patrice", "Ronda", "Sherrie", "Addie", "Francine", "Deloris", "Stacie", "Adriana", "Cheri", "Shelby")
	firstNameSlice = append(firstNameSlice, "Abigail", "Celeste", "Jewel", "Cara", "Adele", "Rebekah", "Lucinda", "Dorthy", "Chris", "Effie", "Trina", "Reba", "Shawn", "Sallie")
	firstNameSlice = append(firstNameSlice, "Aurora", "Lenora", "Etta", "Lottie", "Kerri", "Trisha", "Nikki", "Estella", "Francisca", "Josie", "Tracie", "Marissa", "Karin", "Brittney")
	firstNameSlice = append(firstNameSlice, "Janelle", "Lourdes", "Laurel", "Helene", "Fern", "Elva", "Corinne", "Kelsey", "Ina", "Bettie", "Elisabeth", "Aida", "Caitlin", "Ingrid")
	firstNameSlice = append(firstNameSlice, "Iva", "Eugenia", "Christa", "Goldie", "Cassie", "Maude", "Jenifer", "Therese", "Frankie", "Dena", "Lorna", "Janette", "Latonya", "Candy")
	firstNameSlice = append(firstNameSlice, "Morgan", "Consuelo", "Tamika", "Rosetta", "Debora", "Cherie", "Polly", "Dina", "Jewell", "Fay", "Jillian", "Dorothea", "Nell", "Trudy")
	firstNameSlice = append(firstNameSlice, "Esperanza", "Patrica", "Kimberley", "Shanna", "Helena", "Carolina", "Cleo", "Stefanie", "Rosario", "Ola", "Janine", "Mollie", "Lupe", "Alisa")
	firstNameSlice = append(firstNameSlice, "Lou", "Maribel", "Susanne", "Bette", "Susana", "Elise", "Cecile", "Isabelle", "Lesley", "Jocelyn", "Paige", "Joni", "Rachelle", "Leola")
	firstNameSlice = append(firstNameSlice, "Daphne", "Alta", "Ester", "Petra", "Graciela", "Imogene", "Jolene", "Keisha", "Lacey", "Glenna", "Gabriela", "Keri", "Ursula", "Lizzie")
	firstNameSlice = append(firstNameSlice, "Kirsten", "Shana", "Adeline", "Mayra", "Jayne", "Jaclyn", "Gracie", "Sondra", "Carmela", "Marisa", "Rosalind", "Charity", "Tonia", "Beatriz")
	firstNameSlice = append(firstNameSlice, "Marisol", "Clarice", "Jeanine", "Sheena", "Angeline", "Frieda", "Lily", "Robbie", "Shauna", "Millie", "Claudette", "Cathleen", "Angelia", "Gabrielle")
	firstNameSlice = append(firstNameSlice, "Autumn", "Katharine", "Summer", "Jodie", "Staci", "Lea", "Christi", "Jimmie", "Justine", "Elma", "Luella", "Margret", "Dominique", "Socorro")
	firstNameSlice = append(firstNameSlice, "Rene", "Martina", "Margo", "Mavis", "Callie", "Bobbi", "Maritza", "Lucile", "Leanne", "Jeannine", "Deana", "Aileen", "Lorie", "Ladonna")
	firstNameSlice = append(firstNameSlice, "Willa", "Manuela", "Gale", "Selma", "Dolly", "Sybil", "Abby", "Lara", "Dale", "Ivy", "Dee", "Winnie", "Marcy", "Luisa")
	firstNameSlice = append(firstNameSlice, "Jeri", "Magdalena", "Ofelia", "Meagan", "Audra", "Matilda", "Leila", "Cornelia", "Bianca", "Simone", "Bettye", "Randi", "Virgie", "Latisha")
	firstNameSlice = append(firstNameSlice, "Barbra", "Georgina", "Eliza", "Leann", "Bridgette", "Rhoda", "Haley", "Adela", "Nola", "Bernadine", "Flossie", "Ila", "Greta", "Ruthie")
	firstNameSlice = append(firstNameSlice, "Nelda", "Minerva", "Lilly", "Terrie", "Letha", "Hilary", "Estela", "Valarie", "Brianna", "Rosalyn", "Earline", "Catalina", "Ava", "Mia")
	firstNameSlice = append(firstNameSlice, "Clarissa", "Lidia", "Corrine", "Alexandria", "Concepcion", "Tia", "Sharron", "Rae", "Dona", "Ericka", "Jami", "Elnora", "Chandra", "Lenore")
	firstNameSlice = append(firstNameSlice, "Neva", "Marylou", "Melisa", "Tabatha", "Serena", "Avis", "Allie", "Sofia", "Jeanie", "Odessa", "Nannie", "Harriett", "Loraine", "Penelope")
	firstNameSlice = append(firstNameSlice, "Milagros", "Emilia", "Benita", "Allyson", "Ashlee", "Tania", "Tommie", "Esmeralda", "Karina", "Eve", "Pearlie", "Zelma", "Malinda", "Noreen")
	firstNameSlice = append(firstNameSlice, "Tameka", "Saundra", "Hillary", "Amie", "Althea", "Rosalinda", "Jordan", "Lilia", "Alana", "Gay", "Clare", "Alejandra", "Elinor", "Michael")
	firstNameSlice = append(firstNameSlice, "Lorrie", "Jerri", "Darcy", "Earnestine", "Carmella", "Taylor", "Noemi", "Marcie", "Liza", "Annabelle", "Louisa", "Earlene", "Mallory", "Carlene")
	firstNameSlice = append(firstNameSlice, "Nita", "Selena", "Tanisha", "Katy", "Julianne", "John", "Lakisha", "Edwina", "Maricela", "Margery", "Kenya", "Dollie", "Roxie", "Roslyn")
	firstNameSlice = append(firstNameSlice, "Kathrine", "Nanette", "Charmaine", "Lavonne", "Ilene", "Kris", "Tammi", "Suzette", "Corine", "Kaye", "Jerry", "Merle", "Chrystal", "Lina")
	firstNameSlice = append(firstNameSlice, "Deanne", "Lilian", "Juliana", "Aline", "Luann", "Kasey", "Maryanne", "Evangeline", "Colette", "Melva", "Lawanda", "Yesenia", "Nadia", "Madge")
	firstNameSlice = append(firstNameSlice, "Kathie", "Eddie", "Ophelia", "Valeria", "Nona", "Mitzi", "Mari", "Georgette", "Claudine", "Fran", "Alissa", "Roseann", "Lakeisha", "Susanna")
	firstNameSlice = append(firstNameSlice, "Reva", "Deidre", "Chasity", "Sheree", "Carly", "James", "Elvia", "Alyce", "Deirdre", "Gena", "Briana", "Araceli", "Katelyn", "Rosanne")
	firstNameSlice = append(firstNameSlice, "Wendi", "Tessa", "Berta", "Marva", "Imelda", "Marietta", "Marci", "Leonor", "Arline", "Sasha", "Madelyn", "Janna", "Juliette", "Deena")
	firstNameSlice = append(firstNameSlice, "Aurelia", "Josefa", "Augusta", "Liliana", "Young", "Christian", "Lessie", "Amalia", "Savannah", "Anastasia", "Vilma", "Natalia", "Rosella", "Lynnette")
	firstNameSlice = append(firstNameSlice, "Corina", "Alfreda", "Leanna", "Carey", "Amparo", "Coleen", "Tamra", "Aisha", "Wilda", "Karyn", "Cherry", "Queen", "Maura", "Mai")
	firstNameSlice = append(firstNameSlice, "Evangelina", "Rosanna", "Hallie", "Erna", "Enid", "Mariana", "Lacy", "Juliet", "Jacklyn", "Freida", "Madeleine", "Mara", "Hester", "Cathryn")
	firstNameSlice = append(firstNameSlice, "Lelia", "Casandra", "Bridgett", "Angelita", "Jannie", "Dionne", "Annmarie", "Katina", "Beryl", "Phoebe", "Millicent", "Katheryn", "Diann", "Carissa")
	firstNameSlice = append(firstNameSlice, "Maryellen", "Liz", "Lauri", "Helga", "Gilda", "Adrian", "Rhea", "Marquita", "Hollie", "Tisha", "Tamera", "Angelique", "Francesca", "Britney")
	firstNameSlice = append(firstNameSlice, "Kaitlin", "Lolita", "Florine", "Rowena", "Reyna", "Twila", "Fanny", "Janell", "Ines", "Concetta", "Bertie", "Alba", "Brigitte", "Alyson")
	firstNameSlice = append(firstNameSlice, "Vonda", "Pansy", "Elba", "Noelle", "Letitia", "Kitty", "Deann", "Brandie", "Louella", "Leta", "Felecia", "Sharlene", "Lesa", "Beverley")
	return firstNameSlice
}

func createSurnameSlice() []string {
	var surnameSlice []string
	surnameSlice = append(surnameSlice, "Smith", "Johnson", "Williams", "Jones", "Brown", "Davis", "Miller", "Wilson", "Moore", "Taylor", "Anderson", "Thomas", "Jackson", "White", "Harris")
	surnameSlice = append(surnameSlice, "Martin", "Thompson", "Garcia", "Martinez", "Robinson", "Clark", "Rodriguez", "Lewis", "Lee", "Walker", "Hall", "Allen", "Young", "Hernandez")
	surnameSlice = append(surnameSlice, "King", "Wright", "Lopez", "Hill", "Scott", "Green", "Adams", "Baker", "Gonzalez", "Nelson", "Carter", "Mitchell", "Perez", "Roberts")
	surnameSlice = append(surnameSlice, "Turner", "Phillips", "Campbell", "Parker", "Evans", "Edwards", "Collins", "Stewart", "Sanchez", "Morris", "Rogers", "Reed", "Cook", "Morgan")
	surnameSlice = append(surnameSlice, "Bell", "Murphy", "Bailey", "Rivera", "Cooper", "Richardson", "Cox", "Howard", "Ward", "Torres", "Peterson", "Gray", "Ramirez", "James")
	surnameSlice = append(surnameSlice, "Watson", "Brooks", "Kelly", "Sanders", "Price", "Bennett", "Wood", "Barnes", "Ross", "Henderson", "Coleman", "Jenkins", "Perry", "Powell")
	surnameSlice = append(surnameSlice, "Long", "Patterson", "Hughes", "Flores", "Washington", "Butler", "Simmons", "Foster", "Gonzales", "Bryant", "Alexander", "Russell", "Griffin", "Diaz")
	surnameSlice = append(surnameSlice, "Hayes", "Myers", "Ford", "Hamilton", "Graham", "Sullivan", "Wallace", "Woods", "Cole", "West", "Jordan", "Owens", "Reynolds", "Fisher")
	surnameSlice = append(surnameSlice, "Ellis", "Harrison", "Gibson", "Mcdonald", "Cruz", "Marshall", "Ortiz", "Gomez", "Murray", "Freeman", "Wells", "Webb", "Simpson", "Stevens")
	surnameSlice = append(surnameSlice, "Tucker", "Porter", "Hunter", "Hicks", "Crawford", "Henry", "Boyd", "Mason", "Morales", "Kennedy", "Warren", "Dixon", "Ramos", "Reyes")
	surnameSlice = append(surnameSlice, "Burns", "Gordon", "Shaw", "Holmes", "Rice", "Robertson", "Hunt", "Black", "Daniels", "Palmer", "Mills", "Nichols", "Grant", "Knight")
	surnameSlice = append(surnameSlice, "Ferguson", "Rose", "Stone", "Hawkins", "Dunn", "Perkins", "Hudson", "Spencer", "Gardner", "Stephens", "Payne", "Pierce", "Berry", "Matthews")
	surnameSlice = append(surnameSlice, "Arnold", "Wagner", "Willis", "Ray", "Watkins", "Olson", "Carroll", "Duncan", "Snyder", "Hart", "Cunningham", "Bradley", "Lane", "Andrews")
	surnameSlice = append(surnameSlice, "Ruiz", "Harper", "Fox", "Riley", "Armstrong", "Carpenter", "Weaver", "Greene", "Lawrence", "Elliott", "Chavez", "Sims", "Austin", "Peters")
	surnameSlice = append(surnameSlice, "Kelley", "Franklin", "Lawson", "Fields", "Gutierrez", "Ryan", "Schmidt", "Carr", "Vasquez", "Castillo", "Wheeler", "Chapman", "Oliver", "Montgomery")
	surnameSlice = append(surnameSlice, "Richards", "Williamson", "Johnston", "Banks", "Meyer", "Bishop", "Mccoy", "Howell", "Alvarez", "Morrison", "Hansen", "Fernandez", "Garza", "Harvey")
	surnameSlice = append(surnameSlice, "Little", "Burton", "Stanley", "Nguyen", "George", "Jacobs", "Reid", "Kim", "Fuller", "Lynch", "Dean", "Gilbert", "Garrett", "Romero")
	surnameSlice = append(surnameSlice, "Welch", "Larson", "Frazier", "Burke", "Hanson", "Day", "Mendoza", "Moreno", "Bowman", "Medina", "Fowler", "Brewer", "Hoffman", "Carlson")
	surnameSlice = append(surnameSlice, "Silva", "Pearson", "Holland", "Douglas", "Fleming", "Jensen", "Vargas", "Byrd", "Davidson", "Hopkins", "May", "Terry", "Herrera", "Wade")
	surnameSlice = append(surnameSlice, "Soto", "Walters", "Curtis", "Neal", "Caldwell", "Lowe", "Jennings", "Barnett", "Graves", "Jimenez", "Horton", "Shelton", "Barrett", "Obrien")
	surnameSlice = append(surnameSlice, "Castro", "Sutton", "Gregory", "Mckinney", "Lucas", "Miles", "Craig", "Rodriquez", "Chambers", "Holt", "Lambert", "Fletcher", "Watts", "Bates")
	surnameSlice = append(surnameSlice, "Hale", "Rhodes", "Pena", "Beck", "Newman", "Haynes", "Mcdaniel", "Mendez", "Bush", "Vaughn", "Parks", "Dawson", "Santiago", "Norris")
	surnameSlice = append(surnameSlice, "Hardy", "Love", "Steele", "Curry", "Powers", "Schultz", "Barker", "Guzman", "Page", "Munoz", "Ball", "Keller", "Chandler", "Weber")
	surnameSlice = append(surnameSlice, "Leonard", "Walsh", "Lyons", "Ramsey", "Wolfe", "Schneider", "Mullins", "Benson", "Sharp", "Bowen", "Daniel", "Barber", "Cummings", "Hines")
	surnameSlice = append(surnameSlice, "Baldwin", "Griffith", "Valdez", "Hubbard", "Salazar", "Reeves", "Warner", "Stevenson", "Burgess", "Santos", "Tate", "Cross", "Garner", "Mann")
	surnameSlice = append(surnameSlice, "Mack", "Moss", "Thornton", "Dennis", "Mcgee", "Farmer", "Delgado", "Aguilar", "Vega", "Glover", "Manning", "Cohen", "Harmon", "Rodgers")
	surnameSlice = append(surnameSlice, "Robbins", "Newton", "Todd", "Blair", "Higgins", "Ingram", "Reese", "Cannon", "Strickland", "Townsend", "Potter", "Goodwin", "Walton", "Rowe")
	surnameSlice = append(surnameSlice, "Hampton", "Ortega", "Patton", "Swanson", "Joseph", "Francis", "Goodman", "Maldonado", "Yates", "Becker", "Erickson", "Hodges", "Rios", "Conner")
	surnameSlice = append(surnameSlice, "Adkins", "Webster", "Norman", "Malone", "Hammond", "Flowers", "Cobb", "Moody", "Quinn", "Blake", "Maxwell", "Pope", "Floyd", "Osborne")
	surnameSlice = append(surnameSlice, "Paul", "Mccarthy", "Guerrero", "Lindsey", "Estrada", "Sandoval", "Gibbs", "Tyler", "Gross", "Fitzgerald", "Stokes", "Doyle", "Sherman", "Saunders")
	surnameSlice = append(surnameSlice, "Wise", "Colon", "Gill", "Alvarado", "Greer", "Padilla", "Simon", "Waters", "Nunez", "Ballard", "Schwartz", "Mcbride", "Houston", "Christensen")
	surnameSlice = append(surnameSlice, "Klein", "Pratt", "Briggs", "Parsons", "Mclaughlin", "Zimmerman", "French", "Buchanan", "Moran", "Copeland", "Roy", "Pittman", "Brady", "Mccormick")
	surnameSlice = append(surnameSlice, "Holloway", "Brock", "Poole", "Frank", "Logan", "Owen", "Bass", "Marsh", "Drake", "Wong", "Jefferson", "Park", "Morton", "Abbott")
	surnameSlice = append(surnameSlice, "Sparks", "Patrick", "Norton", "Huff", "Clayton", "Massey", "Lloyd", "Figueroa", "Carson", "Bowers", "Roberson", "Barton", "Tran", "Lamb")
	surnameSlice = append(surnameSlice, "Harrington", "Casey", "Boone", "Cortez", "Clarke", "Mathis", "Singleton", "Wilkins", "Cain", "Bryan", "Underwood", "Hogan", "Mckenzie", "Collier")
	surnameSlice = append(surnameSlice, "Luna", "Phelps", "Mcguire", "Allison", "Bridges", "Wilkerson", "Nash", "Summers", "Atkins", "Wilcox", "Pitts", "Conley", "Marquez", "Burnett")
	surnameSlice = append(surnameSlice, "Richard", "Cochran", "Chase", "Davenport", "Hood", "Gates", "Clay", "Ayala", "Sawyer", "Roman", "Vazquez", "Dickerson", "Hodge", "Acosta")
	surnameSlice = append(surnameSlice, "Flynn", "Espinoza", "Nicholson", "Monroe", "Wolf", "Morrow", "Kirk", "Randall", "Anthony", "Whitaker", "Oconnor", "Skinner", "Ware", "Molina")
	surnameSlice = append(surnameSlice, "Kirby", "Huffman", "Bradford", "Charles", "Gilmore", "Dominguez", "Oneal", "Bruce", "Lang", "Combs", "Kramer", "Heath", "Hancock", "Gallagher")
	surnameSlice = append(surnameSlice, "Gaines", "Shaffer", "Short", "Wiggins", "Mathews", "Mcclain", "Fischer", "Wall", "Small", "Melton", "Hensley", "Bond", "Dyer", "Cameron")
	surnameSlice = append(surnameSlice, "Grimes", "Contreras", "Christian", "Wyatt", "Baxter", "Snow", "Mosley", "Shepherd", "Larsen", "Hoover", "Beasley", "Glenn", "Petersen", "Whitehead")
	surnameSlice = append(surnameSlice, "Meyers", "Keith", "Garrison", "Vincent", "Shields", "Horn", "Savage", "Olsen", "Schroeder", "Hartman", "Woodard", "Mueller", "Kemp", "Deleon")
	surnameSlice = append(surnameSlice, "Booth", "Patel", "Calhoun", "Wiley", "Eaton", "Cline", "Navarro", "Harrell", "Lester", "Humphrey", "Parrish", "Duran", "Hutchinson", "Hess")
	surnameSlice = append(surnameSlice, "Dorsey", "Bullock", "Robles", "Beard", "Dalton", "Avila", "Vance", "Rich", "Blackwell", "York", "Johns", "Blankenship", "Trevino", "Salinas")
	surnameSlice = append(surnameSlice, "Campos", "Pruitt", "Moses", "Callahan", "Golden", "Montoya", "Hardin", "Guerra", "Mcdowell", "Carey", "Stafford", "Gallegos", "Henson", "Wilkinson")
	surnameSlice = append(surnameSlice, "Booker", "Merritt", "Miranda", "Atkinson", "Orr", "Decker", "Hobbs", "Preston", "Tanner", "Knox", "Pacheco", "Stephenson", "Glass", "Rojas")
	surnameSlice = append(surnameSlice, "Serrano", "Marks", "Hickman", "English", "Sweeney", "Strong", "Prince", "Mcclure", "Conway", "Walter", "Roth", "Maynard", "Farrell", "Lowery")
	surnameSlice = append(surnameSlice, "Hurst", "Nixon", "Weiss", "Trujillo", "Ellison", "Sloan", "Juarez", "Winters", "Mclean", "Randolph", "Leon", "Boyer", "Villarreal", "Mccall")
	surnameSlice = append(surnameSlice, "Gentry", "Carrillo", "Kent", "Ayers", "Lara", "Shannon", "Sexton", "Pace", "Hull", "Leblanc", "Browning", "Velasquez", "Leach", "Chang")
	surnameSlice = append(surnameSlice, "House", "Sellers", "Herring", "Noble", "Foley", "Bartlett", "Mercado", "Landry", "Durham", "Walls", "Barr", "Mckee", "Bauer", "Rivers")
	surnameSlice = append(surnameSlice, "Everett", "Bradshaw", "Pugh", "Velez", "Rush", "Estes", "Dodson", "Morse", "Sheppard", "Weeks", "Camacho", "Bean", "Barron", "Livingston")
	surnameSlice = append(surnameSlice, "Middleton", "Spears", "Branch", "Blevins", "Chen", "Kerr", "Mcconnell", "Hatfield", "Harding", "Ashley", "Solis", "Herman", "Frost", "Giles")
	surnameSlice = append(surnameSlice, "Blackburn", "William", "Pennington", "Woodward", "Finley", "Mcintosh", "Koch", "Best", "Solomon", "Mccullough", "Dudley", "Nolan", "Blanchard", "Rivas")
	surnameSlice = append(surnameSlice, "Brennan", "Mejia", "Kane", "Benton", "Joyce", "Buckley", "Haley", "Valentine", "Maddox", "Russo", "Mcknight", "Buck", "Moon", "Mcmillan")
	surnameSlice = append(surnameSlice, "Crosby", "Berg", "Dotson", "Mays", "Roach", "Church", "Chan", "Richmond", "Meadows", "Faulkner", "Oneill", "Knapp", "Kline", "Barry")
	surnameSlice = append(surnameSlice, "Ochoa", "Jacobson", "Gay", "Avery", "Hendricks", "Horne", "Shepard", "Hebert", "Cherry", "Cardenas", "Mcintyre", "Whitney", "Waller", "Holman")
	surnameSlice = append(surnameSlice, "Donaldson", "Cantu", "Terrell", "Morin", "Gillespie", "Fuentes", "Tillman", "Sanford", "Bentley", "Peck", "Key", "Salas", "Rollins", "Gamble")
	surnameSlice = append(surnameSlice, "Dickson", "Battle", "Santana", "Cabrera", "Cervantes", "Howe", "Hinton", "Hurley", "Spence", "Zamora", "Yang", "Mcneil", "Suarez", "Case")
	surnameSlice = append(surnameSlice, "Petty", "Gould", "Mcfarland", "Sampson", "Carver", "Bray", "Rosario", "Macdonald", "Stout", "Hester", "Melendez", "Dillon", "Farley", "Hopper")
	surnameSlice = append(surnameSlice, "Galloway", "Potts", "Bernard", "Joyner", "Stein", "Aguirre", "Osborn", "Mercer", "Bender", "Franco", "Rowland", "Sykes", "Benjamin", "Travis")
	surnameSlice = append(surnameSlice, "Pickett", "Crane", "Sears", "Mayo", "Dunlap", "Hayden", "Wilder", "Mckay", "Coffey", "Mccarty", "Ewing", "Cooley", "Vaughan", "Bonner")
	surnameSlice = append(surnameSlice, "Cotton", "Holder", "Stark", "Ferrell", "Cantrell", "Fulton", "Lynn", "Lott", "Calderon", "Rosa", "Pollard", "Hooper", "Burch", "Mullen")
	surnameSlice = append(surnameSlice, "Fry", "Riddle", "Levy", "David", "Duke", "Odonnell", "Guy", "Michael", "Britt", "Frederick", "Daugherty", "Berger", "Dillard", "Alston")
	surnameSlice = append(surnameSlice, "Jarvis", "Frye", "Riggs", "Chaney", "Odom", "Duffy", "Fitzpatrick", "Valenzuela", "Merrill", "Mayer", "Alford", "Mcpherson", "Acevedo", "Donovan")
	surnameSlice = append(surnameSlice, "Barrera", "Albert", "Cote", "Reilly", "Compton", "Raymond", "Mooney", "Mcgowan", "Craft", "Cleveland", "Clemons", "Wynn", "Nielsen", "Baird")
	surnameSlice = append(surnameSlice, "Stanton", "Snider", "Rosales", "Bright", "Witt", "Stuart", "Hays", "Holden", "Rutledge", "Kinney", "Clements", "Castaneda", "Slater", "Hahn")
	surnameSlice = append(surnameSlice, "Emerson", "Conrad", "Burks", "Delaney", "Pate", "Lancaster", "Sweet", "Justice", "Tyson", "Sharpe", "Whitfield", "Talley", "Macias", "Irwin")
	surnameSlice = append(surnameSlice, "Burris", "Ratliff", "Mccray", "Madden", "Kaufman", "Beach", "Goff", "Cash", "Bolton", "Mcfadden", "Levine", "Good", "Byers", "Kirkland")
	surnameSlice = append(surnameSlice, "Kidd", "Workman", "Carney", "Dale", "Mcleod", "Holcomb", "England", "Finch", "Head", "Burt", "Hendrix", "Sosa", "Haney", "Franks")
	surnameSlice = append(surnameSlice, "Sargent", "Nieves", "Downs", "Rasmussen", "Bird", "Hewitt", "Lindsay", "Le", "Foreman", "Valencia", "Oneil", "Delacruz", "Vinson", "Dejesus")
	surnameSlice = append(surnameSlice, "Hyde", "Forbes", "Gilliam", "Guthrie", "Wooten", "Huber", "Barlow", "Boyle", "Mcmahon", "Buckner", "Rocha", "Puckett", "Langley", "Knowles")
	return surnameSlice
}

func createEmailDomainSlice() []string {
	var emailSlice []string
	emailSlice = append(emailSlice, "gmail.com", "yahoo.com", "hotmail.com", "aol.com", "hotmail.co.uk", "hotmail.fr", "msn.com", "yahoo.fr", "wanadoo.fr", "orange.fr", "comcast.net", "yahoo.co.uk", "yahoo.com.br", "yahoo.co.in", "live.com")
	emailSlice = append(emailSlice, "rediffmail.com", "free.fr", "gmx.de", "web.de", "yandex.ru", "ymail.com", "libero.it", "outlook.com", "uol.com.br", "bol.com.br", "mail.ru", "cox.net", "hotmail.it", "sbcglobal.net")
	emailSlice = append(emailSlice, "sfr.fr", "live.fr", "verizon.net", "live.co.uk", "googlemail.com", "yahoo.es", "ig.com.br", "live.nl", "bigpond.com", "terra.com.br", "yahoo.it", "neuf.fr", "yahoo.de", "alice.it")
	emailSlice = append(emailSlice, "rocketmail.com", "att.net", "laposte.net", "facebook.com", "bellsouth.net", "yahoo.in", "hotmail.es", "charter.net", "yahoo.ca", "yahoo.com.au", "rambler.ru", "hotmail.de", "tiscali.it", "shaw.ca")
	emailSlice = append(emailSlice, "yahoo.co.jp", "sky.com", "earthlink.net", "optonline.net", "freenet.de", "t-online.de", "aliceadsl.fr", "virgilio.it", "home.nl", "qq.com", "telenet.be", "163.com", "yahoo.com.ar", "tiscali.co.uk")
	emailSlice = append(emailSlice, "yahoo.com.mx", "voila.fr", "gmx.net", "mail.com", "planet.nl", "126.com", "live.it", "ntlworld.com", "arcor.de", "yahoo.co.id", "frontiernet.net", "sina.com", "live.com.au", "yahoo.com.sg")
	emailSlice = append(emailSlice, "zonnet.nl", "club-internet.fr", "juno.com", "optusnet.com.au", "blueyonder.co.uk", "bluewin.ch", "skynet.be", "sympatico.ca", "windstream.net", "mac.com", "centurytel.net", "chello.nl", "live.ca", "aim.com")
	return emailSlice
}

func generateRandomName(quantity int, cUsername structUsername) []string {
	/* Code to read in lists of names to create the above functions for firstname and surname
	file, err := os.Open("names/emailDomains.txt")
	cf.CheckError("Unable to Open File", err, true)
	scanner := bufio.NewScanner(file)
	counter := 1
	nameStr := ""
	for scanner.Scan() {
		name := scanner.Text()
		//name = strings.ToUpper(name[0:1]) + strings.ToLower(name[1:])
		nameStr += "\"" + name + "\","
		if counter == 15 {
			fmt.Println("emailSlice  = append(emailSlice, " + nameStr[:len(nameStr)-1] + ")")
			nameStr = ""
			counter = 1
		}
		counter += 1
	}
	*/
	firstNameSlice := createFirstNameSlice()
	surnameSlice := createSurnameSlice()
	emailSlice := createEmailDomainSlice()
	var fakeNameSlice []string
	for len(fakeNameSlice) < quantity {
		randomFirstNameIndex := rand.Intn(len(firstNameSlice))
		randomMiddleInitial := rune('A' + rand.Intn('Z'-'A'+1))
		randomSurnameIndex := rand.Intn(len(surnameSlice))
		randomEmailIndex := rand.Intn(len(emailSlice))
		var userName string
		firstName := strings.ToLower(firstNameSlice[randomFirstNameIndex])
		firstInitial := string(firstName[0])
		lastName := strings.ToLower(surnameSlice[randomSurnameIndex])
		lastInitial := string(lastName[0])
		// Read the config for the way the username should be structured
		// Note that if more than one is yes then it selects the first based on the order below
		if cUsername.FirstNameLastName == "y" {
			userName = firstName + cUsername.Delimeter + lastName
		} else if cUsername.FirstInitialLastName == "y" {
			userName = firstInitial + cUsername.Delimeter + lastName
		} else if cUsername.FirstNameLastInitial == "y" {
			userName = firstName + cUsername.Delimeter + lastInitial
		} else {
			userName = firstName + cUsername.Delimeter + lastName
		}
		email := userName + "@" + emailSlice[randomEmailIndex]
		//fmt.Printf("%s %c. %s\n", firstNameSlice[randomFirstNameIndex], randomMiddleInitial, surnameSlice[randomSurnameIndex])
		// Fullname, First Name, Middle Initial, Last Name, Username, Email
		randomName := fmt.Sprintf("%s %c. %s,%s,%c,%s,%s,%s", firstNameSlice[randomFirstNameIndex], randomMiddleInitial, surnameSlice[randomSurnameIndex], firstNameSlice[randomFirstNameIndex], randomMiddleInitial, surnameSlice[randomSurnameIndex], userName, email)
		fakeNameSlice = append(fakeNameSlice, randomName)
	}
	return fakeNameSlice
}

func generateRandomMaidenName(quantity int) []string {
	surnameSlice := createSurnameSlice()
	var fakeMaidenNameSlice []string
	for len(fakeMaidenNameSlice) < quantity {
		randomSurnameIndex := rand.Intn(len(surnameSlice))
		//fmt.Printf("%s %c. %s\n", firstNameSlice[randomFirstNameIndex], randomMiddleInitial, surnameSlice[randomSurnameIndex])
		// Maiden Name
		randomName := fmt.Sprintf("%s", surnameSlice[randomSurnameIndex])
		fakeMaidenNameSlice = append(fakeMaidenNameSlice, randomName)
	}
	return fakeMaidenNameSlice
}

func containsNumber(s string) bool {
	re := regexp.MustCompile("[0-9]")
	return re.MatchString(s)
}

func containsUpperCase(s string) bool {
	re := regexp.MustCompile("[A-Z]")
	return re.MatchString(s)
}

func containsLowerCase(s string) bool {
	re := regexp.MustCompile("[a-z]")
	return re.MatchString(s)
}

func generateRandomPassword(length int, quantity int) []string {
	// Define the characters allowed in the password
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var password []byte
	passwordSlice := []string{}

	for i := 1; i <= quantity; i++ {
		// Generate the random password
		complexPassword := false
		for complexPassword == false {
			password = make([]byte, length)
			for i := range password {
				password[i] = chars[rand.Intn(len(chars))]
			}
			// Verify the password contains a number, lowercase letter and an upper-case letter
			if containsNumber(string(password)) && containsLowerCase(string(password)) && containsUpperCase(string(password)) {
				complexPassword = true
			}
		}
		passwordSlice = append(passwordSlice, string(password))
	}

	return passwordSlice
}

func main() {
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	QuantityPtr := flag.String("q", "0", "Default is to read from the config.json")
	flag.Parse()

	// Load the config.json file
	var config structConfig
	fmt.Println("Loading the following config file: " + *ConfigPtr + "\n")
	configFile, err := os.Open(*ConfigPtr)
	CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		CheckError("Unable to decode the configuration file", err, true)
	}

	intQuantity, err := strconv.Atoi(*QuantityPtr)
	if err != nil {
		fmt.Println("Error: Unable to convert quantity string to integer", err)
		return
	}
	var quantity int
	// Placed this condition back in the code so that the batch script mentioned above is still valid
	if intQuantity > 0 {
		quantity = intQuantity
	} else {
		quantity = config.OutputQuantity
	}
	//filetype := "csv" // Future filetypes json

	/*
		var delimeter string
		if filetype == "csv" {
			delimeter = "," // Future delimeters "\t" or "\s"
		} else {
			delimeter = ""
		}
	*/
	// Create a fake list of CC Numbers with a BIN of 12345, for Visa with a specified qunatity of 5
	fakeCCList := generateFakeCCNumbers(config.CreditCard.VisaBinNumber, "Visa", quantity)

	// Create fake expiration dates for the CC
	fakeExpirationDateList := generateFakeExpirationDate(quantity, config.ExpDateCard.MinExpDateMonths, config.ExpDateCard.MaxExpDateMonths)

	// Create Fake CVV
	fakeCVVList := generateFakeCVV(quantity)

	// Create Fake ATM Pin
	fakePinList := generateFakeATMPin(quantity)

	// Create a fake list of SSN Numbers
	fakeSSNList := generateFakeSSN(quantity)

	// Create a fake list of Dates of Birth between 18 and 102
	minAge := config.DOB.MinAgeYears
	maxAge := config.DOB.MaxAgeYears
	fakeDOBSlice := generateFakeDOB(minAge, maxAge, quantity)

	// Create random name
	fakeNameSlice := generateRandomName(quantity, config.UserName)
	//fmt.Println(fakeNameSlice)

	// Create random maiden name
	fakeMaidenSlice := generateRandomMaidenName(quantity)

	// Create random phone number slice
	fakePhoneSlice := generateFakePhone(quantity)
	fakeCellPhoneSlice := generateFakePhone(quantity)
	fakeWorkPhoneSlice := generateFakePhone(quantity)

	// Generate a random password
	fakePasswordSlice := generateRandomPassword(config.Password.Length, quantity)

	csvOutput := ""

	csvOutput += fmt.Sprintln("CC,ExpDate,CVV,ATMPin,SSN,DOB,Fullname,FirstName,MiddleInitial,LastName,Username,Email,HomePhone,CellPhone,WorkPhone,MothersMaidenName,Password")
	for i := 0; i < quantity; i++ {
		// Making it a CSV
		csvOutput += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n", fakeCCList[i], fakeExpirationDateList[i], fakeCVVList[i], fakePinList[i], fakeSSNList[i], fakeDOBSlice[i], fakeNameSlice[i], fakePhoneSlice[i], fakeCellPhoneSlice[i], fakeWorkPhoneSlice[i], fakeMaidenSlice[i], fakePasswordSlice[i])
	}
	// Output the original CSV without the ordering specified in the configuration
	//fmt.Print(csvOutput)
	SaveOutputFile(csvOutput, config.OriginalFileOutput)
	fmt.Printf("Saved the original fake data to: %s\n\n", config.OriginalFileOutput)

	// Convert the variable of the CSV Output into a reader and then arrange the
	// order based on the provided order from the config file...
	reader := csv.NewReader(strings.NewReader(csvOutput))
	records, err := reader.ReadAll()
	CheckError("Unable to read the csvOutput generated", err, true)

	columnsToOrder := []int{}
	for i := 1; i <= 17; i++ {
		// CC
		if config.CreditCard.Order == i && strings.ToLower(config.CreditCard.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 0)
		}
		// Expiration Date
		if config.ExpDateCard.Order == i && strings.ToLower(config.ExpDateCard.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 1)
		}
		// CVV Card
		if config.CVVCard.Order == i && strings.ToLower(config.CVVCard.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 2)
		}
		// ATM Pin
		if config.ATMPin.Order == i && strings.ToLower(config.ATMPin.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 3)
		}
		// SSN
		if config.SSN.Order == i && strings.ToLower(config.SSN.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 4)
		}
		// DOB
		if config.DOB.Order == i && strings.ToLower(config.DOB.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 5)
		}
		// Fullname
		if config.Fullname.Order == i && strings.ToLower(config.Fullname.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 6)
		}
		// Firstname
		if config.FirstName.Order == i && strings.ToLower(config.FirstName.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 7)
		}
		// Middle Initial
		if config.MiddleInitial.Order == i && strings.ToLower(config.MiddleInitial.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 8)
		}
		// LastName
		if config.LastName.Order == i && strings.ToLower(config.LastName.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 9)
		}
		// Username
		if config.UserName.Order == i && strings.ToLower(config.UserName.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 10)
		}
		// Email
		if config.Email.Order == i && strings.ToLower(config.Email.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 11)
		}
		// HomePhone
		if config.HomePhone.Order == i && strings.ToLower(config.HomePhone.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 12)
		}
		// CellPhone
		if config.CellPhone.Order == i && strings.ToLower(config.CellPhone.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 13)
		}
		// WorkPhone
		if config.WorkPhone.Order == i && strings.ToLower(config.WorkPhone.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 14)
		}
		// Mothers Maiden Name
		if config.MothersMaidenName.Order == i && strings.ToLower(config.MothersMaidenName.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 15)
		}
		// Random Password
		if config.Password.Order == i && strings.ToLower(config.Password.Display) == "y" {
			columnsToOrder = append(columnsToOrder, 16)
		}

	}

	// Pull the order of the columns from the configuration file
	//columnsToOrder := []int{3, 2}
	newCSVOutput := ""
	for _, row := range records {
		newRowSlice := []string{}
		for _, colIndex := range columnsToOrder {
			//fmt.Println(strings.Join(row, ","))
			newRowSlice = append(newRowSlice, fmt.Sprintf(row[colIndex]))
		}
		newCSVOutput += fmt.Sprintln(strings.Join(newRowSlice, ","))
	}

	SaveOutputFile(newCSVOutput, config.FileOutput)
	fmt.Printf("Saved the re-ordered fake data to: %s\n\n", config.FileOutput)
}
