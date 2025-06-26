package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	input     string
	started   bool
	startTime time.Time
}

// Define messages
type tickMsg time.Time

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// Initial state
func initialModel() model {
	return model{}
}

// Update function: handle input & state changes
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
			m.input = ""
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

func (m model) Init() tea.Cmd {
	return nil
}

/*
ESSENTIAL:
1. Generate a list of 100 words
2. On first key input, start a timer
3. For each key input, check if letter is correct or incorrect

NICE TO HAVE:
1. Live WPM calculation
2. Limit inputs to length of word
3. Customizable text size
4. Customizable cursor color
*/
