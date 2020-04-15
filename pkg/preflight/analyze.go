package preflight

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob"
	analyze "github.com/replicatedhq/troubleshoot/pkg/analyze"
)

// Analyze runs the analyze phase of preflight checks
func (c CollectResult) Analyze() []*analyze.AnalyzeResult {
	getCollectedFileContents := func(fileName string) ([]byte, error) {
		contents, ok := c.AllCollectedData[fileName]
		if !ok {
			return nil, fmt.Errorf("file %s was not collected", fileName)
		}

		return contents, nil
	}
	getChildCollectedFileContents := func(prefix string) (map[string][]byte, error) {
		matching := make(map[string][]byte)
		for k, v := range c.AllCollectedData {
			if strings.HasPrefix(k, prefix) {
				matching[k] = v
			}
		}

		g, err := glob.Compile(prefix)
		// don't treat this as a true error as glob is a late addition
		if err != nil {
			return matching, nil
		}

		for k, v := range c.AllCollectedData {
			if g.Match(k) {
				matching[k] = v
			}
		}

		return matching, nil
	}

	analyzeResults := []*analyze.AnalyzeResult{}
	for _, analyzer := range c.Spec.Spec.Analyzers {
		analyzeResult, err := analyze.Analyze(analyzer, getCollectedFileContents, getChildCollectedFileContents)
		if err != nil {
			analyzeResult = []*analyze.AnalyzeResult{
				{
					IsFail:  true,
					Title:   "Analyzer Failed",
					Message: err.Error(),
				},
			}
		}

		if analyzeResult != nil {
			analyzeResults = append(analyzeResults, analyzeResult...)
		}
	}

	return analyzeResults
}
