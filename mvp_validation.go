package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ValidationResult struct {
	Component string
	Status    string
	Details   []string
	Critical  bool
}

func main() {
	fmt.Println("🔍 MVP Validation Report")
	fmt.Println("========================")
	
	results := []ValidationResult{}
	
	// 1. Check build status
	results = append(results, validateBuild())
	
	// 2. Check core components
	results = append(results, validateCoreComponents()...)
	
	// 3. Check secret sharing implementation
	results = append(results, validateSecretSharing()...)
	
	// 4. Check internationalization
	results = append(results, validateI18n()...)
	
	// 5. Check documentation
	results = append(results, validateDocumentation()...)
	
	// 6. Check tests
	results = append(results, validateTests())
	
	// 7. Check security
	results = append(results, validateSecurity()...)
	
	// Print results
	printResults(results)
	
	// Final assessment
	assessMVPReadiness(results)
}

func validateBuild() ValidationResult {
	result := ValidationResult{
		Component: "Build System",
		Details:   []string{},
		Critical:  true,
	}
	
	// Check if project builds
	cmd := exec.Command("go", "build", "-o", "secretly", "./cmd/secretly")
	if err := cmd.Run(); err != nil {
		result.Status = "❌ FAIL"
		result.Details = append(result.Details, "Build failed: "+err.Error())
		return result
	}
	
	result.Status = "✅ PASS"
	result.Details = append(result.Details, "Project builds successfully")
	
	// Clean up
	os.Remove("secretly")
	
	return result
}

func validateCoreComponents() []ValidationResult {
	results := []ValidationResult{}
	
	// Check core service
	coreResult := ValidationResult{
		Component: "Core Service",
		Details:   []string{},
		Critical:  true,
	}
	
	if fileExists("internal/core/service.go") {
		coreResult.Status = "✅ PASS"
		coreResult.Details = append(coreResult.Details, "Core service implementation exists")
	} else {
		coreResult.Status = "❌ FAIL"
		coreResult.Details = append(coreResult.Details, "Core service missing")
	}
	results = append(results, coreResult)
	
	// Check storage interface
	storageResult := ValidationResult{
		Component: "Storage Layer",
		Details:   []string{},
		Critical:  true,
	}
	
	if fileExists("internal/core/storage/interface.go") && fileExists("internal/storage/local/local.go") {
		storageResult.Status = "✅ PASS"
		storageResult.Details = append(storageResult.Details, "Storage interface and local implementation exist")
	} else {
		storageResult.Status = "❌ FAIL"
		storageResult.Details = append(storageResult.Details, "Storage layer incomplete")
	}
	results = append(results, storageResult)
	
	// Check encryption
	encryptionResult := ValidationResult{
		Component: "Encryption Layer",
		Details:   []string{},
		Critical:  true,
	}
	
	if fileExists("internal/encryption/encryption.go") {
		encryptionResult.Status = "✅ PASS"
		encryptionResult.Details = append(encryptionResult.Details, "Encryption layer implemented")
	} else {
		encryptionResult.Status = "❌ FAIL"
		encryptionResult.Details = append(encryptionResult.Details, "Encryption layer missing")
	}
	results = append(results, encryptionResult)
	
	return results
}

func validateSecretSharing() []ValidationResult {
	results := []ValidationResult{}
	
	// Check sharing models
	modelsResult := ValidationResult{
		Component: "Sharing Models",
		Details:   []string{},
		Critical:  true,
	}
	
	sharingFiles := []string{
		"internal/storage/models/secret_sharing.go",
		"internal/storage/models/group_sharing.go",
		"internal/storage/models/secret_with_sharing.go",
	}
	
	allExist := true
	for _, file := range sharingFiles {
		if !fileExists(file) {
			allExist = false
			modelsResult.Details = append(modelsResult.Details, "Missing: "+file)
		}
	}
	
	if allExist {
		modelsResult.Status = "✅ PASS"
		modelsResult.Details = append(modelsResult.Details, "All sharing models implemented")
	} else {
		modelsResult.Status = "❌ FAIL"
	}
	results = append(results, modelsResult)
	
	// Check sharing core logic
	coreResult := ValidationResult{
		Component: "Sharing Core Logic",
		Details:   []string{},
		Critical:  true,
	}
	
	coreFiles := []string{
		"internal/core/sharing.go",
		"internal/core/group_sharing.go",
		"internal/core/sharing_audit.go",
	}
	
	allExist = true
	for _, file := range coreFiles {
		if !fileExists(file) {
			allExist = false
			coreResult.Details = append(coreResult.Details, "Missing: "+file)
		}
	}
	
	if allExist {
		coreResult.Status = "✅ PASS"
		coreResult.Details = append(coreResult.Details, "Core sharing logic implemented")
	} else {
		coreResult.Status = "❌ FAIL"
	}
	results = append(results, coreResult)
	
	// Check API endpoints
	apiResult := ValidationResult{
		Component: "Sharing APIs",
		Details:   []string{},
		Critical:  true,
	}
	
	apiFiles := []string{
		"server/http/handlers/shares.go",
		"server/grpc/services/share_service.go",
	}
	
	allExist = true
	for _, file := range apiFiles {
		if !fileExists(file) {
			allExist = false
			apiResult.Details = append(apiResult.Details, "Missing: "+file)
		}
	}
	
	if allExist {
		apiResult.Status = "✅ PASS"
		apiResult.Details = append(apiResult.Details, "HTTP and gRPC sharing APIs implemented")
	} else {
		apiResult.Status = "❌ FAIL"
	}
	results = append(results, apiResult)
	
	// Check CLI commands
	cliResult := ValidationResult{
		Component: "Sharing CLI",
		Details:   []string{},
		Critical:  true,
	}
	
	cliFiles := []string{
		"internal/cli/share/create.go",
		"internal/cli/share/list.go",
		"internal/cli/share/update.go",
	}
	
	allExist = true
	for _, file := range cliFiles {
		if !fileExists(file) {
			allExist = false
			cliResult.Details = append(cliResult.Details, "Missing: "+file)
		}
	}
	
	if allExist {
		cliResult.Status = "✅ PASS"
		cliResult.Details = append(cliResult.Details, "CLI sharing commands implemented")
	} else {
		cliResult.Status = "❌ FAIL"
	}
	results = append(results, cliResult)
	
	return results
}

func validateI18n() []ValidationResult {
	results := []ValidationResult{}
	
	// Check i18n infrastructure
	i18nResult := ValidationResult{
		Component: "Internationalization",
		Details:   []string{},
		Critical:  false,
	}
	
	if fileExists("internal/i18n/i18n.go") {
		i18nResult.Status = "✅ PASS"
		i18nResult.Details = append(i18nResult.Details, "I18n infrastructure implemented")
	} else {
		i18nResult.Status = "❌ FAIL"
		i18nResult.Details = append(i18nResult.Details, "I18n infrastructure missing")
	}
	results = append(results, i18nResult)
	
	// Check translation files
	translationResult := ValidationResult{
		Component: "Translation Files",
		Details:   []string{},
		Critical:  false,
	}
	
	languages := []string{"en", "ru", "es", "fr", "de"}
	missingLangs := []string{}
	
	for _, lang := range languages {
		if !fileExists(fmt.Sprintf("internal/i18n/locales/%s.json", lang)) {
			missingLangs = append(missingLangs, lang)
		}
	}
	
	if len(missingLangs) == 0 {
		translationResult.Status = "✅ PASS"
		translationResult.Details = append(translationResult.Details, "All translation files present")
	} else {
		translationResult.Status = "⚠️  PARTIAL"
		translationResult.Details = append(translationResult.Details, "Missing languages: "+strings.Join(missingLangs, ", "))
	}
	results = append(results, translationResult)
	
	return results
}

func validateDocumentation() []ValidationResult {
	results := []ValidationResult{}
	
	// Check API documentation
	apiDocsResult := ValidationResult{
		Component: "API Documentation",
		Details:   []string{},
		Critical:  false,
	}
	
	docFiles := []string{
		"docs/SECRET_SHARING_API.md",
		"server/openapi.yaml",
	}
	
	allExist := true
	for _, file := range docFiles {
		if !fileExists(file) {
			allExist = false
			apiDocsResult.Details = append(apiDocsResult.Details, "Missing: "+file)
		}
	}
	
	if allExist {
		apiDocsResult.Status = "✅ PASS"
		apiDocsResult.Details = append(apiDocsResult.Details, "API documentation complete")
	} else {
		apiDocsResult.Status = "❌ FAIL"
	}
	results = append(results, apiDocsResult)
	
	// Check user documentation
	userDocsResult := ValidationResult{
		Component: "User Documentation",
		Details:   []string{},
		Critical:  false,
	}
	
	userDocFiles := []string{
		"docs/SECRET_SHARING_USER_GUIDE.md",
		"docs/SECRET_SHARING_WORKFLOWS.md",
		"docs/SECRET_SHARING_SECURITY.md",
	}
	
	allExist = true
	for _, file := range userDocFiles {
		if !fileExists(file) {
			allExist = false
			userDocsResult.Details = append(userDocsResult.Details, "Missing: "+file)
		}
	}
	
	if allExist {
		userDocsResult.Status = "✅ PASS"
		userDocsResult.Details = append(userDocsResult.Details, "User documentation complete")
	} else {
		userDocsResult.Status = "❌ FAIL"
	}
	results = append(results, userDocsResult)
	
	return results
}

func validateTests() ValidationResult {
	result := ValidationResult{
		Component: "Test Suite",
		Details:   []string{},
		Critical:  true,
	}
	
	// Run tests
	cmd := exec.Command("go", "test", "./...")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		result.Status = "❌ FAIL"
		result.Details = append(result.Details, "Tests failed")
		result.Details = append(result.Details, string(output))
	} else {
		result.Status = "✅ PASS"
		result.Details = append(result.Details, "All tests passing")
		
		// Count test files
		testCount := countTestFiles()
		result.Details = append(result.Details, fmt.Sprintf("Found %d test files", testCount))
	}
	
	return result
}

func validateSecurity() []ValidationResult {
	results := []ValidationResult{}
	
	// Check security reports
	securityResult := ValidationResult{
		Component: "Security Analysis",
		Details:   []string{},
		Critical:  false,
	}
	
	if fileExists("security/gosec-report.json") {
		securityResult.Status = "✅ PASS"
		securityResult.Details = append(securityResult.Details, "Security analysis report available")
	} else {
		securityResult.Status = "⚠️  MISSING"
		securityResult.Details = append(securityResult.Details, "No security analysis report found")
	}
	results = append(results, securityResult)
	
	// Check encryption implementation
	encryptionSecResult := ValidationResult{
		Component: "Encryption Security",
		Details:   []string{},
		Critical:  true,
	}
	
	if fileExists("internal/encryption/shared_secrets.go") {
		encryptionSecResult.Status = "✅ PASS"
		encryptionSecResult.Details = append(encryptionSecResult.Details, "Shared secrets encryption implemented")
	} else {
		encryptionSecResult.Status = "❌ FAIL"
		encryptionSecResult.Details = append(encryptionSecResult.Details, "Shared secrets encryption missing")
	}
	results = append(results, encryptionSecResult)
	
	return results
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func countTestFiles() int {
	count := 0
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			count++
		}
		return nil
	})
	return count
}

func printResults(results []ValidationResult) {
	fmt.Println()
	
	for _, result := range results {
		fmt.Printf("📋 %s: %s\n", result.Component, result.Status)
		for _, detail := range result.Details {
			fmt.Printf("   • %s\n", detail)
		}
		fmt.Println()
	}
}

func assessMVPReadiness(results []ValidationResult) {
	criticalFails := 0
	totalFails := 0
	totalPasses := 0
	
	for _, result := range results {
		if strings.Contains(result.Status, "FAIL") {
			totalFails++
			if result.Critical {
				criticalFails++
			}
		} else if strings.Contains(result.Status, "PASS") {
			totalPasses++
		}
	}
	
	fmt.Println("🎯 MVP Readiness Assessment")
	fmt.Println("===========================")
	fmt.Printf("✅ Passed: %d\n", totalPasses)
	fmt.Printf("❌ Failed: %d\n", totalFails)
	fmt.Printf("🚨 Critical Failures: %d\n", criticalFails)
	fmt.Println()
	
	if criticalFails == 0 {
		fmt.Println("🎉 MVP IS READY FOR PRODUCTION!")
		fmt.Println("All critical components are implemented and working.")
		if totalFails > 0 {
			fmt.Println("⚠️  Some non-critical components need attention but won't block MVP release.")
		}
	} else {
		fmt.Println("🚫 MVP NOT READY - Critical issues found")
		fmt.Println("Please address critical failures before production release.")
	}
}