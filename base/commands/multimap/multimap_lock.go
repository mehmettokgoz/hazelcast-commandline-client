//go:build std || multimap

package multimap

import (
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func init() {
	c := commands.NewLockCommand("MultiMap", getMultiMap)
	check.Must(plug.Registry.RegisterCommand("multi-map:lock", c, plug.OnlyInteractive{}))
}
