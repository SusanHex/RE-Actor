package utils

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/SusanHex/RE-Actor/src/actions"
	"github.com/SusanHex/RE-Actor/src/config"
)

func SelectAction(action_name string, app_config *config.Config) (actions.Action, error) {

	switch action_name {
	case "discord_webhook":
		return actions.DiscordWebHook{URLs: app_config.DiscordWebHookURLs}, nil
	case "smtp":
		return actions.SMTPMail{
			SMTPHost: app_config.SMTPHost,
			SMTPPort: app_config.SMTPPort,
			SendFrom: app_config.SMTPSendFrom,
			SendTo:   app_config.SMTPSendTo,
			Subject:  app_config.SMTPSubject,
			Password: app_config.SMTPPassword,
		}, nil
	case "test":
		return actions.TestAction{Delay: app_config.TestActionDelay}, nil
	default:
		return nil, fmt.Errorf(`"%s" does not match an action`, action_name)
	}
}

func PerformActionIfMatch(app_config *config.Config, action actions.Action, message []byte) error {
	match_indexes := app_config.CompiledPattern.FindSubmatchIndex(message)
	if len(match_indexes) == 0 {
		return nil
	}
	result := []byte{}
	result = app_config.CompiledPattern.Expand(result, []byte(app_config.Template), message, match_indexes)
	text_result := string(result)
	slog.Info(fmt.Sprintf(`Acting on "%s"`, text_result))
	err := action.Act(text_result)
	return err
}

func SetupLogger(app_config *config.Config) error {
	var log_level slog.Level
	add_source := false
	switch strings.ToUpper(app_config.LogLevel) {
	case "DEBUG":
		log_level = slog.LevelDebug
		add_source = true
	case "INFO":
		log_level = slog.LevelInfo
	case "WARNING":
		log_level = slog.LevelWarn
	case "ERROR":
		log_level = slog.LevelError
	default:
		return fmt.Errorf(`log level of "%s" does not match one of the following: DEBUG, INFO, WARNING, or ERROR`, app_config.LogLevel)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: add_source, Level: log_level}))
	slog.SetDefault(logger)
	return nil
}

func GetContainerLog(log_reader io.Reader, is_tty bool) ([]byte, error) {
	if is_tty {
		buf_reader := bufio.NewReader(log_reader)
		raw_log_message, err := buf_reader.ReadString('\n')
		if err != nil {
			return []byte{}, err
		}
		return []byte(strings.TrimSpace(raw_log_message)), nil
	}
	header := make([]byte, 8)
	head_bytes_read, err := log_reader.Read(header)
	if err != nil {
		return []byte{}, err
	}
	if head_bytes_read < 8 {
		return []byte{}, nil
	}
	slog.Debug(string(header))
	log_message_length := binary.BigEndian.Uint32(header[4:])
	log_message := make([]byte, log_message_length)
	_, err = log_reader.Read(log_message)
	if err != nil {
		return []byte{}, err
	}
	return log_message, nil

}
