# System Initialization Example

This example demonstrates the complete system initialization and validation process for Secretly.

## What This Example Shows

1. **System Validation** - How to validate an existing Secretly setup
2. **File Structure** - What files should exist after initialization
3. **Available Commands** - All system and encryption management commands
4. **Configuration Structure** - Overview of configuration sections
5. **Security Recommendations** - Best practices for secure deployment

## Running the Example

```bash
# From the project root directory
go run examples/system_init/main.go
```

## Prerequisites

Before running this example, you should have initialized your Secretly system:

```bash
# Initialize the system first
secretly system init

# Then run the example to see the validation results
go run examples/system_init/main.go
```

## Expected Output

The example will show:
- ✅ Validation results for your current setup
- 📁 File structure status (which files exist)
- 🔧 Available commands for system management
- ⚙️ Configuration sections overview
- 🛡️ Security best practices

## What You'll Learn

- How to validate your Secretly system setup
- What files are created during initialization
- How to use system management commands
- Security best practices for production deployment
- Configuration structure and options

## Next Steps

After running this example:
1. Try the various commands shown in the output
2. Experiment with different initialization options
3. Practice system validation and auditing
4. Review the security recommendations