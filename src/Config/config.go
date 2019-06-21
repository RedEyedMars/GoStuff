package Config

import "log"

var Log ConfigType

type ConfigType struct {
	state  int
	states map[string]int
}

func (state *ConfigType) HasState(div string) bool {
	return (state.state % (state.states[div])) == 0
}

func Setup() {
	setup(
		SetupLogConfig,
		StdConfig)
	log.Println(Log)
}

func setup(funcs ...func()) {
	for _, f := range funcs {
		f()
	}
}

func SetupLogConfig() {
	Log.states = map[string]int{
		"verbose":      3,
		"very_verbose": 5,
		"debug":        7,
		"error":        11,
		"assert":       13,
		"event":        17,
		"warning":      19}
}

func StdConfig() {
	Log.state = 1
	for _, value := range Log.states {
		Log.state = Log.state * value
	}
	//Log.state = Log.states["verbose"] * Log.states["very_verbose"] * Log.states["debug"] * Log.states["error"] * Log.states["assert"] * Log.states["event"] * Log.states["warning"]
}
