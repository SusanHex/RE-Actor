package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/SusanHex/RE-Actor/src/actions"
	"github.com/docker/docker/api/types/container"
	"github.com/spf13/viper"
)

type Config struct {
	ContainerName   string `mapstructure:"container_name"`
	Pattern         string `mapstructure:"pattern" json:"pattern"`
	Template        string `mapstructure:"template" json:"template"`
	ActionName      string `mapstructure:"action_name"`
	CompiledPattern *regexp.Regexp
	LogLevel        string `mapstructure:"log_level"`
	// Discord Webhook Option
	DiscordWebHookURLs []string
	// SMTP Config Options
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     string `mapstructure:"smtp_port"`
	SMTPSendFrom string `mapstructure:"smtp_send_from"`
	SMTPSendTo   string `mapstructure:"smtp_send_to"`
	SMTPSubject  string `mapstructure:"smtp_subject"`
	SMTPPassword string `mapstructure:"smtp_password"`
	// Test Action Option
	TestActionDelay uint `mapstructure:"test_action_delay"`
	Containers      []ContainerConfig
}

type ContainerConfig struct {
	Name        string
	ID          string
	Enabled     bool
	Action      *actions.Action
	PatternInfo PatternConfig
}

type PatternConfig struct {
	Pattern  *regexp.Regexp
	Template string
}

func GetConfigFromViper(viper_instance *viper.Viper) (*Config, error) {
	app_config := Config{}

	viper_instance.BindEnv("pattern")
	viper_instance.BindEnv("template")
	viper_instance.BindEnv("action_name")
	viper_instance.SetDefault("log_level", "INFO")
	viper_instance.SetDefault("action_name", "discord_webhook")
	viper_instance.BindEnv("test_action_delay")
	viper_instance.SetDefault("test_action_delay", 0)
	viper_instance.BindEnv("container_name")
	viper_instance.BindEnv("discord_webhook_urls")
	viper_instance.BindEnv("discord_webhook_url_separator")
	viper_instance.SetDefault("discord_webhook_url_separator", ";;;")
	viper_instance.BindEnv("log_level")
	viper_instance.BindEnv("smtp_host")
	viper_instance.BindEnv("smtp_port")
	viper_instance.BindEnv("smtp_send_from")
	viper_instance.BindEnv("smtp_send_to")
	viper_instance.BindEnv("smtp_subject")
	viper_instance.BindEnv("smtp_password")
	viper_instance.BindEnv("template_name")
	viper_instance.SetDefault("template_name", "")
	viper_instance.AutomaticEnv()

	// check for tempate name
	template_name := viper_instance.GetString("template_name")
	if len(template_name) > 0 {
		template_url := fmt.Sprintf("https://raw.githubusercontent.com/SusanHex/RE-Actor-Templates/refs/heads/production/templates/%s.json", template_name)
		config_err := getConfigFileFromURL(template_url, &app_config)
		if config_err != nil {
			return nil, config_err
		}
	}

	// Get all the Discord Webhook URLs
	raw_discord_webhook_urls := viper_instance.GetString("discord_webhook_urls")
	discord_webhook_url_separator := viper_instance.GetString("discord_webhook_url_separator")
	app_config.DiscordWebHookURLs = strings.Split(raw_discord_webhook_urls, discord_webhook_url_separator)

	err := viper_instance.Unmarshal(&app_config)
	if err != nil {
		return nil, err
	}

	if len(app_config.ContainerName) == 0 {
		return nil, fmt.Errorf("no container name supplied")
	}
	if len(app_config.Pattern) == 0 {
		return nil, fmt.Errorf("no pattern supplied")
	}
	if len(app_config.Template) == 0 {
		return nil, fmt.Errorf("no template supplied")
	}
	compiled_pattern, err := regexp.Compile(app_config.Pattern)
	if err != nil {
		return nil, err
	}
	app_config.CompiledPattern = compiled_pattern
	return &app_config, nil
}

func getConfigFileFromURL(url string, app_config *Config) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	body_content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Request to \"%s\" status: %d, body: \"%s\"", url, resp.StatusCode, body_content)
	}
	err = json.Unmarshal(body_content, app_config)
	if err != nil {
		return err
	}
	return nil
}

// TODO: Write a function to create `ContainerConfig` instances from the Docker labels

func fetchContainerConfigsFromLabels(container_summaries []container.Summary) []ContainerConfig {
	container_configs := make([]ContainerConfig, 0)
	for _, container_summary := range container_summaries {
		reactor_enabled, ok := container_summary.Labels["reactor.enabled"]
		if !ok || reactor_enabled != "true" {
			continue
		}

		raw_pattern, has_pattern := container_summary.Labels["reactor.pattern"]
		if has_pattern && len(raw_pattern) == 0 {
			has_pattern = false
		}
		if !has_pattern {
			continue
		}
		compiled_pattern, err := regexp.Compile(raw_pattern)
		if err != nil {
			slog.Error(fmt.Sprintf(`failed to compile pattern for "%v"`, container_summary.Names[0]))
			continue
		}

		template, has_template := container_summary.Labels["reactor.template"]
		if !has_template || len(template) == 0 {
			template = "\\0"
		}

		container_configs = append(container_configs, ContainerConfig{
			Name:        container_summary.Names[0],
			ID:          container_summary.ID,
			Enabled:     true,
			Action:      nil,
			PatternInfo: PatternConfig{Pattern: compiled_pattern, Template: template},
		})
	}
	return container_configs
}
