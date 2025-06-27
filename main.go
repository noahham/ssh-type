package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func getWords(filepath string) ([]string, error) {
	// Open the file.
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	// Ensure the file is closed when the function exits.
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("error closing file: %v\n", err)
		}
	}(file)

	var lines []string
	// Create a new scanner to read line by line.
	scanner := bufio.NewScanner(file)

	// Iterate through each line in the file.
	for scanner.Scan() {
		lines = append(lines, scanner.Text()) // Append the current line (as text) to the slice.
	}

	// Check for any errors during scanning.
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return lines, nil
}

func getRandomWords(numWords int) string {
	words, _ := getWords("words.txt")
	wordList := make([]string, numWords)
	for i := 0; i < numWords; i++ {
		wordList[i] = words[rand.Intn(len(words))]
	}

	return strings.Join(wordList, " ")
}

// Defines model structure
type model struct {
	input     string
	started   bool
	startTime time.Time
}

func main() {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// Initial state
func initialModel() model {
	return model{input: getRandomWords(10)}
}

// Update function: handle input and state changes
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if !m.started {
				m.started = true
				m.startTime = time.Now()
			}
			m.input = getRandomWords(10)
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			m.input += msg.String()
		}
	}

	return m, nil
}

// View function: render UI
func (m model) View() string {
	elapsed := time.Since(m.startTime).Seconds()

	return fmt.Sprintf(`
Type something! (Press Enter to reset, Esc to quit)

> %s

Time elapsed: %.1fs
`, m.input, elapsed)
}

func (m model) Init() tea.Cmd { return nil }

/*
ESSENTIAL:
1. Generate a list of x words
2. On first key input, start timer
3. For each key input, check if a letter is correct or incorrect

NICE TO HAVE:
1. Live WPM calculation
2. Limit inputs to length of word
3. Customizable text size
4. Customizable cursor color
*/
