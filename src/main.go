/*
Copyright © 2025 Susan Hex
*/
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/SusanHex/RE-Actor/src/config"
	"github.com/SusanHex/RE-Actor/src/utils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/spf13/viper"
)

// TODO: Move all docker logic to a new "provider" type. Ideally, the provider system should make implementing new providers easier, with needing large rewrites.
// TODO: Find a way to have a good provider interface
func main() {
	viper_instance := viper.NewWithOptions()
	app_config, err := config.GetConfig(viper_instance)
	if err != nil {
		panic(err)
	}
	err = utils.SetupLogger(app_config)
	if err != nil {
		panic(err)
	}
	slog.Debug("Current", "config", fmt.Sprintf("%+v", app_config))
	action, err := utils.SelectAction(app_config.ActionName, app_config)
	if err != nil {
		panic(err)
	}
	slog.Debug(fmt.Sprintf(`Found action "%T"`, action))
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	// TODO: fix this function, it has lots of lint errors. Thank you, good day!
	container_name_filter := filters.NewArgs(filters.KeyValuePair{Key: "name", Value: app_config.ContainerName})

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{Filters: container_name_filter})
	if err != nil {
		panic(err)
	} else if len(containers) == 0 {
		panic(fmt.Sprintf(`Could not find a container named "%s"`, app_config.ContainerName))
	} else if len(containers) > 1 {
		panic(fmt.Sprintf(`Container name: "%s" matched %d containers. Please ensure that the container name is unique to one container.`, app_config.ContainerName, len(containers)))
	}
	ctr := containers[0]
	ctr_inspection, err := cli.ContainerInspect(ctx, ctr.ID)
	if err != nil {
		panic(err)
	}
	is_tty := ctr_inspection.Config.Tty
	slog.Debug("Found container:", "ID", ctr.ID, "TTY", is_tty)
	reader, err := cli.ContainerLogs(ctx, ctr.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Since:      "1s",
	})
	if err != nil {
		panic(err)
	}
	for {
		message, err := utils.GetContainerLog(reader, is_tty)

		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Error("Container closed, exiting...")
				os.Exit(0)
			}
			panic(err)
		}
		if len(message) == 0 {
			continue
		}
		slog.Debug(fmt.Sprintf(`Got message of %d bytes`, len(message)))
		err = utils.PerformActionIfMatch(app_config, action, message)
		if err != nil {
			panic(err)
		}
	}
}
