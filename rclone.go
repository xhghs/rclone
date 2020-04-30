// Sync files and directories to and from local and remote object stores
//
// Nick Craig-Wood <nick@craig-wood.com>
package main

import (
	_ "backend/all" // import all backends
	"cmd"
	_ "cmd/all"    // import all commands
	_ "lib/plugin" // import plugins
)

func main() {
	cmd.Main()
}
