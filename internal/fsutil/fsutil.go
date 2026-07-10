// Package fsutil holds small filesystem helpers shared across gyroscope's writers.
package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteGuarded writes content to repoDir/relPath, creating parent dirs. It refuses
// to clobber an existing file unless force is set (O_EXCL is the guarantee); the
// refusal error names relPath and points at --force. With force it truncates.
func WriteGuarded(repoDir, relPath string, content []byte, force bool) error {
	dest := filepath.Join(repoDir, relPath)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if force {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	fh, err := os.OpenFile(dest, flags, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("refusing to overwrite %s (use --force)", relPath)
		}
		return err
	}
	_, werr := fh.Write(content)
	if cerr := fh.Close(); werr == nil {
		werr = cerr
	}
	return werr
}

// WriteAtomic writes content to repoDir/relPath via a temp file + rename, creating
// parent dirs. Unlike WriteGuarded it OVERWRITES an existing file — the atomic
// rename means a reader never sees a partial file and an interrupted write cannot
// truncate the target. Reserved for gyroscope-performed rewrites of files that are
// expected to already exist (the TODO→DONE archive move), never first-time
// scaffolds — those keep WriteGuarded's clobber refusal.
func WriteAtomic(repoDir, relPath string, content []byte) error {
	dest := filepath.Join(repoDir, relPath)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	tmp := dest + ".gyroscope.tmp"
	if err := os.WriteFile(tmp, content, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("write %s: %w", relPath, err)
	}
	return nil
}
