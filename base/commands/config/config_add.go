//go:build std || config

package config

import (
	"context"
	"fmt"
	"math"
	"path/filepath"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/str"
	"github.com/hazelcast/hazelcast-commandline-client/internal/types"
)

type AddCommand struct{}

func (AddCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("add")
	short := "Adds a configuration"
	long := `Adds a configuration with the given KEY=VALUE pairs and saves it with configuration name.
	
Overrides the previous configuration if it exists.
	
The following keys are supported:
	
	* cluster.name             STRING
	* cluster.address          STRING
	* cluster.user             STRING
	* cluster.password         STRING
	* cluster.discovery-token  STRING
	* cluster.api-base         STRING
	* cluster.viridian-id      STRING
	* ssl.enabled              BOOLEAN (true / false)
	* ssl.server               STRING
	* ssl.skip-verify          BOOLEAN (true / false)
	* ssl.ca-path              STRING
	* ssl.key-path             STRING
	* ssl.key-password         STRING
	* log.path                 STRING
	* log.level                ENUM (error / warn / info / debug)
	
`
	cc.SetCommandHelp(long, short)
	cc.AddStringArg(argConfigName, argTitleConfigName)
	cc.AddStringSliceArg(argKeyValues, argTitleKeyValues, 0, math.MaxInt)
	return nil
}

func (AddCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	target := ec.GetStringArg(argConfigName)
	var opts types.KeyValues[string, string]
	for _, arg := range ec.GetStringSliceArg(argKeyValues) {
		k, v := str.ParseKeyValue(arg)
		if k == "" {
			return fmt.Errorf("invalid key=value pair: %s", arg)
		}
		opts = append(opts, types.KeyValue[string, string]{
			Key:   k,
			Value: v,
		})
	}
	path, stop, err := cmd.ExecuteBlocking(ctx, ec, func(ctx context.Context, sp clc.Spinner) (string, error) {
		sp.SetText("Creating the configuration")
		dir, cfgPath, err := config.Create(target, opts)
		if err != nil {
			return "", err
		}
		mopt := config.ConvertKeyValuesToMap(opts)
		// ignoring the JSON path for now
		_, _, err = config.CreateJSON(target, mopt)
		if err != nil {
			ec.Logger().Warn("Failed creating the JSON configuration: %s", err.Error())
		}
		return filepath.Join(dir, cfgPath), nil
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Created the configuration at: %s.", path)
	ec.PrintlnUnnecessary(msg)
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("config:add", &AddCommand{}))
}
