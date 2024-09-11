package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/sanitize"
)

type FolderHandler struct {
	location string
}

func (h *FolderHandler) CRDs() ([]*v1.CustomResourceDefinition, error) {
	if _, err := os.Stat(h.location); os.IsNotExist(err) {
		return nil, fmt.Errorf("file under '%s' does not exist", h.location)
	}

	var crds []*v1.CustomResourceDefinition

	if err := filepath.Walk(h.location, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".yaml" {
			fmt.Fprintln(os.Stderr, "skipping file "+path)

			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		content, err = sanitize.Sanitize(content)
		if err != nil {
			return fmt.Errorf("failed to sanitize content: %w", err)
		}

		crd := &v1.CustomResourceDefinition{}
		if err := yaml.Unmarshal(content, crd); err != nil {
			fmt.Fprintln(os.Stderr, "skipping none CRD file: "+path)

			return nil //nolint:nilerr // intentional
		}

		crds = append(crds, crd)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk the selected folder: %w", err)
	}

	return crds, nil
}
