package merror

func SpawnCommonError(module ModuleNumber, specificErrNum uint32) uint32 {
	return SpawnError(Common, module, specificErrNum)
}

func SpawnClientError(module ModuleNumber, specificErrNum uint32) uint32 {
	return SpawnError(Client, module, specificErrNum)
}

func SpawnSystemError(module ModuleNumber, specificErrNum uint32) uint32 {
	return SpawnError(System, module, specificErrNum)
}

func SpawnThirdPartyError(module ModuleNumber, specificErrNum uint32) uint32 {
	return SpawnError(ThirdParty, module, specificErrNum)
}

// [ common/client/system/third-party/unknown ]x1 [ module ]x2 [ specific error ]x2
// e.g., 10001 means common error, user module, user exist.
func SpawnError(areaNum AreaNumber, module ModuleNumber, specificErrNum uint32) uint32 {
	return uint32(areaNum) + module*100 + specificErrNum
}
