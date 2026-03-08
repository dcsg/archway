package scaffold

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
)

func RunWizard(wizardConfig *WizardConfig, manifest *Manifest, initial map[string]interface{}) (map[string]interface{}, error) {
	if wizardConfig == nil {
		return nil, fmt.Errorf("wizard config is nil")
	}
	if manifest == nil {
		return nil, fmt.Errorf("manifest is nil")
	}

	state := manifest.Defaults()
	for k, v := range initial {
		state[k] = v
	}
	groups, err := buildWizardGroups(wizardConfig, manifest, state)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return state, nil
	}

	form := huh.NewForm(groups...)
	if err := form.Run(); err != nil {
		return nil, err
	}
	return state, nil
}

// RunProviderWizard runs the provider-level intent wizard and returns
// the selected template name plus any state collected.
func RunProviderWizard(cfg *ProviderWizardConfig, initial map[string]interface{}) (templateName string, state map[string]interface{}, err error) {
	state = map[string]interface{}{}
	for k, v := range initial {
		state[k] = v
	}

	// Run each step as its own form so that derived values (e.g. HasAPIOrWorker)
	// computed after step N are available as `when` conditions in step N+1.
	manifest := &Manifest{Name: "intent", Language: "any"}
	for _, step := range cfg.Steps {
		singleStepCfg := &WizardConfig{Steps: []WizardStep{step}}
		groups, err := buildWizardGroups(singleStepCfg, manifest, state)
		if err != nil {
			return "", nil, err
		}
		if len(groups) > 0 {
			form := huh.NewForm(groups...)
			if err := form.Run(); err != nil {
				return "", nil, err
			}
		}
		// Recompute derived booleans after each step so subsequent
		// steps can use them in `when` conditions.
		computeDerived(state)
	}

	templateName, err = cfg.ResolveTemplate(state)
	if err != nil {
		return "", nil, err
	}
	return templateName, state, nil
}

// RunWizardWithFastPath runs the template wizard, skipping steps listed in
// the fast path and applying fast-path defaults.
func RunWizardWithFastPath(wizardConfig *WizardConfig, manifest *Manifest, initial map[string]interface{}, fastPath *FastPath) (map[string]interface{}, error) {
	if wizardConfig == nil {
		return nil, fmt.Errorf("wizard config is nil")
	}
	if manifest == nil {
		return nil, fmt.Errorf("manifest is nil")
	}

	state := manifest.Defaults()
	for k, v := range initial {
		state[k] = v
	}

	// Apply fast-path defaults.
	if fastPath != nil {
		for k, v := range fastPath.Defaults {
			state[k] = v
		}
	}

	// Build skip set.
	skipSet := map[string]bool{}
	if fastPath != nil {
		for _, s := range fastPath.SkipSteps {
			skipSet[s] = true
		}
	}

	// Filter steps.
	filteredSteps := make([]WizardStep, 0, len(wizardConfig.Steps))
	for _, step := range wizardConfig.Steps {
		if !skipSet[step.ID] {
			filteredSteps = append(filteredSteps, step)
		}
	}
	filteredConfig := &WizardConfig{Steps: filteredSteps}

	groups, err := buildWizardGroups(filteredConfig, manifest, state)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return state, nil
	}
	form := huh.NewForm(groups...)
	if err := form.Run(); err != nil {
		return nil, err
	}
	return state, nil
}

// computeDerived computes boolean flags from multiselect ProjectCapabilities.
func computeDerived(state map[string]interface{}) {
	caps, _ := state["ProjectCapabilities"].([]string)
	hasAPI := sliceContains(caps, "api")
	hasWorker := sliceContains(caps, "worker")
	hasCLI := sliceContains(caps, "cli")
	state["HasAPI"] = hasAPI
	state["HasWorker"] = hasWorker
	state["HasCLI"] = hasCLI
	state["CLIOnly"] = hasCLI && !hasAPI && !hasWorker
	state["HasAPIOrWorker"] = hasAPI || hasWorker
}

func sliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func buildWizardGroups(cfg *WizardConfig, manifest *Manifest, state map[string]interface{}) ([]*huh.Group, error) {
	variableDefs := map[string]VariableDefinition{}
	for _, def := range manifest.Variables {
		variableDefs[def.Name] = def
	}

	groups := make([]*huh.Group, 0, len(cfg.Steps))
	for _, step := range cfg.Steps {
		fields := []huh.Field{}
		for _, q := range step.Questions {
			if !evaluateWhen(q.When, state) {
				continue
			}

			def := variableDefs[q.Variable]
			field, err := buildField(q, def, state)
			if err != nil {
				return nil, err
			}
			if field != nil {
				fields = append(fields, field)
			}
		}
		if len(fields) > 0 {
			groups = append(groups, huh.NewGroup(fields...))
		}
	}
	return groups, nil
}

func buildField(q WizardQuestion, def VariableDefinition, state map[string]interface{}) (huh.Field, error) {
	questionType := q.Type
	if questionType == "" {
		questionType = "input"
	}

	switch questionType {
	case "input":
		var value string
		if v, ok := state[q.Variable]; ok && v != nil {
			value = strings.TrimSpace(fmt.Sprint(v))
		}
		state[q.Variable] = value
		field := huh.NewInput().Title(q.Prompt).Value(&value)
		if q.Validate != "" {
			re, err := regexp.Compile(q.Validate)
			if err != nil {
				return nil, fmt.Errorf("invalid regex for %s: %w", q.Variable, err)
			}
			field.Validate(func(in string) error {
				if in == "" && !def.Required {
					return nil
				}
				if !re.MatchString(in) {
					return fmt.Errorf("value does not match %q", q.Validate)
				}
				return nil
			})
		}
		field.Validate(func(in string) error {
			if def.Required && strings.TrimSpace(in) == "" {
				return fmt.Errorf("%s is required", q.Variable)
			}
			state[q.Variable] = strings.TrimSpace(in)
			return nil
		})
		return field, nil
	case "confirm":
		value, _ := state[q.Variable].(bool)
		state[q.Variable] = value
		field := huh.NewConfirm().Title(q.Prompt).Value(&value)
		field.Validate(func(_ bool) error {
			state[q.Variable] = value
			return nil
		})
		return field, nil
	case "select":
		var selected string
		if v, ok := state[q.Variable]; ok && v != nil {
			selected = strings.TrimSpace(fmt.Sprint(v))
		}
		opts := flexOptionsForQuestion(q, def)
		if selected == "" && len(opts) > 0 {
			selected = opts[0].Value
		}
		state[q.Variable] = selected
		huhOpts := make([]huh.Option[string], 0, len(opts))
		for _, opt := range opts {
			huhOpts = append(huhOpts, huh.NewOption(opt.Label, opt.Value))
		}
		field := huh.NewSelect[string]().Title(q.Prompt).Options(huhOpts...).Value(&selected)
		field.Validate(func(_ string) error {
			state[q.Variable] = selected
			if def.Required && strings.TrimSpace(selected) == "" {
				return fmt.Errorf("%s is required", q.Variable)
			}
			return nil
		})
		return field, nil
	case "multiselect":
		values, _ := state[q.Variable].([]string)
		opts := flexOptionsForQuestion(q, def)
		huhOpts := make([]huh.Option[string], 0, len(opts))
		for _, opt := range opts {
			huhOpts = append(huhOpts, huh.NewOption(opt.Label, opt.Value))
		}
		field := huh.NewMultiSelect[string]().Title(q.Prompt).Options(huhOpts...).Value(&values)
		field.Validate(func(_ []string) error {
			state[q.Variable] = values
			if def.Required && len(values) == 0 {
				return fmt.Errorf("%s requires at least one value", q.Variable)
			}
			return nil
		})
		return field, nil
	default:
		return nil, fmt.Errorf("unsupported question type %q", questionType)
	}
}

// flexOptionsForQuestion returns FlexOption slice from question options or manifest choices.
func flexOptionsForQuestion(q WizardQuestion, def VariableDefinition) []FlexOption {
	if len(q.Options) > 0 {
		return q.Options
	}
	// Convert plain string choices from manifest to FlexOption.
	opts := make([]FlexOption, 0, len(def.Choices))
	for _, c := range def.Choices {
		opts = append(opts, FlexOption{Label: c, Value: c})
	}
	return opts
}

func evaluateWhen(expr string, values map[string]interface{}) bool {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return true
	}

	if strings.Contains(expr, "==") {
		parts := strings.SplitN(expr, "==", 2)
		left := strings.TrimSpace(parts[0])
		right := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		return strings.EqualFold(fmt.Sprint(values[left]), right)
	}
	if strings.Contains(expr, "!=") {
		parts := strings.SplitN(expr, "!=", 2)
		left := strings.TrimSpace(parts[0])
		right := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		return !strings.EqualFold(fmt.Sprint(values[left]), right)
	}

	value, exists := values[expr]
	if !exists {
		return false
	}
	if b, ok := value.(bool); ok {
		return b
	}
	return strings.TrimSpace(fmt.Sprint(value)) != ""
}
