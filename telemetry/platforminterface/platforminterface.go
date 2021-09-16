package platforminterface

type PlatformInterface interface {
	CheckIfFileExists(filepath string) (bool, error)
	ExecuteCommand(command string) (string, error)
	ClearNetworkConfiguration() (bool, error)
	KillProcessByName(processName string) error
}
