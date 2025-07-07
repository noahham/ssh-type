package main

import (
	"bufio"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
)

func makeWordList(filepath string) ([]string, error) {
	// Reads a file and returns a slice of strings, each representing a line in the file.
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("error closing file: %v\n", err)
		}
	}(file)

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Check for any errors during scanning.
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return lines, nil
}

func getRandomWords(numWords int) string {
	// Gets numWords random words from a file named "words.txt"
	words, _ := makeWordList("words.txt")
	wordList := make([]string, numWords)
	for i := 0; i < numWords; i++ {
		wordList[i] = words[rand.Intn(len(words))]
	}

	return strings.Join(wordList, " ")
}

func calculateWPM(typed string, toType string, timeLeft int, totalTime int, backspaceErrors int) float32 {
	// Calculates the WPM based on the typed text, the text to type, time left, total time, and backspace errors.
	if typed == "" {
		return 0
	}

	if len(typed) < len(toType) {
		toType = toType[:len(typed)]
	}

	wordCount := len(strings.Fields(toType))
	accuracy := getAccuracy(typed, toType, backspaceErrors)
	timeElapsed := totalTime - timeLeft
	if timeElapsed <= 0 {
		return 0
	}

	rawWPM := (float32(wordCount) / float32(timeElapsed)) * 60
	adjustedWPM := rawWPM * accuracy

	return float32(math.Round(float64(adjustedWPM)*10) / 10)
}

func getAccuracy(toType string, typed string, backspaceErrors int) float32 {
	// Calculates the accuracy of the typed text compared to the text to type.
	correct := 0
	for i := range toType {
		if toType[i] == typed[i] {
			correct++
		}
	}

	correctness := float32(correct) / float32(len(toType)+backspaceErrors)
	return correctness
}

type model struct {
	toType          string
	typed           string
	started         bool
	timeLeft        int
	styles          map[string]lipgloss.Style
	timeSetting     int
	liveWPMEnabled  bool
	backspaceErrors int // Count of backspace errors
	windowWidth     int
	windowHeight    int
}

func initialModel() model {
	// Defining text styles map using lipgloss
	textStyles := make(map[string]lipgloss.Style)
	textStyles["typed"] = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"})
	textStyles["notTyped"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#595959"))
	textStyles["incorrect"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff5f5f"))
	textStyles["header"] = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"}).
		PaddingTop(2)
	textStyles["spacer"] = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"})
	textStyles["keybind"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#595959")).
		PaddingTop(2)
	textStyles["timer"] = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff0000")).
		PaddingTop(1)
	textStyles["width"] = lipgloss.NewStyle().
		Width(56)
	textStyles["widthAndCenter"] = lipgloss.NewStyle().
		Width(56).
		Align(lipgloss.Center)
	textStyles["cursor"] = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#000000"}).
		Background(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"})

	return model{
		toType:          getRandomWords(50), // Sets 50 random words as text
		typed:           "",
		styles:          textStyles,
		started:         false,
		timeLeft:        30,
		timeSetting:     30, // Default time is 30s
		liveWPMEnabled:  false,
		backspaceErrors: 0,
	}
}

type tickMsg time.Time

func (m model) tick() tea.Cmd {
	// Returns a command to tick every second if the timer has already started
	if m.started {
		return tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Updates the model state based on the received message
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc": // Quit
			return m, tea.Quit
		case "enter": // Restart
			m.started = false

			m.toType = getRandomWords(50)
			m.typed = ""
			m.timeLeft = m.timeSetting

		case "backspace": // Backspace
			if m.timeLeft != 0 {
				m.backspaceErrors++
				runes := []rune(m.typed)
				if len(runes) > 0 {
					m.typed = string(runes[:len(runes)-1])
				}
			}

		case "1": // Change Length
			if m.timeSetting == 15 {
				m.timeSetting = 30
			} else if m.timeSetting == 30 {
				m.timeSetting = 45
			} else {
				m.timeSetting = 15
			}
			m.timeLeft = m.timeSetting
			m.started = false
			m.typed = ""
			m.backspaceErrors = 0

		case "2": // Enable/Disable Live WPM
			m.liveWPMEnabled = !m.liveWPMEnabled

		default: // Normal Typing
			// Only allows A–Z or a–z keys
			if len(msg.String()) == 1 {
				r := rune(msg.String()[0])
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == ' ' {
					if m.timeLeft != 0 && len(m.typed) < len(m.toType) {
						m.typed += string(r)
						if !m.started {
							m.started = true
							return m, m.tick() // Start the timer on first input
						}
					}
				}
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

	// If the window is too small...
	if m.windowWidth < 56 || m.windowHeight < 13 {
		message := m.styles["header"].Render("Terminal is too small!")
		return lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, message)
	}

	// If the timer is running, show the typing interface
	if m.timeLeft != 0 {
		keybinds := lipgloss.JoinHorizontal(lipgloss.Top,
			keybind(" esc"), header(" Exit "),
			header(" |  "), keybind("enter"), header(" Restart "),
			header(" |  "), keybind("1"), header(" Length "),
			header(" |  "), keybind("2"), header(" Live WPM"),
		)

		headerLine := m.styles["spacer"].Render("--------------------------------------------------------")

		typedRunes := []rune(m.typed)
		targetRunes := []rune(m.toType)

		var typedStyled strings.Builder
		for i, r := range typedRunes {
			if i < len(targetRunes) {
				if r == targetRunes[i] {
					typedStyled.WriteString(m.styles["typed"].Render(string(r)))
				} else {
					typedStyled.WriteString(m.styles["incorrect"].Render(string(r))) // red for wrong
				}
			}
		}

		remaining := ""
		if len(typedRunes) < len(targetRunes) {
			remainingRunes := targetRunes[len(typedRunes):]

			if len(remainingRunes) > 0 {
				first := string(remainingRunes[0]) // The first untyped character is the cursor
				rest := string(remainingRunes[1:])

				remaining = m.styles["cursor"].Render(first) + m.styles["notTyped"].Render(rest)
			}
		}

		words := m.styles["width"].Render(
			typedStyled.String() + m.styles["notTyped"].Render(remaining),
		)

		var info string
		if m.liveWPMEnabled {
			info = m.styles["timer"].Width(56).Align(lipgloss.Center).Render(fmt.Sprintf("%ds", m.timeLeft) + "    " + fmt.Sprint(calculateWPM(m.typed, m.toType, m.timeLeft, m.timeSetting, m.backspaceErrors)))
		} else {
			info = m.styles["timer"].Width(56).Align(lipgloss.Center).Render(fmt.Sprintf("%ds", m.timeLeft))
		}

		noCenter := lipgloss.JoinVertical(lipgloss.Top,
			keybinds,
			headerLine,
			"",
			words,
			"",
			info,
		)

		return lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, noCenter)
	} else { // If the timer has stopped, show the summary
		wpm := calculateWPM(m.typed, m.toType, m.timeLeft, m.timeSetting, m.backspaceErrors)

		summary := lipgloss.JoinHorizontal(lipgloss.Top,
			keybind("WPM "),
			header(fmt.Sprintf("%d", int(wpm))),
			keybind("    Accuracy "),
			header(fmt.Sprintf("%.0f%%", getAccuracy(m.toType[:len(m.typed)], m.typed, m.backspaceErrors)*100)),
		)

		reset := lipgloss.JoinHorizontal(lipgloss.Top,
			m.styles["timer"].Foreground(lipgloss.Color("#595959")).Render("enter "),
			m.styles["timer"].Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"}).Render("Restart"),
		)

		noCenter := lipgloss.JoinVertical(lipgloss.Top,
			m.styles["widthAndCenter"].Render(summary),
			m.styles["widthAndCenter"].Render(reset),
		)

		return lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, noCenter)
	}
}

func (m model) Init() tea.Cmd { return nil }

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
