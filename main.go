package main

import (
	"bufio"
	"fmt"
	"github.com/charmbracelet/lipgloss"
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
	input    string
	started  bool
	timeLeft int
	styles   map[string]lipgloss.Style
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
	return model{
		input:    getRandomWords(10),
		timeLeft: 30,
	}
}

type tickMsg time.Time

func (m model) tick() tea.Cmd {
	if m.started {
		return tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return nil
}

// Update function: handle input and state changes
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			m.started = false

			m.input = getRandomWords(10)
			m.timeLeft = 40

		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			m.input += msg.String()
			if !m.started {
				m.started = true
				return m, m.tick() // Start the timer on first input
			}
		}

	case tickMsg:
		if !m.started {
			return m, nil // Don't tick or decrement time if not started
		}
		if m.timeLeft > 0 {
			m.timeLeft--
		}
		if m.timeLeft == 0 {
			m.started = false // auto-stop
			return m, nil
		}
		return m, m.tick() // schedule next tick
	}

	return m, nil
}

// View function: render UI
func (m model) View() string {
	return fmt.Sprintf(`
Type something! (Press Enter to reset, Esc to quit)

> %s

Time Left: %ds
`, m.input, m.timeLeft)
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
