package tooldocs

// commandSkillNames maps a command to the agent skill that documents it. A static table rather than
// a name derived from the command: `enable-pause-point` is documented by `tloop-pause-point`, not by
// a `tloop-enable-pause-point` skill that does not exist, and four commands share that one skill.
// A command absent from this table gets no guidance line at all, so custom commands are never
// pointed at a skill nobody installed.
var commandSkillNames = map[string]string{
	"clear-console":        "tloop-clear-console",
	"compile":              "tloop-compile",
	"control-play-mode":    "tloop-control-play-mode",
	"execute-dynamic-code": "tloop-execute-dynamic-code",
	"find-game-objects":    "tloop-find-game-objects",
	"focus-window":         "tloop-focus-window",
	"get-hierarchy":        "tloop-get-hierarchy",
	"get-logs":             "tloop-get-logs",
	"record-video":         "tloop-record-video",
	"run-tests":            "tloop-run-tests",
	"screenshot":           "tloop-screenshot",
	"set-game-view-size":   "tloop-set-game-view-size",
	"game-view-mode":       "tloop-game-view-mode",
	"simulate-keyboard":    "tloop-simulate-keyboard",
	"simulate-mouse-input": "tloop-simulate-mouse-input",
	"simulate-mouse-ui":    "tloop-simulate-mouse-ui",
	"hot-reload":           "tloop-hot-reload",

	"launch": "tloop-launch",
	"status": "tloop-status",
	// The status skill documents cancel in references/cancel.md: cancel is the way out of a Busy
	// that status reports, and a skill of its own would sit in every session for a rare recovery.
	"cancel": "tloop-status",

	// One skill covers the four pause-point commands and the three watch commands: watch
	// expressions are documented by the pause-point skill's references/watch-expressions.md.
	"enable-pause-point": "tloop-pause-point",
	"clear-pause-point":  "tloop-pause-point",
	"await-pause-point":  "tloop-pause-point",
	"pause-point-status": "tloop-pause-point",
	"enable-watch":       "tloop-pause-point",
	"clear-watch":        "tloop-pause-point",
	"get-watch-values":   "tloop-pause-point",
}

// SkillGuidanceLine returns the closing line of a command's --help output: an instruction to load
// the skill that documents it. `--help` can only list option names and one-line summaries, so the
// workflow rules, response shapes, and failure diagnoses live in the skill; without this line an
// agent that found the command through --help has no way to learn the skill exists.
//
// Phrased as an instruction rather than a cross-reference, because a line that merely mentions a
// document is easy to read past. It names what the skill adds rather than referring to the options
// above, because commands such as focus-window have no options for that phrasing to point at.
func SkillGuidanceLine(command string) (string, bool) {
	skillName, ok := commandSkillNames[command]
	if !ok {
		return "", false
	}
	return "Load the " + skillName +
		" skill for workflow rules and response fields that --help does not cover.", true
}
