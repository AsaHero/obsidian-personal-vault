package compute

type CommandID uint8

const (
	UNKNOWN CommandID = iota
	GET
	SET
	DEL
)

var (
	unknownCommand = "UNKNOWN"
	getCommand     = "GET"
	setCommand     = "SET"
	delCommand     = "DEL"
)

var nameToID = map[string]CommandID{
	unknownCommand: UNKNOWN,
	getCommand:     GET,
	setCommand:     SET,
	delCommand:     DEL,
}

func CommandNameToCommandID(name string) CommandID {
	id, found := nameToID[name]
	if !found {
		return UNKNOWN
	}

	return id
}

const (
	setCommandArgumentNumber = 2
	getCommandArgumentNumber = 1
	delCommandArgumentNumber = 1
)

var commandIDToArgumentNum = map[CommandID]int{
	GET: getCommandArgumentNumber,
	SET: setCommandArgumentNumber,
	DEL: delCommandArgumentNumber,
}

func CommandIDToArgumentNUmber(cmd CommandID) int {
	return commandIDToArgumentNum[cmd]
}
