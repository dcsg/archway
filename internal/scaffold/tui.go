package scaffold

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
)

func RunWizard(wizardConfig *WizardConfig, manifest *Manifest) (map[string]interface{}, error) {
	if wizardConfig == nil {
		return nil, fmt.Errorf("wizard config is nil")
	}
	if manifest == nil {
		return nil, fmt.Errorf("manifest is nil")
	}

	state := manifest.Defaults()
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
		value := strings.TrimSpace(fmt.Sprint(state[q.Variable]))
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
		selected := strings.TrimSpace(fmt.Sprint(state[q.Variable]))
		opts := optionsForQuestion(q, def)
		if selected == "" && len(opts) > 0 {
			selected = opts[0]
		}
		state[q.Variable] = selected
		huhOpts := make([]huh.Option[string], 0, len(opts))
		for _, opt := range opts {
			huhOpts = append(huhOpts, huh.NewOption(opt, opt))
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
		opts := optionsForQuestion(q, def)
		huhOpts := make([]huh.Option[string], 0, len(opts))
		for _, opt := range opts {
			huhOpts = append(huhOpts, huh.NewOption(opt, opt))
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

func optionsForQuestion(q WizardQuestion, def VariableDefinition) []string {
	if len(q.Options) > 0 {
		return q.Options
	}
	return def.Choices
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
