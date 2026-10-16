package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	fmt.Println("====================================================================")
	fmt.Println("             VOXASSESS COMPREHENSIVE BENCHMARK RUNNER               ")
	fmt.Println("====================================================================")
	fmt.Printf("Timestamp: %s\n", time.Now().Format(time.RFC3339))
	fmt.Println()

	suites := []struct {
		name string
		dir  string
		args []string
	}{
		{
			name: "1. Grounding Validator Benchmark (Precision, Recall, Hallucination Catch Rate)",
			dir:  "services/feedback-worker/internal/worker",
			args: []string{"test", "-v", "-run=TestGroundingValidatorBenchmark", "-bench=BenchmarkValidateGrounding"},
		},
		{
			name: "2. Sentence Boundary & Turn Latency Benchmark",
			dir:  "services/interview-orchestrator/internal/handler",
			args: []string{"test", "-v", "-run=TestOrchestratorPipeliningAndInterruption", "-bench=BenchmarkSentenceSegmentation"},
		},
		{
			name: "3. Adaptive Engine Simulation & Topic Drift Ablation",
			dir:  "services/interview-orchestrator/internal/adaptive",
			args: []string{"test", "-v", "-run=TestAdaptiveEngineAblation"},
		},
		{
			name: "4. Live Evaluation Reliability Benchmark (Groq LLM)",
			dir:  "services/feedback-worker/internal/worker",
			args: []string{"test", "-v", "-run=TestEvaluationReliabilityBenchmark", "-timeout=180s"},
		},
	}

	for _, s := range suites {
		fmt.Println(strings.Repeat("-", 68))
		fmt.Printf("RUNNING: %s\n", s.name)
		fmt.Printf("DIRECTORY: %s\n", s.dir)
		fmt.Println(strings.Repeat("-", 68))

		cmd := exec.Command("go", s.args...)
		cmd.Dir = s.dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("FAILED: %s (error: %v)\n\n", s.name, err)
		} else {
			fmt.Printf("COMPLETED: %s\n\n", s.name)
		}
	}

	fmt.Println("====================================================================")
	fmt.Println("ALL BENCHMARKS FINISHED. See docs/resume/MEASUREMENTS.md for audit trail.")
	fmt.Println("====================================================================")
}
