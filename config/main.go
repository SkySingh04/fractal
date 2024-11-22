package config

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

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
func LoadConfig(configFile string) (map[string]string, error) {
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := map[string]string{
		"inputMethod":  viper.GetString("inputMethod"),
		"outputMethod": viper.GetString("outputMethod"),
	}

	return config, nil
}

// SetupConfigInteractively prompts the user to set up input and output methods interactively
func SetupConfigInteractively() (map[string]string, error) {
	// Prompt for Input Method
	inputPrompt := promptui.Select{
		Label: "Select Input Method",
		Items: []string{"CSV", "SQL Database", "Kafka Queue", "Cloud Storage"},
	}

	_, inputMethod, err := inputPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to get input method: %w", err)
	}

	// Prompt for Output Method
	outputPrompt := promptui.Select{
		Label: "Select Output Method",
		Items: []string{"SQL Database", "NoSQL Database", "CSV Output", "Kafka Queue", "Cloud Storage"},
	}

	_, outputMethod, err := outputPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to get output method: %w", err)
	}

	// Save user selections to the config
	config := map[string]string{
		"inputMethod":  inputMethod,
		"outputMethod": outputMethod,
	}

	// Optionally save the config to a file for future runs
	saveConfig(config)

	return config, nil
}

// saveConfig writes the configuration to a config.yaml file
func saveConfig(config map[string]string) {
	viper.Set("inputMethod", config["inputMethod"])
	viper.Set("outputMethod", config["outputMethod"])

	if err := viper.WriteConfigAs("config.yaml"); err != nil {
		fmt.Println("Failed to save configuration:", err)
	} else {
		fmt.Println("Configuration saved to config.yaml")
	}
}