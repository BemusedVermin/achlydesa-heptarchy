package harness_test

import (
	"bytes"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

func load(t *testing.T) *content.Registry {
	t.Helper()
	bundle := contenttest.SampleBundle(t)
	contenttest.StandingCrawler(&bundle)
	return contenttest.Registry(t, bundle)
}

func runCSV(t *testing.T, reg *content.Registry, seeds, steps int) []byte {
	t.Helper()
	var buf bytes.Buffer
	cfg := harness.RunConfig{Seeds: seeds, SeedStart: 1, Steps: steps, Content: reg}
	if err := harness.Run(cfg, &buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestRunIsByteIdentical(t *testing.T) {
	reg := load(t)
	if a, b := runCSV(t, reg, 2, 15), runCSV(t, reg, 2, 15); !bytes.Equal(a, b) {
		t.Fatal("two runs with the same arguments differ")
	}
}
