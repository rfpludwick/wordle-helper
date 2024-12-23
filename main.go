package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	color "github.com/TwiN/go-color"
	"golang.org/x/exp/slices"
)

const (
	GuessPositionUnknown   = 0
	GuessPositionWrong     = 1
	GuessPositionMisplaced = 2
	GuessPositionCorrect   = 3
)

type AlphabetLetter struct {
	Letter rune
}

type AlphabetLetterPosition struct {
	Status uint8
}

type DictionaryWord struct {
	Value   uint64
	Valid   bool
	Letters []rune
}

type Guess struct {
	Valid     bool
	Positions []GuessPosition
}

type GuessPosition struct {
	Letter rune
	Status uint8
}

func newDictionaryWord(word string) DictionaryWord {
	dw := DictionaryWord{
		Value:   0,
		Valid:   true,
		Letters: make([]rune, 5),
	}

	for letterPosition, letter := range word {
		dw.Letters[letterPosition] = letter
	}

	return dw
}

func newGuess() Guess {
	return Guess{
		Positions: make([]GuessPosition, 5),
		Valid:     false,
	}
}

func (dw *DictionaryWord) getWord() string {
	var str string

	for _, letter := range dw.Letters {
		str += string(letter)
	}

	return str
}

func (g *Guess) getWord() string {
	var str string

	for _, position := range g.Positions {
		str += string(position.Letter)
	}

	return str
}

var (
	dictionaryWords []DictionaryWord
	currentGuess    uint8
	guesses         []Guess
)

func init() {
	// Initialize dictionary words
	file, err := os.Open("/usr/share/dict/words")

	if err != nil {
		log.Fatalln(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		word := scanner.Text()

		if len(word) != 5 {
			continue
		}

		word = strings.ToUpper(word)

		dictionaryWords = append(dictionaryWords, newDictionaryWord(word))
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}

	// Scrub previous answers from dictionary words
	file, err = os.Open("./answers.txt")

	if err != nil {
		log.Fatalln(err)
	}

	defer file.Close()

	scanner = bufio.NewScanner(file)

	for scanner.Scan() {
		scrubWord(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}

	// Remaining init
	currentGuess = 0
	guesses = make([]Guess, 5)
}

func main() {
	showHelp()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n> Enter command: ")

		command := getInput(scanner)

		fmt.Println()

		switch command {
		case "h":
			showHelp()
		case "b":
			if currentGuess == 0 {
				fmt.Println("No guesses logged yet")

				continue
			}

			for _, guess := range guesses {
				if !guess.Valid {
					break
				}

				output := color.Bold + color.White

				for _, guessPosition := range guess.Positions {
					switch guessPosition.Status {
					case GuessPositionWrong:
						output += color.GrayBackground
					case GuessPositionMisplaced:
						output += color.YellowBackground
					case GuessPositionCorrect:
						output += color.GreenBackground
					default:
						log.Fatalf("Invalid guess position status encountered: %d\n", guessPosition.Status)
					}

					output += string(guessPosition.Letter)
				}

				fmt.Println(output + color.Reset)
			}
		case "c":
			guess := promptGuess(scanner, "possibility", false)
			guess.Valid = false

			for _, dictionaryWord := range dictionaryWords {
				if dictionaryWord.getWord() == guess.getWord() {
					guess.Valid = dictionaryWord.Valid

					break
				}
			}

			if guess.Valid {
				fmt.Printf("\n%s is a valid guess\n", guess.getWord())
			} else {
				fmt.Printf("\n%s is an invalid guess\n", guess.getWord())
			}
		case "p":
			showTopPossibilities()
		case "g":
			promptGuess(scanner, fmt.Sprintf("%d", (currentGuess+1)), true)
		case "q":
			quit()
		default:
			fmt.Printf("Unknown command: %s\n", command)
		}
	}
}

func showHelp() {
	fmt.Println("Commands")
	fmt.Println("  h: help")
	fmt.Println("  b: show board")
	fmt.Println("  c: check possible guess")
	fmt.Println("  p: show top possibilities")
	fmt.Println("  g: make guess")
	fmt.Println("  q: quit")
	fmt.Println("Position Statuses")
	fmt.Println("  h: help (convenience helper)")
	fmt.Println("  w: wrong")
	fmt.Println("  m: misplaced")
	fmt.Println("  c: correct")
	fmt.Println("  q: quit (convenience helper)")
}

func getInput(scanner *bufio.Scanner) string {
	scanner.Scan()

	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}

	return scanner.Text()
}

func quit() {
	fmt.Println("Quitting")

	os.Exit(0)
}

func promptGuess(scanner *bufio.Scanner, guessPrompt string, realGuess bool) *Guess {
	var guess Guess

	for {
		fmt.Printf("> Enter guess %s: ", guessPrompt)

		word := getInput(scanner)

		if word == "q" {
			quit()
		}

		if word == "h" {
			showHelp()

			continue
		}

		if len(word) != 5 {
			fmt.Printf("\nInvalid guess: %s\n\n", word)

			continue
		}

		guess = newGuess()
		guess.Valid = true

	guessLoop:
		for guessPosition, letter := range word {
			for {
				guessPositionCast := uint8(guessPosition)
				guessPositionIncrement := guessPositionCast + 1
				letterIsUpper := (letter >= 'A') && (letter <= 'Z')
				letterIsLower := (letter >= 'a') && (letter <= 'z')

				if !(letterIsUpper || letterIsLower) {
					fmt.Printf("\nInvalid character for position %d: %c\n", guessPositionIncrement, letter)

					guess.Valid = false

					break guessLoop
				}

				if letterIsLower {
					letter -= 32
				}

				positionStatus := uint8(GuessPositionUnknown)

				if realGuess {
					fmt.Printf("> Enter status for position %d; character %c: ", guessPositionIncrement, letter)

					status := getInput(scanner)

					switch status {
					case "h":
						showHelp()
					case "w":
						positionStatus = GuessPositionWrong

						scrubPositionWrong(guessPositionCast, letter)
					case "m":
						positionStatus = GuessPositionMisplaced

						scrubPositionWrong(guessPositionCast, letter) // Misplaced, so this position is still wrong
					case "c":
						positionStatus = GuessPositionCorrect

						scrubPositionCorrect(guessPositionCast, letter)
					case "q":
						quit()
					default:
						fmt.Printf("\nUnknown status: %s\n\n", status)

						continue
					}
				}

				position := GuessPosition{
					Letter: letter,
					Status: positionStatus,
				}

				guess.Positions[guessPositionCast] = position

				break
			}
		}

		if !(guess.Valid && realGuess) {
			break
		}

		guesses[currentGuess] = guess

		currentGuess++

		// The following combinations are considered "complete" letters:
		// Letters with only wrong positions
		//     Scrub all words with these letters present
		// Letters with 1+ correct positions and 1+ wrong positions and 0 misplaced positions
		//     Scrub all words with these letters in any incorrect positions

		correctLetterPositions := make(map[rune][]int, 0)
		wrongLetters := make(map[rune]bool, 0)
		misplacedLetters := make(map[rune]bool, 0)

		for i, guessPosition := range guess.Positions {
			switch guessPosition.Status {
			case GuessPositionWrong:
				wrongLetters[guessPosition.Letter] = true
			case GuessPositionMisplaced:
				misplacedLetters[guessPosition.Letter] = true

				scrubLetterMustBePresent(guessPosition.Letter)
			case GuessPositionCorrect:
				correctLetterPositions[guessPosition.Letter] = append(correctLetterPositions[guessPosition.Letter], i)
			}
		}

		for misplacedLetter := range misplacedLetters {
			delete(wrongLetters, misplacedLetter)
			delete(correctLetterPositions, misplacedLetter)
		}

		for i, dictionaryWord := range dictionaryWords {
			if !dictionaryWord.Valid {
				continue
			}

			for j, dwLetter := range dictionaryWord.Letters {
				_, wrongLetterPresent := wrongLetters[dwLetter]
				_, correctLetterPresent := correctLetterPositions[dwLetter]

				if !wrongLetterPresent {
					continue
				}

				if !correctLetterPresent {
					dictionaryWords[i].Valid = false

					break
				}

				if !slices.Contains(correctLetterPositions[dwLetter], j) {
					dictionaryWords[i].Valid = false
				}

				break
			}
		}

		break
	}

	return &guess
}

func showTopPossibilities() {
	shown := 0

	for _, dictionaryWord := range dictionaryWords {
		if !dictionaryWord.Valid {
			continue
		}

		fmt.Printf("%s\n", dictionaryWord.getWord())

		shown++

		if shown == 10 {
			break
		}
	}
}

func scrubWord(word string) {
	for i, dictionaryWord := range dictionaryWords {
		if dictionaryWord.getWord() == word {
			dictionaryWords[i].Valid = false
		}
	}
}

func scrubPositionWrong(position uint8, letter rune) {
	for i, dictionaryWord := range dictionaryWords {
		if !dictionaryWord.Valid {
			continue
		}

		if dictionaryWord.Letters[position] == letter {
			dictionaryWords[i].Valid = false
		}
	}
}

func scrubPositionCorrect(position uint8, letter rune) {
	for i, dictionaryWord := range dictionaryWords {
		if !dictionaryWord.Valid {
			continue
		}

		if dictionaryWord.Letters[position] != letter {
			dictionaryWords[i].Valid = false
		}
	}
}

func scrubLetterMustBePresent(letter rune) {
	for i, dictionaryWord := range dictionaryWords {
		if !dictionaryWord.Valid {
			continue
		}

		dictionaryWords[i].Valid = false

		for _, dwLetter := range dictionaryWord.Letters {
			if dwLetter == letter {
				dictionaryWords[i].Valid = true

				break
			}
		}
	}
}
