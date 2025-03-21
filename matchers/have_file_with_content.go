package matchers

import (
	"archive/tar"
	"fmt"
	"io"
	"path/filepath"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func HaveFileWithContent(path, matcher interface{}) types.GomegaMatcher {
	return &haveFileWithContentMatcher{
		path:    path,
		matcher: matcher,
	}
}

type haveFileWithContentMatcher struct {
	path          interface{}
	matcher       interface{}
	actual        interface{}
	failedMatcher types.GomegaMatcher
	foundContent  string
}

func (m *haveFileWithContentMatcher) Match(actual interface{}) (bool, error) {
	m.actual = actual
	return matchImage(m.path, actual, m.matchFileContent)
}

func (m *haveFileWithContentMatcher) getMatcher() (types.GomegaMatcher, error) {
	matcher, ok := m.matcher.(types.GomegaMatcher)
	if !ok {
		str, ok := m.matcher.(string)
		if !ok {
			return nil, fmt.Errorf("expected must be a <string> or matcher, received %#v", m.matcher)
		}

		matcher = gomega.Equal(str)
	}

	m.failedMatcher = m

	return matcher, nil
}

func (m *haveFileWithContentMatcher) matchFileContent(hdr *tar.Header, tr io.Reader) (bool, error) {
	matcher, err := m.getMatcher()
	if err != nil {
		return false, err
	}
	if hdr.Typeflag == tar.TypeSymlink {
		followSymlinkPath := filepath.Join(filepath.Dir(hdr.Name), hdr.Linkname)
		return matchImage(followSymlinkPath, m.actual, m.matchFileContent)
	}
	if hdr.Typeflag != tar.TypeReg {
		m.failedMatcher = m
		return false, nil
	}

	b, err := io.ReadAll(tr)
	if err != nil {
		return false, err
	}

	m.foundContent = string(b)
	match, err := matcher.Match(m.foundContent)
	if err != nil {
		return false, err
	}

	m.failedMatcher = matcher
	return match, nil
}

func (m *haveFileWithContentMatcher) name(actual interface{}) string {
	if image, ok := actual.(v1.Image); ok {
		id, _ := image.ConfigName()
		return fmt.Sprintf("image %s", id)
	}

	if layer, ok := actual.(v1.Layer); ok {
		id, _ := layer.DiffID()
		return fmt.Sprintf("layer %s", id)
	}

	return ""
}

func (m *haveFileWithContentMatcher) FailureMessage(actual interface{}) string {
	if _, ok := m.failedMatcher.(*haveFileWithContentMatcher); ok {
		return fmt.Sprintf("Expected\n\t%s\nto have file\n\t%#v", m.name(actual), m.path)
	}

	return m.failedMatcher.FailureMessage(m.foundContent)
}

func (m *haveFileWithContentMatcher) NegatedFailureMessage(actual interface{}) string {
	if _, ok := m.failedMatcher.(*haveFileWithContentMatcher); ok {
		return fmt.Sprintf("Expected\n\t%s\nnot to have file\n\t%#v", m.name(actual), m.path)
	}

	return m.failedMatcher.NegatedFailureMessage(m.foundContent)
}
