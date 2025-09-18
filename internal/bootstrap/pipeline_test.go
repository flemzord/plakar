package bootstrap

import (
	"bytes"
	"testing"
)

func TestNewPipeline(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	pipeline := NewPipeline(stdout, stderr)

	if pipeline == nil {
		t.Fatal("NewPipeline() returned nil")
	}

	if pipeline.stdout != stdout {
		t.Error("Pipeline stdout not set correctly")
	}

	if pipeline.stderr != stderr {
		t.Error("Pipeline stderr not set correctly")
	}

	if pipeline.stages == nil {
		t.Error("Pipeline stages not initialized")
	}
}

func TestPipelineWithConfig(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	pipeline := NewPipeline(stdout, stderr)
	pipeline, err := pipeline.WithConfig()

	if err != nil {
		t.Fatalf("WithConfig() failed: %v", err)
	}

	if pipeline == nil {
		t.Fatal("WithConfig() returned nil pipeline")
	}

	// Check that config stage was added
	foundConfig := false
	for _, stage := range pipeline.stages {
		if _, ok := stage.(*ConfigStage); ok {
			foundConfig = true
			break
		}
	}

	if !foundConfig {
		t.Error("Config stage not added to pipeline")
	}
}

func TestPipelineWithProfiling(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	pipeline := NewPipeline(stdout, stderr)
	pipeline, err := pipeline.WithProfiling()

	if err != nil {
		t.Fatalf("WithProfiling() failed: %v", err)
	}

	if pipeline == nil {
		t.Fatal("WithProfiling() returned nil pipeline")
	}

	// Check that profiling stage was added
	foundProfiling := false
	for _, stage := range pipeline.stages {
		if _, ok := stage.(*ProfilingStage); ok {
			foundProfiling = true
			break
		}
	}

	if !foundProfiling {
		t.Error("Profiling stage not added to pipeline")
	}
}

func TestPipelineWithSecurity(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	pipeline := NewPipeline(stdout, stderr)
	pipeline, err := pipeline.WithSecurity()

	if err != nil {
		t.Fatalf("WithSecurity() failed: %v", err)
	}

	if pipeline == nil {
		t.Fatal("WithSecurity() returned nil pipeline")
	}

	// Check that security stage was added
	foundSecurity := false
	for _, stage := range pipeline.stages {
		if _, ok := stage.(*SecurityStage); ok {
			foundSecurity = true
			break
		}
	}

	if !foundSecurity {
		t.Error("Security stage not added to pipeline")
	}
}

func TestPipelineWithSignals(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	pipeline := NewPipeline(stdout, stderr)
	pipeline = pipeline.WithSignals()

	if pipeline == nil {
		t.Fatal("WithSignals() returned nil pipeline")
	}

	// Check that signals stage was added
	foundSignals := false
	for _, stage := range pipeline.stages {
		if _, ok := stage.(*SignalsStage); ok {
			foundSignals = true
			break
		}
	}

	if !foundSignals {
		t.Error("Signals stage not added to pipeline")
	}
}

func TestPipelineBuilderPattern(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	// Test chaining of builder methods
	pipeline := NewPipeline(stdout, stderr)
	pipeline, err := pipeline.
		WithConfig().
		WithProfiling().
		WithSecurity().
		WithSignals()

	if err != nil {
		t.Fatalf("Pipeline builder chain failed: %v", err)
	}

	if pipeline == nil {
		t.Fatal("Pipeline builder chain returned nil")
	}

	// Count stages
	if len(pipeline.stages) != 4 {
		t.Errorf("Expected 4 stages, got %d", len(pipeline.stages))
	}
}