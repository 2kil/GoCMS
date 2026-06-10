package handlers

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"shuaitesteel.com/cms/config"
)

var slugPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

func IsValidSlug(slug string) bool {
	return slugPattern.MatchString(slug)
}

func safePublicFilePath(requestPath string) (string, bool) {
	if strings.Contains(requestPath, "\\") || strings.ContainsRune(requestPath, 0) {
		return "", false
	}
	for _, segment := range strings.Split(requestPath, "/") {
		if segment == ".." {
			return "", false
		}
	}

	cleanPath := path.Clean("/" + requestPath)
	if cleanPath == "/" {
		cleanPath = "/index.html"
	}

	rel := strings.TrimPrefix(cleanPath, "/")
	if rel == "" || rel == "." || rel == ".." || strings.HasPrefix(rel, "../") {
		return "", false
	}

	baseAbs, err := filepath.Abs(config.Cfg.Static.Dir)
	if err != nil {
		return "", false
	}
	fullAbs, err := filepath.Abs(filepath.Join(baseAbs, filepath.FromSlash(rel)))
	if err != nil {
		return "", false
	}
	relToBase, err := filepath.Rel(baseAbs, fullAbs)
	if err != nil || relToBase == ".." || strings.HasPrefix(relToBase, ".."+string(os.PathSeparator)) {
		return "", false
	}

	return fullAbs, true
}
