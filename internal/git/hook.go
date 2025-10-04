package git

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	hookDirName  = ".git/hooks"
	hookFileName = "prepare-commit-msg"
	binName      = "commitgen"
)

/*
hookScript is the shell script that will be written to the Git hook file.

It invokes the commitgen binary to generate the commit message.
The script is designed to be self-contained and robust.
*/
const hookScript = `#!/bin/sh
#
# This Git hook is managed by the 'commitgen' tool.
# To remove it, run: commitgen uninstall-hook
# Debugging: Log hook execution and commitgen output

# The commitgen binary should be in the user's PATH.
# If you are having issues, please add its location to your PATH.
if ! command -v %s &> /dev/null
then
    echo "commitgen: could not find the binary in your PATH."
    echo "Please ensure the commitgen binary is accessible."
    exit 1
fi

# Execute the commitgen binary to generate a message and
# write it to the commit message file.
commit_msg_file="$1"
original_content=$(cat "$commit_msg_file")
%s --commit-msg-file "$commit_msg_file"
`

/*
Install creates or overwrites the prepare-commit-msg Git hook in the current
repository. It returns an error if the hook file cannot be created or written to.
*/
func Install() error {
	repoRoot, err := findGitRoot()
	if err != nil {
		return fmt.Errorf("could not find Git repository root: %w", err)
	}

	hookPath := filepath.Join(repoRoot, hookDirName, hookFileName)
	scriptContent := fmt.Sprintf(hookScript, binName, binName)
	if err := os.WriteFile(hookPath, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("could not write hook file at %s: %w", hookPath, err)
	}

	return nil
}

/*
Uninstall removes the prepare-commit-msg Git hook from the current repository.
It returns a nil error if the file does not exist.
*/
func Uninstall() error {
	repoRoot, err := findGitRoot()
	if err != nil {
		return fmt.Errorf("could not uninstall hook: %w", err)
	}

	hookPath := filepath.Join(repoRoot, hookDirName, hookFileName)
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		// Check if the file exists before attempting to remove it
		return nil
	}

	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("could not remove hook file at %s: %w", hookPath, err)
	}
	return nil
}

// findGitRoot traverses up the directory tree to find the root of the Git repository.
func findGitRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get current working directory: %w", err)
	}

	currentDir := wd
	for {
		gitPath := filepath.Join(currentDir, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return currentDir, nil
		}

		// Move up
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached the filesystem root
			return "", fmt.Errorf("not a git repository")
		}
		currentDir = parentDir
	}
}
