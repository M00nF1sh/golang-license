package module

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
	"net/http"
	"strings"
)

type RepositoryLocator interface {
	Locate(mod *packages.Module) (string, error)
}

func NewRepositoryLocator() RepositoryLocator {
	return &repositoryLocator{}
}

type repositoryLocator struct {
}

func (l *repositoryLocator) Locate(mod *packages.Module) (string, error) {
	moduleURL := fmt.Sprintf("https://%v?go-get=1", mod.Path)
	resp, err := http.Get(moduleURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.Errorf("failed to get repository meta, status_code: %v", resp.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}
	selection := doc.Find("head > meta[name=go-import]")
	if selection.Length() != 1 {
		return "", errors.Errorf("failed to find go-import meta")
	}
	goImportMeta := selection.Get(0)
	var goImportContent string
	for _, attr := range goImportMeta.Attr {
		if attr.Key == "content" {
			goImportContent = attr.Val
		}
	}
	return l.parseRepoURL(goImportContent), nil
}

func (l *repositoryLocator) parseRepoURL(goImportContent string) string {
	fields := strings.Fields(goImportContent)
	var repoURL string
	if len(fields) == 3 {
		repoURL = fields[2]
	}
	repoURL = strings.TrimSuffix(repoURL, ".git")
	return repoURL
}
