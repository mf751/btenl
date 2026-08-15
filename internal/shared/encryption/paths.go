package encryption

import (
	"fmt"
	"os"
	"path/filepath"
)

func dir(base, sub string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("encryption: home dir: %w", err)
	}
	dir := filepath.Join(home, base, sub)
	if err = os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("encryption mkdir %v: %w", dir, err)
	}
	return dir, nil
}

func CertsDir() (string, error) {
	return dir(".btenl", "")
}

func CertsPaths() (string, error) {
	d, err := CertsDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(d, "btenl.crt"), nil
}

func IdentityDir() (string, error) {
	return dir(".btenl", "")
}

func IdentityPaths() (string, string, error) {
	d, err := IdentityDir()
	if err != nil {
		return "", "", err
	}

	return filepath.Join(d, "id_ed25519"), filepath.Join(d, "id_ed25519.pub"), nil
}

// INFO: writes to a temp file and after finishing renames to path to prevent partial writes
func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".btenl-*")
	if err != nil {
		return fmt.Errorf("encryption: temp file: %w", err)
	}
	tmpName := tmp.Name()

	_, err = tmp.Write(data)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("encryption: write %s: %w", path, err)
	}

	if err = os.Chmod(tmpName, mode); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("encryption: chmos %v: %w", path, err)
	}

	if err = os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("encryption: rename %s: %w", path, err)
	}

	return nil
}
