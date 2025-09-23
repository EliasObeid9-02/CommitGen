package git

import (
	"fmt"
	"os/exec"
)

// GetStagedDiff returns the diff of the currently staged files.
func GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) == 0 {
			return "", nil
		}
		return "", fmt.Errorf("could not get staged diff: %w, output: %s", err, string(output))
	}
	return string(output), nil
}

// Commit creates a new commit with the given message, author, and committer info.
func Commit(message, authorName, authorEmail string) error {
	var cmd *exec.Cmd
	if authorName != "" && authorEmail != "" {
		cmd = exec.Command("git", "commit", "-m", message)
		cmd.Env = append(cmd.Environ(),
			fmt.Sprintf("GIT_AUTHOR_NAME=%s", authorName),
			fmt.Sprintf("GIT_AUTHOR_EMAIL=%s", authorEmail),
			fmt.Sprintf("GIT_COMMITTER_NAME=%s", authorName),
			fmt.Sprintf("GIT_COMMITTER_EMAIL=%s", authorEmail),
		)
	} else {
		cmd = exec.Command("git", "commit", "-m", message)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to commit: %w, output: %s", err, string(output))
	}

	// TODO: remove
	fmt.Println(string(output))
	return nil
}
