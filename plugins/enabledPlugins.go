package plugins

import (
	"plugin"
)

// LoadedPlugins stores references to the stored plugins
var LoadedPlugins map[string]*plugin.Plugin
