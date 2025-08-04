package command

import (
	"fmt"
	"os"
	"strings"
)

// WizardStep defines a single step in the wizard.
type WizardStep struct {
	ID       string
	Type     StepType
	Question string
	Options  []Option
	Required bool
	Default  any // For default values.

	// Conditional execution.
	Condition func(results WizardResults) bool

	// Custom validation.
	Validate func(value any) error

	// Transform/modify result before storing.
	Transform func(value any) any

	// Process function for long-running operations.
	Process func(results WizardResults, progress func(string)) (any, error)
}

// StepType defines the type of wizard step.
type StepType int

const (
	StepChoice StepType = iota
	StepMultiChoice
	StepPrompt
	StepConfirm
	StepInfo    // For displaying information only.
	StepProcess // For long-running operations.
)

// WizardResults holds all results from wizard steps.
type WizardResults map[string]any

// WizardConfig configuration for wizard.
type WizardConfig struct {
	Title         string
	Description   string
	ShowProgress  bool
	ClearScreen   bool
	ClearHistory  bool // Clear terminal scroll history.
	ResultColor   ResultColor
	ShowFinish    bool   // Show finish screen.
	FinishMessage string // Custom finish message.
}

// CommandWizard extends command with wizard functionality.
type CommandWizard struct {
	*command
	config  *WizardConfig
	steps   []WizardStep
	results WizardResults
}

var defaultWizardConfig = WizardConfig{
	ShowProgress:  true,
	ClearScreen:   true,
	ClearHistory:  true,
	ResultColor:   ColorDefault,
	ShowFinish:    true,
	FinishMessage: "✅ Setup completed successfully!",
}

// NewCommandWizard creates a new command wizard instance.
func NewCommandWizard(config ...any) *CommandWizard {
	var cmdConfig Config
	var wizardConfig WizardConfig

	// Parse configs.
	for _, cfg := range config {
		switch v := cfg.(type) {
		case Config:
			cmdConfig = v
		case WizardConfig:
			wizardConfig = v
		}
	}

	// Default wizard config.
	if wizardConfig == (WizardConfig{}) {
		wizardConfig = defaultWizardConfig
	}

	cmd := NewCommand(cmdConfig)

	return &CommandWizard{
		command: cmd,
		config:  &wizardConfig,
		steps:   []WizardStep{},
		results: make(WizardResults),
	}
}

// AddStep adds a step to the wizard.
func (w *CommandWizard) AddStep(step WizardStep) *CommandWizard {
	w.steps = append(w.steps, step)
	return w
}

// AddChoice adds a choice step.
func (w *CommandWizard) AddChoice(id, question string, options []Option, required ...bool) *CommandWizard {
	req := true
	if len(required) > 0 {
		req = required[0]
	}

	return w.AddStep(WizardStep{
		ID:       id,
		Type:     StepChoice,
		Question: question,
		Options:  options,
		Required: req,
	})
}

// AddMultiChoice adds a multi-choice step.
func (w *CommandWizard) AddMultiChoice(id, question string, options []Option, required ...bool) *CommandWizard {
	req := true
	if len(required) > 0 {
		req = required[0]
	}

	return w.AddStep(WizardStep{
		ID:       id,
		Type:     StepMultiChoice,
		Question: question,
		Options:  options,
		Required: req,
	})
}

// AddPrompt adds a prompt step.
func (w *CommandWizard) AddPrompt(id, question string, required ...bool) *CommandWizard {
	req := true
	if len(required) > 0 {
		req = required[0]
	}

	return w.AddStep(WizardStep{
		ID:       id,
		Type:     StepPrompt,
		Question: question,
		Required: req,
	})
}

// AddPromptWithDefault adds a prompt with default value step.
func (w *CommandWizard) AddPromptWithDefault(id, question string, defaultValue string) *CommandWizard {
	return w.AddStep(WizardStep{
		ID:       id,
		Type:     StepPrompt,
		Question: question,
		Default:  defaultValue,
		Required: false,
	})
}

// AddConfirm adds a confirm step.
func (w *CommandWizard) AddConfirm(id, question string, defaultValue ...bool) *CommandWizard {
	def := true
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}

	return w.AddStep(WizardStep{
		ID:       id,
		Type:     StepConfirm,
		Question: question,
		Default:  def,
		Required: false,
	})
}

// AddInfo adds an info step (display only).
func (w *CommandWizard) AddInfo(id, message string) *CommandWizard {
	return w.AddStep(WizardStep{
		ID:       id,
		Type:     StepInfo,
		Question: message,
		Required: false,
	})
}

// AddConditionalStep adds a step with condition.
func (w *CommandWizard) AddConditionalStep(step WizardStep, condition func(results WizardResults) bool) *CommandWizard {
	step.Condition = condition
	return w.AddStep(step)
}

// AddProcess adds a processing step for long-running operations.
func (w *CommandWizard) AddProcess(
	id, initialMessage string,
	processFn func(results WizardResults, progress func(string)) (any, error),
) *CommandWizard {
	return w.AddStep(WizardStep{
		ID:       id,
		Type:     StepProcess,
		Question: initialMessage,
		Process:  processFn,
		Required: false,
	})
}

// AddConditionalProcess adds a conditional processing step.
func (w *CommandWizard) AddConditionalProcess(
	id, initialMessage string,
	processFn func(results WizardResults, progress func(string)) (any, error),
	condition func(WizardResults) bool,
) *CommandWizard {
	return w.AddConditionalStep(WizardStep{
		ID:       id,
		Type:     StepProcess,
		Question: initialMessage,
		Process:  processFn,
		Required: false,
	}, condition)
}

// Run executes the wizard.
func (w *CommandWizard) Run() (WizardResults, error) {
	if w.config.ClearScreen {
		if w.config.ClearHistory {
			clearTerminalHistory()
		} else {
			clearScreen()
		}
	}

	// Display title and description.
	if w.config.Title != "" {
		fmt.Fprintf(w.writer, "%s\n", w.config.Title)
		if w.config.Description != "" {
			fmt.Fprintf(w.writer, "%s\n", w.config.Description)
		}
		fmt.Fprintln(w.writer)
	}

	totalSteps := len(w.steps)
	currentStep := 0

	for _, step := range w.steps {
		currentStep++

		// Check condition if exists.
		if step.Condition != nil && !step.Condition(w.results) {
			continue
		}

		// Show progress if enabled.
		if w.config.ShowProgress && totalSteps > 1 {
			progress := fmt.Sprintf("(%d/%d)", currentStep, totalSteps)
			fmt.Fprintf(w.writer, "%s ", progress)
		}

		// Execute step based on type.
		var result any
		var err error

		switch step.Type {
		case StepChoice:
			output, stepErr := w.command.Choice(step.Question, step.Options, step.Required)
			if stepErr != nil {
				return w.results, stepErr
			}
			result = output.Value()

		case StepMultiChoice:
			output, stepErr := w.command.MultiChoice(step.Question, step.Options, step.Required)
			if stepErr != nil {
				return w.results, stepErr
			}
			result = output.Value()

		case StepPrompt:
			var output *Output[string]
			var stepErr error

			if step.Default != nil {
				if defaultVal, ok := step.Default.(string); ok {
					output, stepErr = w.command.PromptWithDefault(step.Question, defaultVal)
				} else {
					output, stepErr = w.command.Prompt(step.Question, step.Required)
				}
			} else {
				output, stepErr = w.command.Prompt(step.Question, step.Required)
			}

			if stepErr != nil {
				return w.results, stepErr
			}

			result = strings.TrimSpace(output.Value())

		case StepConfirm:
			defaultVal := true
			if step.Default != nil {
				if v, ok := step.Default.(bool); ok {
					defaultVal = v
				}
			}
			output, stepErr := w.command.ConfirmWithDefault(step.Question, defaultVal)
			if stepErr != nil {
				return w.results, stepErr
			}
			result = output.Value()

		case StepInfo:
			if w.config.ClearScreen {
				if w.config.ClearHistory {
					clearTerminalHistory()
				} else {
					clearScreen()
				}
			}
			fmt.Fprintf(w.writer, "◆ %s\n", step.Question)
			fmt.Fprintln(w.writer)
			continue // Info steps don't store results.

		case StepProcess:
			if w.config.ClearScreen {
				if w.config.ClearHistory {
					clearTerminalHistory()
				} else {
					clearScreen()
				}
				w.showPreviousResults()
			}

			// Show initial message.
			fmt.Fprintf(w.writer, "◆ %s\n", step.Question)

			// Progress callback function.
			progress := func(message string) {
				if w.config.ClearScreen {
					if w.config.ClearHistory {
						clearTerminalHistory()
					} else {
						clearScreen()
					}
					w.showPreviousResults()
				}
				fmt.Fprintf(w.writer, "◆ %s\n", message)
				os.Stdout.Sync() // Force flush for real-time updates.
			}

			// Execute process function dan langsung store.
			if step.Process != nil {
				processResult, processErr := step.Process(w.results, progress)
				if processErr != nil {
					// Show error message.
					progress(fmt.Sprintf("❌ Error: %v", processErr))
					return w.results, processErr // Stop wizard execution.
				}

				// Store successful result.
				w.results[step.ID] = processResult
			}

			continue
		}

		// Custom validation if exists.
		if step.Validate != nil {
			if err = step.Validate(result); err != nil {
				fmt.Fprintf(w.writer, "Error: %v\n", err)
				currentStep-- // Retry the same step.
				continue
			}
		}

		// Transform result if function exists.
		if step.Transform != nil {
			result = step.Transform(result)
		}

		// Store result.
		w.results[step.ID] = result

		// Show confirmation of selection in wizard style.
		w.showStepResult()
	}

	// Final clear screen untuk menghindari sisa prompt.
	if w.config.ClearScreen {
		if w.config.ClearHistory {
			clearTerminalHistory()
		} else {
			clearScreen()
		}

		// Show final summary with all completed steps.
		if w.config.Title != "" {
			fmt.Fprintf(w.writer, "%s\n", w.config.Title)
			if w.config.Description != "" {
				fmt.Fprintf(w.writer, "%s\n", w.config.Description)
			}
			fmt.Fprintln(w.writer)
		}

		// Show all results.
		for _, step := range w.steps {
			if result, exists := w.results[step.ID]; exists {
				w.formatStepResult(step, result, true)
			}
		}
	}

	// Show finish screen if enabled.
	if w.config.ShowFinish {
		fmt.Fprintln(w.writer)
		fmt.Fprintf(w.writer, "%s\n", w.config.FinishMessage)
	}
	return w.results, nil
}

// showStepResult displays the result in wizard style with checkmark.
func (w *CommandWizard) showStepResult() {
	if w.config.ClearScreen {
		if w.config.ClearHistory {
			clearTerminalHistory()
		} else {
			clearScreen()
		}

		// Redisplay title.
		if w.config.Title != "" {
			fmt.Fprintf(w.writer, "%s\n", w.config.Title)
			if w.config.Description != "" {
				fmt.Fprintf(w.writer, "%s\n", w.config.Description)
			}
			fmt.Fprintln(w.writer)
		}

		// Show all previous results with checkmarks.
		for _, prevStep := range w.steps {
			if prevResult, exists := w.results[prevStep.ID]; exists {
				w.formatStepResult(prevStep, prevResult, true)
			}
		}
	}
}

// formatStepResult formats a step result for display.
func (w *CommandWizard) formatStepResult(step WizardStep, result any, completed bool) {
	icon := "◆"
	if completed {
		icon = "\033[92m◇\033[0m" // Different icon for completed steps.
	}

	question := strings.TrimSuffix(step.Question, w.command.config.QuestionMark)
	question = strings.TrimSpace(question)

	var resultText string

	switch v := result.(type) {
	case Option:
		resultText = v.Value
	case []Option:
		if len(v) == 0 {
			resultText = "None"
		} else {
			var values []string
			for _, opt := range v {
				values = append(values, opt.Value)
			}
			resultText = strings.Join(values, ", ")
		}
	case bool:
		if v {
			resultText = "Yes"
		} else {
			resultText = "No"
		}
	case string:
		resultText = v
	default:
		resultText = fmt.Sprintf("%v", v)
	}

	fmt.Fprintf(w.writer, "%s %s\n", icon, question)
	if completed && resultText != "" {
		colorCode := w.config.ResultColor.getColorCode()
		fmt.Fprintf(w.writer, "│ %s%s\033[0m\n", colorCode, resultText)
		fmt.Fprintln(w.writer, "│")
	}
}

// showPreviousResults displays all completed steps.
func (w *CommandWizard) showPreviousResults() {
	// Redisplay title.
	if w.config.Title != "" {
		fmt.Fprintf(w.writer, "%s\n", w.config.Title)
		if w.config.Description != "" {
			fmt.Fprintf(w.writer, "%s\n", w.config.Description)
		}
		fmt.Fprintln(w.writer)
	}

	// Show all previous results with checkmarks.
	for _, prevStep := range w.steps {
		if prevResult, exists := w.results[prevStep.ID]; exists {
			w.formatStepResult(prevStep, prevResult, true)
		}
	}
}

// GetResult gets a specific result by ID.
func (w *CommandWizard) GetResult(id string) (any, bool) {
	result, exists := w.results[id]
	return result, exists
}

// GetStringResult gets a string result by ID.
func (w *CommandWizard) GetStringResult(id string) (string, bool) {
	if result, exists := w.results[id]; exists {
		if str, ok := result.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetBoolResult gets a boolean result by ID.
func (w *CommandWizard) GetBoolResult(id string) (bool, bool) {
	if result, exists := w.results[id]; exists {
		if b, ok := result.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// GetOptionResult gets an Option result by ID.
func (w *CommandWizard) GetOptionResult(id string) (Option, bool) {
	if result, exists := w.results[id]; exists {
		if opt, ok := result.(Option); ok {
			return opt, true
		}
	}
	return Option{}, false
}

// GetMultiOptionResult gets a []Option result by ID.
func (w *CommandWizard) GetMultiOptionResult(id string) ([]Option, bool) {
	if result, exists := w.results[id]; exists {
		if opts, ok := result.([]Option); ok {
			return opts, true
		}
	}
	return []Option{}, false
}

// PrintSummary prints a summary of all wizard results.
func (w *CommandWizard) PrintSummary() {
	fmt.Fprintln(w.writer)
	fmt.Fprintln(w.writer, "=== Wizard Summary ===")

	for _, step := range w.steps {
		if result, exists := w.results[step.ID]; exists {
			question := strings.TrimSuffix(step.Question, w.command.config.QuestionMark)
			question = strings.TrimSpace(question)

			fmt.Fprintf(w.writer, "%s: ", question)

			switch v := result.(type) {
			case Option:
				fmt.Fprintf(w.writer, "%s\n", v.Value)
			case []Option:
				var values []string
				for _, opt := range v {
					values = append(values, opt.Value)
				}
				fmt.Fprintf(w.writer, "%s\n", strings.Join(values, ", "))
			case bool:
				if v {
					fmt.Fprintln(w.writer, "Yes")
				} else {
					fmt.Fprintln(w.writer, "No")
				}
			default:
				fmt.Fprintf(w.writer, "%v\n", v)
			}
		}
	}
}
