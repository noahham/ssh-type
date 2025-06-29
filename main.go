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

func calculateWPM(words int, timeLeft int) int {
	if words == 0 && timeLeft == 0 {
		return 0
	}
	return words / 12 // TODO: Implement the live timer
}

type model struct {
	toType      string
	typed       string
	started     bool
	timeLeft    int
	styles      map[string]lipgloss.Style
	timeSetting int
	liveWPM     bool
}

func initialModel() model {
	textStyles := make(map[string]lipgloss.Style)
	textStyles["typed"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Height(3)
	textStyles["notTyped"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#595959")).
		Height(3)
	textStyles["header"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		PaddingTop(2)
	textStyles["spacer"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff"))
	textStyles["keybind"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#595959")).
		PaddingTop(2)
	textStyles["timer"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff0000")).
		Width(54).
		Align(lipgloss.Center).
		PaddingTop(1)
	textStyles["width"] = lipgloss.NewStyle().
		Width(54)

	return model{
		toType:      getRandomWords(50),
		typed:       "",
		styles:      textStyles,
		started:     false,
		timeLeft:    30, // Default time setting
		timeSetting: 30,
		liveWPM:     false,
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, tea.Quit
		case "enter":
			m.started = false

			m.toType = getRandomWords(50)
			m.timeLeft = m.timeSetting

		case "backspace":
			runes := []rune(m.typed)
			if len(runes) > 0 {
				m.typed = string(runes[:len(runes)-1])
			}

		case "1":
			if m.timeSetting == 15 {
				m.timeSetting = 30
			} else if m.timeSetting == 30 {
				m.timeSetting = 45
			} else {
				m.timeSetting = 15
			}
			m.timeLeft = m.timeSetting

		case "2":
			m.liveWPM = !m.liveWPM

		default:
			m.typed += msg.String()
			if msg.String()[0] == m.toType[0] {

			}
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

func (m model) View() string {
	keybind := m.styles["keybind"].Render
	header := m.styles["header"].Render

	keybinds := lipgloss.JoinHorizontal(lipgloss.Top,
		keybind(" esc"), header(" Exit "),
		header(" |  "), keybind("enter"), header(" Reset "),
		header(" |  "), keybind("1"), header(" Length "),
		header(" |  "), keybind("2"), header(" Live WPM"),
	)

	headerLine := m.styles["spacer"].Render("------------------------------------------------------")

	typedRunes := []rune(m.typed)
	targetRunes := []rune(m.toType)

	var typedStyled strings.Builder
	for i, r := range typedRunes {
		if i < len(targetRunes) {
			if r == targetRunes[i] {
				typedStyled.WriteString(m.styles["typed"].Render(string(r)))
			} else {
				typedStyled.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f5f")).Render(string(r))) // red for wrong
			}
		}
	}

	remaining := ""
	if len(typedRunes) < len(targetRunes) {
		remaining = string(targetRunes[len(typedRunes):])
	}

	words := m.styles["width"].Render(
		typedStyled.String() + m.styles["notTyped"].Render(remaining),
	)

	var info string
	if m.liveWPM {
		info = m.styles["timer"].Render(fmt.Sprintf("%ds", m.timeLeft) + "    " + fmt.Sprint(12)) // TODO: Replace 12 with actual WPM calculation
	} else {
		info = m.styles["timer"].Render(fmt.Sprintf("%ds", m.timeLeft))
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		keybinds,
		headerLine,
		"",
		words,
		"",
		info,
	)
}

func (m model) Init() tea.Cmd { return nil }

func main() {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

/*
ESSENTIAL:
1. Generate a list of x words
2. On first key input, start timer
3. For each key input, check if a letter is correct or incorrect

NICE TO HAVE:
1. Live WPM calculation
2. Limit inputs to length of word
*/
