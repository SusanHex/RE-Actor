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
	ActionName string `mapstructure:"action_name"`
	LogLevel   string `mapstructure:"log_level"`
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
	Name        string `mapstructure:"container_name"`
	ID          string
	Enabled     bool
	Action      *actions.Action
	PatternInfo PatternConfig
}

type PatternConfig struct {
	Pattern    *regexp.Regexp
	Template   string `mapstructure:"template" json:"template"`
	RawPattern string `mapstructure:"pattern" json:"pattern"`
}

func GetConfig(viper_instance *viper.Viper) (*Config, error) {
	app_config := Config{}

	viper_instance.BindEnv("action_name")
	viper_instance.SetDefault("log_level", "INFO")
	viper_instance.SetDefault("action_name", "discord_webhook")
	viper_instance.BindEnv("test_action_delay")
	viper_instance.SetDefault("test_action_delay", 0)
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

	var err error
	app_config.Containers, err = fetchContainerConfigFromEnv(viper_instance)
	if err != nil {
		return &app_config, err
	}

	// Get all the Discord Webhook URLs
	raw_discord_webhook_urls := viper_instance.GetString("discord_webhook_urls")
	discord_webhook_url_separator := viper_instance.GetString("discord_webhook_url_separator")
	app_config.DiscordWebHookURLs = strings.Split(raw_discord_webhook_urls, discord_webhook_url_separator)

	err = viper_instance.Unmarshal(&app_config)
	if err != nil {
		return nil, err
	}
	return &app_config, nil
}

func getConfigFileFromURL(url string) (PatternConfig, error) {
	pattern_config := PatternConfig{}

	resp, err := http.Get(url)
	if err != nil {
		return pattern_config, err
	}

	body_content, err := io.ReadAll(resp.Body)
	if err != nil {
		return pattern_config, err
	}

	if resp.StatusCode != 200 {
		return pattern_config, fmt.Errorf("Request to \"%s\" status: %d, body: \"%s\"", url, resp.StatusCode, body_content)
	}

	err = json.Unmarshal(body_content, pattern_config)
	if err != nil {
		return pattern_config, err
	}

	pattern_config.Pattern, err = regexp.Compile(pattern_config.RawPattern)
	if err != nil {
		return pattern_config, err
	}

	return pattern_config, nil
}

// TODO: Adjust this function to be able to fetch the config using the template name

func FetchContainerConfigsFromLabels(container_summaries []container.Summary) []ContainerConfig {
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
			PatternInfo: PatternConfig{Pattern: compiled_pattern, Template: template, RawPattern: raw_pattern},
		})
	}
	return container_configs
}

func fetchContainerConfigFromEnv(viper_instance *viper.Viper) ([]ContainerConfig, error) {
	container_config := ContainerConfig{Name: "", ID: "", Enabled: false, PatternInfo: PatternConfig{}}

	viper_instance.BindEnv("template_name")
	viper_instance.SetDefault("template_name", "")
	template_name := viper_instance.GetString("template_name")
	if len(template_name) > 0 {
		template_url := fmt.Sprintf("https://raw.githubusercontent.com/SusanHex/RE-Actor-Templates/refs/heads/production/templates/%s.json", template_name)
		remote_pattern_config, config_err := getConfigFileFromURL(template_url)
		if config_err != nil {
			return nil, config_err
		}
		container_config.PatternInfo = remote_pattern_config
	}

	viper_instance.BindEnv("container_name")
	err := viper_instance.Unmarshal(&container_config)
	if err != nil {
		return nil, err
	}

	if len(container_config.Name) == 0 {
		return make([]ContainerConfig, 0), nil
	}

	viper_instance.BindEnv("pattern")
	raw_pattern := viper_instance.GetString("pattern")
	if len(raw_pattern) > 0 {
		compiled_pattern, err := regexp.Compile(raw_pattern)
		if err != nil {
			return nil, err
		}
		container_config.PatternInfo.Pattern = compiled_pattern
		container_config.PatternInfo.RawPattern = raw_pattern
	}

	viper_instance.BindEnv("template")
	template := viper_instance.GetString("pattern")
	if len(template) > 0 {
		container_config.PatternInfo.Template = template
	}

	configs := make([]ContainerConfig, 0)
	configs = append(configs, container_config)
	return configs, nil
}
