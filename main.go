package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type projectType int

const (
	Vanilla projectType = iota
	React
	NextJS
	Vue
	Svelte
	Automatic
)

func (p projectType) String() string {
	return [...]string{"Vanilla JS", "React", "Next.js", "Vue", "Svelte", "Automatic"}[p]
}

type optionItem int

const (
	optionProjectType optionItem = iota
	optionDirectory
	optionModel
	optionExecute
)

type llmModel int

const (
	modelGPT4 llmModel = iota
	modelClaude
	modelLocal
)

func (l llmModel) String() string {
	return [...]string{"GPT-4", "Claude Sonnet", "Local (Ollama)"}[l]
}

type section int

const (
	sectionOptions section = iota
	sectionPrompt
)

type step struct {
	title       string
	description string
	status      string
	timestamp   string
}

type tickMsg time.Time

type model struct {
	projectType    projectType
	projectDir     string
	llmModel       llmModel
	promptInput    textinput.Model
	steps          []step
	width          int
	height         int
	activeSection  section
	selectedOption optionItem
	processing     bool
	animFrame      int
	statusMessage  string
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Describe your coding task in detail..."
	ti.CharLimit = 1000
	ti.Width = 50

	return model{
		projectType:    Automatic,
		projectDir:     "./orbis-project",
		llmModel:       modelClaude,
		promptInput:    ti,
		steps:          []step{},
		activeSection:  sectionOptions,
		selectedOption: optionProjectType,
		processing:     false,
		animFrame:      0,
		statusMessage:  "Ready to build",
		width:          100,
		height:         30,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tickCmd())
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab":
			m.activeSection = (m.activeSection + 1) % 2
			if m.activeSection == sectionPrompt {
				m.promptInput.Focus()
			} else {
				m.promptInput.Blur()
			}
			return m, nil

		case "up":
			if m.activeSection == sectionOptions {
				if m.selectedOption > 0 {
					m.selectedOption--
				}
			}
			return m, nil

		case "down":
			if m.activeSection == sectionOptions {
				if m.selectedOption < optionExecute {
					m.selectedOption++
				}
			}
			return m, nil

		case "left":
			if m.activeSection == sectionOptions {
				switch m.selectedOption {
				case optionProjectType:
					m.projectType = (m.projectType + 5) % 6
				case optionModel:
					m.llmModel = (m.llmModel + 2) % 3
				}
			}
			return m, nil

		case "right":
			if m.activeSection == sectionOptions {
				switch m.selectedOption {
				case optionProjectType:
					m.projectType = (m.projectType + 1) % 6
				case optionModel:
					m.llmModel = (m.llmModel + 1) % 3
				}
			}
			return m, nil

		case "enter":
			if m.activeSection == sectionOptions && m.selectedOption == optionExecute {
				if m.promptInput.Value() != "" && !m.processing {
					m.processing = true
					m.statusMessage = "Initializing..."
					m.steps = []step{
						{title: "Analyzing prompt", description: "Understanding requirements", status: "running", timestamp: currentTime()},
						{title: "Planning structure", description: "Designing architecture", status: "pending", timestamp: ""},
						{title: "Generating code", description: "Creating files", status: "pending", timestamp: ""},
						{title: "Finalizing", description: "Installing dependencies", status: "pending", timestamp: ""},
					}
					return m, m.processWithLLM()
				}
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		if m.processing {
			m.animFrame = (m.animFrame + 1) % 4
		}
		return m, tickCmd()

	case llmResponseMsg:
		m.processing = false
		m.statusMessage = "Build complete!"
		for i := range m.steps {
			m.steps[i].status = "complete"
			if m.steps[i].timestamp == "" {
				m.steps[i].timestamp = currentTime()
			}
		}
		return m, nil
	}

	if m.activeSection == sectionPrompt {
		m.promptInput, cmd = m.promptInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	// Calculate dynamic dimensions
	contentWidth := m.width - 4
	if contentWidth < 60 {
		contentWidth = 60
	}
	if contentWidth > 120 {
		contentWidth = 120
	}

	contentHeight := m.height - 4
	if contentHeight < 20 {
		contentHeight = 20
	}

	// Main container
	mainStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	// Border style
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("39")).
		Width(contentWidth).
		Height(contentHeight).
		Padding(1, 2)

	var content strings.Builder

	// Header with logo
	content.WriteString(m.renderHeader(contentWidth - 4))
	content.WriteString("\n")

	// Status bar
	content.WriteString(m.renderStatusBar(contentWidth - 4))
	content.WriteString("\n\n")

	// Calculate available height for sections
	usedHeight := 12 // header + status + dividers
	availableHeight := contentHeight - usedHeight

	optionsHeight := availableHeight / 2
	promptHeight := availableHeight - optionsHeight

	// Options Section
	content.WriteString(m.renderOptionsSection(contentWidth-4, optionsHeight))
	content.WriteString("\n")

	// Divider
	content.WriteString(m.renderDivider(contentWidth - 4))
	content.WriteString("\n")

	// Prompt Section
	content.WriteString(m.renderPromptSection(contentWidth-4, promptHeight))

	return mainStyle.Render(borderStyle.Render(content.String()))
}

func (m model) renderHeader(width int) string {
	logo := `
    ▗▄▖ ▗▄▄▖ ▗▄▄▖ ▗▄▄▄▖▗▄▄▖ 
   ▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌  █  ▐▌   
   ▐▌ ▐▌▐▛▀▚▖▐▛▀▚▖  █  ▐▌▝▜▌
   ▝▚▄▞▘▐▌ ▐▌▐▌ ▐▌▗▄█▄▖▝▚▄▞▘`

	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("87")).
		Bold(true)

	subtitle := "AI-Powered Coding Agent"
	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	version := "v1.0.0"
	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	centered := lipgloss.NewStyle().
		Width(width).
		AlignHorizontal(lipgloss.Center)

	return centered.Render(logoStyle.Render(logo)) + "\n" +
		centered.Render(subtitleStyle.Render(subtitle)) + " " +
		versionStyle.Render(version)
}

func (m model) renderStatusBar(width int) string {
	var indicator string
	if m.processing {
		indicators := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		indicator = indicators[m.animFrame%len(indicators)]
	} else {
		indicator = "●"
	}

	statusColor := lipgloss.Color("42")
	if m.processing {
		statusColor = lipgloss.Color("226")
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(statusColor).
		Bold(true)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("231"))

	leftSide := fmt.Sprintf("%s %s", statusStyle.Render(indicator), messageStyle.Render(m.statusMessage))

	rightSide := ""
	if m.activeSection == sectionOptions {
		rightSide = "OPTIONS MODE"
	} else {
		rightSide = "PROMPT MODE"
	}

	rightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)

	padding := width - lipgloss.Width(leftSide) - lipgloss.Width(rightSide)
	if padding < 0 {
		padding = 0
	}

	return leftSide + strings.Repeat(" ", padding) + rightStyle.Render(rightSide)
}

func (m model) renderOptionsSection(width, height int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Underline(m.activeSection == sectionOptions)

	content.WriteString(titleStyle.Render("⚙  CONFIGURATION"))
	content.WriteString("\n\n")

	options := []struct {
		label string
		value string
		arrow bool
	}{
		{"Project Type", m.projectType.String(), true},
		{"Directory", m.projectDir, false},
		{"LLM Model", m.llmModel.String(), true},
		{"Execute Build", "", false},
	}

	for i, opt := range options {
		selected := m.selectedOption == optionItem(i) && m.activeSection == sectionOptions

		var line strings.Builder
		line.WriteString("  ")

		if selected {
			line.WriteString("▶ ")
		} else {
			line.WriteString("  ")
		}

		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Width(15)
		line.WriteString(labelStyle.Render(opt.label))

		if optionItem(i) == optionExecute {
			execStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("231")).
				Background(lipgloss.Color("39")).
				Bold(true).
				Padding(0, 2)

			if selected {
				line.WriteString(execStyle.Render("[ START BUILD ]"))
			} else {
				dimStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					Padding(0, 2)
				line.WriteString(dimStyle.Render("[ START BUILD ]"))
			}
		} else {
			valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("231"))
			if selected && opt.arrow {
				line.WriteString("◀ ")
				line.WriteString(valueStyle.Bold(true).Render(opt.value))
				line.WriteString(" ▶")
			} else {
				line.WriteString(valueStyle.Render(opt.value))
			}
		}

		content.WriteString(line.String())
		content.WriteString("\n")
	}

	return content.String()
}

func (m model) renderPromptSection(width, height int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Underline(m.activeSection == sectionPrompt)

	content.WriteString(titleStyle.Render("✎  TASK PROMPT"))
	content.WriteString("\n\n")

	// Prompt input
	m.promptInput.Width = width - 4

	promptBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 1).
		Width(width - 2)

	if m.activeSection == sectionPrompt {
		promptBoxStyle = promptBoxStyle.BorderForeground(lipgloss.Color("39"))
	}

	content.WriteString(promptBoxStyle.Render(m.promptInput.View()))
	content.WriteString("\n\n")

	// Steps if processing or complete
	if len(m.steps) > 0 {
		stepsTitle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Render("Execution Log:")
		content.WriteString(stepsTitle)
		content.WriteString("\n\n")

		for _, step := range m.steps {
			var icon string
			var statusStyle lipgloss.Style

			switch step.status {
			case "complete":
				icon = "✓"
				statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			case "running":
				icon = "◉"
				statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
			case "error":
				icon = "✗"
				statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
			default:
				icon = "○"
				statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			}

			titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("231"))
			descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)

			content.WriteString(fmt.Sprintf("  %s %s\n", statusStyle.Render(icon), titleStyle.Render(step.title)))
			content.WriteString(fmt.Sprintf("    %s", descStyle.Render(step.description)))
			if step.timestamp != "" {
				content.WriteString(fmt.Sprintf(" %s", timeStyle.Render(step.timestamp)))
			}
			content.WriteString("\n")
		}
	}

	return content.String()
}

func (m model) renderDivider(width int) string {
	dividerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	return dividerStyle.Render(strings.Repeat("─", width))
}

func currentTime() string {
	return time.Now().Format("15:04:05")
}

// ============================================================================
// LLM Integration
// ============================================================================

type llmResponseMsg struct {
	response string
	err      error
}

func (m model) processWithLLM() tea.Cmd {
	return func() tea.Msg {
		// Simulate processing time
		time.Sleep(3 * time.Second)

		// TODO: Integrate with actual LLM API
		prompt := m.promptInput.Value()
		projectType := m.projectType.String()
		projectDir := m.projectDir
		llmModel := m.llmModel.String()

		response := callLLMAPI(prompt, projectType, projectDir, llmModel)

		return llmResponseMsg{
			response: response,
			err:      nil,
		}
	}
}

func callLLMAPI(prompt, projectType, projectDir, model string) string {
	// TODO: Replace with actual LLM API call
	//
	// Example OpenAI Integration:
	// import "github.com/sashabaranov/go-openai"
	// client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	// resp, err := client.CreateChatCompletion(
	//     context.Background(),
	//     openai.ChatCompletionRequest{
	//         Model: openai.GPT4,
	//         Messages: []openai.ChatCompletionMessage{
	//             {
	//                 Role:    openai.ChatMessageRoleSystem,
	//                 Content: "You are Orbis, an AI coding agent...",
	//             },
	//             {
	//                 Role:    openai.ChatMessageRoleUser,
	//                 Content: prompt,
	//             },
	//         },
	//     },
	// )
	//
	// Example Anthropic Claude Integration:
	// import "github.com/anthropics/anthropic-sdk-go"
	// client := anthropic.NewClient(os.Getenv("ANTHROPIC_API_KEY"))
	// message, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
	//     Model: anthropic.F(anthropic.ModelClaude3_5Sonnet20241022),
	//     Messages: anthropic.F([]anthropic.MessageParam{
	//         anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
	//     }),
	//     MaxTokens: anthropic.Int(1024),
	// })
	//
	// Example Local Ollama Integration:
	// import "github.com/ollama/ollama/api"
	// client, _ := api.ClientFromEnvironment()
	// req := &api.GenerateRequest{
	//     Model:  "codellama",
	//     Prompt: prompt,
	// }
	// client.Generate(context.Background(), req, func(resp api.GenerateResponse) error {
	//     return nil
	// })

	fmt.Printf("\n[LLM Request]\n")
	fmt.Printf("Model: %s\n", model)
	fmt.Printf("Project Type: %s\n", projectType)
	fmt.Printf("Directory: %s\n", projectDir)
	fmt.Printf("Prompt: %s\n", prompt)

	return "Success"
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
