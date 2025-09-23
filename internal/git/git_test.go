package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

/*
setupTestRepo initializes a new Git repository in a temporary directory
and returns the path to that directory. It's a helper for integration tests.
*/
func setupTestRepo(t *testing.T) string {
	t.Helper()
	repoPath := t.TempDir()

	runCmd := func(args ...string) {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repoPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command %q failed: %v\nOutput: %s", strings.Join(args, " "), err, string(output))
		}
	}

	runCmd("git", "init")
	runCmd("git", "config", "user.name", "Test User")
	runCmd("git", "config", "user.email", "test@example.com")
	return repoPath
}

// TestGetStagedDiff covers the primary scenarios for git diff command.
func TestGetStagedDiff(t *testing.T) {
	t.Run("no staged changes", func(t *testing.T) {
		repoPath := setupTestRepo(t)
		os.Chdir(repoPath)

		diff, err := GetStagedDiff()
		if err != nil {
			t.Fatalf("GetStagedDiff() returned an unexpected error: %v", err)
		}
		if diff != "" {
			t.Errorf("expected an empty diff, but got %q", diff)
		}
	})

	t.Run("with staged changes", func(t *testing.T) {
		repoPath := setupTestRepo(t)
		os.Chdir(repoPath)

		// Create and stage a file
		filePath := filepath.Join(repoPath, "test.txt")
		os.WriteFile(filePath, []byte("hello world"), 0644)
		exec.Command("git", "add", "test.txt").Run()

		diff, err := GetStagedDiff()
		if err != nil {
			t.Fatalf("GetStagedDiff() returned an unexpected error: %v", err)
		}
		if !strings.Contains(diff, "+hello world") {
			t.Errorf("expected diff to contain '+hello world', but got %q", diff)
		}
	})
}

/*
TestCommit covers the primary scenarios for committing staged changes.

Each subtest goes through the following steps:

1. Create and stage a file.

2. Execute the commit function.

3. Verify correct author name/email.
*/
func TestCommit(t *testing.T) {
	t.Run("commit with author override", func(t *testing.T) {
		repoPath := setupTestRepo(t)
		os.Chdir(repoPath)

		// Stage a file to be committed
		filePath := filepath.Join(repoPath, "file.txt")
		os.WriteFile(filePath, []byte("content"), 0644)
		exec.Command("git", "add", "file.txt").Run()

		// Perform the commit with custom author info
		authorName, authorEmail := "Custom Author", "custom@example.com"
		err := Commit("feat: test commit", authorName, authorEmail)
		if err != nil {
			t.Fatalf("Commit() failed: %v", err)
		}

		// Verify the author of the last commit
		cmd := exec.Command("git", "log", "-1", "--pretty=format:%an <%ae>")
		output, _ := cmd.CombinedOutput()
		expectedAuthor := "Custom Author <custom@example.com>"
		if strings.TrimSpace(string(output)) != expectedAuthor {
			t.Errorf("expected commit author to be %q, but got %q", expectedAuthor, string(output))
		}
	})

	t.Run("commit without author override", func(t *testing.T) {
		repoPath := setupTestRepo(t)
		os.Chdir(repoPath)

		// Stage a file to be committed
		filePath := filepath.Join(repoPath, "file.txt")
		os.WriteFile(filePath, []byte("content"), 0644)
		exec.Command("git", "add", "file.txt").Run()

		// Perform the commit without author info (should use git config)
		err := Commit("fix: default author commit", "", "")
		if err != nil {
			t.Fatalf("Commit() failed: %v", err)
		}

		// Verify the author of the last commit
		cmd := exec.Command("git", "log", "-1", "--pretty=format:%an <%ae>")
		output, _ := cmd.CombinedOutput()
		expectedAuthor := "Test User <test@example.com>"
		if strings.TrimSpace(string(output)) != expectedAuthor {
			t.Errorf("expected commit author to be %q, but got %q", expectedAuthor, string(output))
		}
	})
}

// TestInstallAndUninstallHook tests Install and Uninstall functions for pre-commit hook.
func TestInstallAndUninstallHook(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Change into the repo directory for the duration of the test
	originalWD, _ := os.Getwd()
	os.Chdir(repoPath)
	defer os.Chdir(originalWD)

	hookPath := filepath.Join(repoPath, ".git", "hooks", "prepare-commit-msg")

	// --- Test Install ---
	if err := Install(); err != nil {
		t.Fatalf("Install() failed: %v", err)
	}

	// Verify the hook file was created
	info, err := os.Stat(hookPath)
	if os.IsNotExist(err) {
		t.Fatal("expected hook file to be created, but it was not")
	}

	// Verify the hook file is executable
	if info.Mode().Perm()&0111 == 0 {
		t.Errorf("expected hook file to be executable, but it was not (mode: %s)", info.Mode().Perm())
	}

	// --- Test Uninstall ---
	if err := Uninstall(); err != nil {
		t.Fatalf("Uninstall() failed: %v", err)
	}

	// Verify the hook file was removed
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Fatal("expected hook file to be removed, but it still exists")
	}

	// --- Test Uninstall when already removed ---
	if err := Uninstall(); err != nil {
		t.Fatalf("Uninstall() on a non-existent hook should not fail, but got: %v", err)
	}
}

// TestFindGitRoot covers the primary scenarios for git root directory finding.
func TestFindGitRoot(t *testing.T) {
	originalWD, _ := os.Getwd()
	defer os.Chdir(originalWD)

	t.Run("from a subdirectory", func(t *testing.T) {
		repoPath := setupTestRepo(t)
		subDir := filepath.Join(repoPath, "deep", "nested", "folder")
		os.MkdirAll(subDir, 0755)
		os.Chdir(subDir)

		foundRoot, err := findGitRoot()
		if err != nil {
			t.Fatalf("findGitRoot() failed: %v", err)
		}

		// Clean paths for reliable comparison
		expected, _ := filepath.EvalSymlinks(repoPath)
		found, _ := filepath.EvalSymlinks(foundRoot)
		if found != expected {
			t.Errorf("expected to find root at %q, but got %q", expected, found)
		}
	})

	t.Run("not in a git repository", func(t *testing.T) {
		// A temporary directory is not a git repo by default
		nonRepoPath := t.TempDir()
		os.Chdir(nonRepoPath)

		_, err := findGitRoot()
		if err == nil {
			t.Fatal("expected an error when running outside a git repository, but got nil")
		}
	})
}
