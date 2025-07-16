package system

//func runAudit() {
//	cfg, err := config.Load("")
//	if err != nil {
//		fmt.Println("Failed to load config:", err)
//		os.Exit(1)
//	}
//
//	files := []securefiles.FilePermSpec{
//		{Path: "config.yaml", Mode: 0600},
//		{Path: cfg.Storage.Encryption.KEKPath, Mode: 0600},
//		{Path: cfg.Storage.Encryption.DEKPath, Mode: 0600},
//	}
//
//	// Just check - without fixes
//	err = securefiles.FixFilePerms(files, false)
//	if err != nil {
//		fmt.Println("\nAudit finished with warnings/errors. Please fix the issues.")
//		os.Exit(1)
//	}
//
//	fmt.Println("Audit passed: all critical files have correct permissions and ownership.")
//}
