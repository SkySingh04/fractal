package config

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/SkySingh04/fractal/logger"
	"github.com/SkySingh04/fractal/registry"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

// Config represents the entire configuration structure
type Config struct {
	InputMethod     string                 `yaml:"inputMethod"`
	OutputMethod    string                 `yaml:"outputMethod"`
	InputConfig     map[string]interface{} `yaml:"inputconfig"`
	OutputConfig    map[string]interface{} `yaml:"outputconfig"`
	Validations     []string               `yaml:"validations"`
	Transformations []string               `yaml:"transformations"`
	ErrorHandling   ErrorHandling          `yaml:"errorhandling"`
}

// ErrorHandling represents the error handling configuration
type ErrorHandling struct {
	Strategy         string           `yaml:"strategy"`
	QuarantineOutput QuarantineOutput `yaml:"quarantineoutput"`
}

// QuarantineOutput represents the quarantine output configuration
type QuarantineOutput struct {
	Type     string `yaml:"type"`
	Location string `yaml:"location"`
}

// AskForMode prompts the user to select between starting the HTTP server or using the CLI
func AskForMode() (string, error) {
	modePrompt := promptui.Select{
		Label: "Choose Application Mode",
		Items: []string{"Start HTTP Server", "Use CLI"},
	}

	_, mode, err := modePrompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to select application mode: %w", err)
	}

	return mode, nil
}

// LoadConfig attempts to read the configuration from a file
func LoadConfig(configFile string) (map[string]interface{}, error) {
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := map[string]interface{}{
		"inputMethod":     viper.GetString("inputMethod"),
		"outputMethod":    viper.GetString("outputMethod"),
		"inputconfig":     viper.GetStringMap("inputconfig"),
		"outputconfig":    viper.GetStringMap("outputconfig"),
		"errorhandling":   viper.GetStringMap("errorhandling"), // Keep this as a map if it contains structured data
		"validations":     viper.GetString("validations"),      // Changed to GetString
		"transformations": viper.GetString("transformations"),  // Changed to GetString
	}

	logger.Infof("Configuration loaded from %s", configFile)
	return config, nil
}

// SetupConfigInteractively prompts the user to set up input and output methods interactively,
// including all required fields for the selected integrations.
func SetupConfigInteractively() (map[string]interface{}, error) {
	// Dynamically retrieve registered input and output options
	inputMethods := getRegisteredDataSources()
	outputMethods := getRegisteredDataDestinations()

	// Prompt for Input Method
	inputPrompt := promptui.Select{
		Label: "Select Input Method",
		Items: inputMethods,
	}
	_, inputMethod, err := inputPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to get input method: %w", err)
	}

	// Read additional fields for the input method
	inputconfig, err := readIntegrationFields(inputMethod, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get fields for input method: %w", err)
	}

	// Prompt for Output Method
	outputPrompt := promptui.Select{
		Label: "Select Output Method",
		Items: outputMethods,
	}
	_, outputMethod, err := outputPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to get output method: %w", err)
	}

	// Read additional fields for the output method
	outputconfig, err := readIntegrationFields(outputMethod, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get fields for output method: %w", err)
	}

	// Read validations and transformations
	validations, err := readRules("validations")
	if err != nil {
		return nil, fmt.Errorf("failed to read validation rules: %w", err)
	}
	transformations, err := readRules("transformations")
	if err != nil {
		return nil, fmt.Errorf("failed to read transformation rules: %w", err)
	}

	// Read error handling
	errorhandling, err := readErrorHandlingConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read error handling configuration: %w", err)
	}

	// Combine all configurations
	config := map[string]interface{}{
		"inputMethod":     inputMethod,
		"outputMethod":    outputMethod,
		"inputconfig":     inputconfig,
		"outputconfig":    outputconfig,
		"validations":     validations,
		"transformations": transformations,
		"errorhandling":   errorhandling,
	}
	saveConfig(config)

	return config, nil
}

// readIntegrationFields dynamically prompts for and reads all fields in the selected integration struct
func readIntegrationFields(method string, isSource bool) (map[string]interface{}, error) {
	var integration interface{}
	var found bool

	// Get the appropriate integration
	if isSource {
		integration, found = registry.GetSource(method)
	} else {
		integration, found = registry.GetDestination(method)
	}

	if !found {
		return nil, errors.New("integration not found in registry")
	}

	// Use reflection to inspect the integration struct
	val := reflect.ValueOf(integration)
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // Dereference if it's a pointer
	}
	if val.Kind() != reflect.Struct {
		return nil, errors.New("integration is not a struct")
	}

	config := make(map[string]interface{})
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fieldName := field.Name
		fieldType := field.Type

		// Prompt the user for the field value
		prompt := promptui.Prompt{
			Label: fmt.Sprintf("Enter %s (%s)", fieldName, fieldType),
		}
		value, err := prompt.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to get value for field %s: %w", fieldName, err)
		}

		// Assign the value to the config
		config[fieldName] = value
	}

	return config, nil
}

// readRules reads validation or transformation rules interactively
func readRules(ruleType string) (string, error) {
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("Enter %s rules (multiline, finish with empty line):", ruleType),
	}
	rules := ""
	for {
		line, err := prompt.Run()
		if err != nil {
			return "", fmt.Errorf("failed to read %s: %w", ruleType, err)
		}
		if line == "" {
			break
		}
		rules += line + "\n"
	}
	return rules, nil
}

// readErrorHandlingConfig prompts for error handling strategy and quarantine details
func readErrorHandlingConfig() (map[string]interface{}, error) {
	prompt := promptui.Prompt{
		Label: "Enter Error Handling Strategy (e.g., LOG_AND_CONTINUE, STOP_ON_ERROR):",
	}
	strategy, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to read error handling strategy: %w", err)
	}

	return map[string]interface{}{
		"strategy": strategy,
	}, nil
}

// saveConfig writes the configuration to a config.yaml file
func saveConfig(config map[string]interface{}) {
	for key, value := range config {
		viper.Set(key, value)
	}

	if err := viper.WriteConfigAs("config.yaml"); err != nil {
		fmt.Println("Failed to save configuration:", err)
	} else {
		fmt.Println("Configuration saved to config.yaml")
	}
}

// Helper function to retrieve registered input methods
func getRegisteredDataSources() []string {
	var sources []string
	for source := range registry.GetSources() {
		sources = append(sources, source)
	}
	return sources
}

// Helper function to retrieve registered output methods
func getRegisteredDataDestinations() []string {
	var destinations []string
	for dest := range registry.GetDestinations() {
		destinations = append(destinations, dest)
	}
	return destinations
}
