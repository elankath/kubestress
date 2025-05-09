package cli

const (
	ExitSuccess = iota
	ExitBasicInvocation
	ExitCreateServices
	ExitExecuteLoader
	ExitMissingNecessaryArgument
	ExitInvalidArg
)
