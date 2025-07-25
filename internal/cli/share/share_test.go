package share

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestShareCommandsExist(t *testing.T) {
	// Test that all expected commands are registered
	assert.NotNil(t, ShareCmd)
	assert.NotNil(t, createCmd)
	assert.NotNil(t, listCmd)
	assert.NotNil(t, updateCmd)
	assert.NotNil(t, revokeCmd)
	assert.NotNil(t, sharedSecretsCmd)
}

func TestShareCommandHelp(t *testing.T) {
	// Test that help text is generated correctly
	buf := new(bytes.Buffer)
	ShareCmd.SetOut(buf)
	ShareCmd.SetArgs([]string{"--help"})
	ShareCmd.Execute()
	
	output := buf.String()
	assert.Contains(t, output, "Manage secret sharing")
	assert.Contains(t, output, "create")
	assert.Contains(t, output, "list")
	assert.Contains(t, output, "update")
	assert.Contains(t, output, "revoke")
	assert.Contains(t, output, "shared-secrets")
}

func TestCreateCommandFlags(t *testing.T) {
	// Test that required flags are properly set
	cmd := &cobra.Command{}
	cmd.AddCommand(createCmd)
	
	// Reset flags to default values
	createSecretID = 0
	createRecipientID = 0
	createIsGroup = false
	createPermission = "read"
	
	// Test with missing required flags
	err := createCmd.Execute()
	assert.Error(t, err)
	
	// Test with all required flags
	createCmd.SetArgs([]string{"--secret-id=1", "--recipient-id=2"})
	err = createCmd.Execute()
	// This will still fail because it tries to connect to the database,
	// but at least we know the flags are being parsed correctly
	assert.Error(t, err)
	assert.Equal(t, uint(1), createSecretID)
	assert.Equal(t, uint(2), createRecipientID)
}

func TestUpdateCommandFlags(t *testing.T) {
	// Test that required flags are properly set
	cmd := &cobra.Command{}
	cmd.AddCommand(updateCmd)
	
	// Reset flags to default values
	updateShareID = 0
	updatePermission = ""
	
	// Test with missing required flags
	err := updateCmd.Execute()
	assert.Error(t, err)
	
	// Test with all required flags
	updateCmd.SetArgs([]string{"--share-id=1", "--permission=write"})
	err = updateCmd.Execute()
	// This will still fail because it tries to connect to the database,
	// but at least we know the flags are being parsed correctly
	assert.Error(t, err)
	assert.Equal(t, uint(1), updateShareID)
	assert.Equal(t, "write", updatePermission)
}

func TestRevokeCommandFlags(t *testing.T) {
	// Test that required flags are properly set
	cmd := &cobra.Command{}
	cmd.AddCommand(revokeCmd)
	
	// Reset flags to default values
	revokeShareID = 0
	
	// Test with missing required flags
	err := revokeCmd.Execute()
	assert.Error(t, err)
	
	// Test with all required flags
	revokeCmd.SetArgs([]string{"--share-id=1"})
	err = revokeCmd.Execute()
	// This will still fail because it tries to connect to the database,
	// but at least we know the flags are being parsed correctly
	assert.Error(t, err)
	assert.Equal(t, uint(1), revokeShareID)
}