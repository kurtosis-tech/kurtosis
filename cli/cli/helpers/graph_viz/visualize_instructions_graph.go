package graph_viz

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-graphviz"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/dependency_graph"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	graphFilePermissions = 0644
)

// OutputGraphVisual generates a Graphviz DOT format graph and renders it as a PNG image
func OutputGraphVisual(instructions []dependency_graph.InstructionWithDependencies, path string) error {
	if len(instructions) == 0 {
		return stacktrace.NewError("No instructions provided to generate graph")
	}

	var dotGraph strings.Builder

	// Start the DOT graph
	dotGraph.WriteString("digraph KurtosisInstructions {\n")
	dotGraph.WriteString("  rankdir=TB;\n")
	dotGraph.WriteString("  node [shape=box, style=rounded];\n\n")

	// Add nodes with labels (UUIDs are used only for internal node IDs, not displayed)
	for _, instruction := range instructions {
		if instruction.IsPrintInstruction {
			continue
		}
		nodeLabel := escapeLabel(instruction.ShortDescriptor)
		dotGraph.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\"];\n",
			instruction.InstructionUuid, nodeLabel))
	}

	dotGraph.WriteString("\n")

	// Add edges (dependencies) - skip print instructions
	for _, instruction := range instructions {
		if instruction.IsPrintInstruction {
			continue
		}
		for _, dependency := range instruction.Dependencies {
			dotGraph.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n",
				dependency, instruction.InstructionUuid))
		}
	}

	dotGraph.WriteString("}\n")

	// Create graphviz instance
	ctx := context.Background()
	g, err := graphviz.New(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create graphviz instance")
	}
	defer func() {
		if err := g.Close(); err != nil {
			// Log the error but don't fail the function
			fmt.Fprintf(os.Stderr, "Warning: failed to close graphviz instance: %v\n", err)
		}
	}()

	// Parse the DOT string
	graph, err := graphviz.ParseBytes([]byte(dotGraph.String()))
	if err != nil {
		return stacktrace.Propagate(err, "Failed to parse DOT graph")
	}
	defer func() {
		if err := graph.Close(); err != nil {
			// Log the error but don't fail the function
			fmt.Fprintf(os.Stderr, "Warning: failed to close graph: %v\n", err)
		}
	}()

	// Render the graph to PNG format
	var buf bytes.Buffer
	if err := g.Render(ctx, graph, graphviz.PNG, &buf); err != nil {
		return stacktrace.Propagate(err, "Failed to render graph to PNG")
	}

	// Write the PNG to file
	err = os.WriteFile(path, buf.Bytes(), graphFilePermissions)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to write PNG graph to file '%s'", path)
	}

	return nil
}

// OutputMermaidGraph generates a Mermaid format graph and saves it to the specified path
func OutputMermaidGraph(instructions []dependency_graph.InstructionWithDependencies, path string) error {
	if len(instructions) == 0 {
		return stacktrace.NewError("No instructions provided to generate mermaid graph")
	}

	var mermaidGraph strings.Builder

	// Start the Mermaid graph
	mermaidGraph.WriteString("```mermaid\n")
	mermaidGraph.WriteString("graph TD\n")

	// Add nodes with labels (skip print instructions)
	for _, instruction := range instructions {
		if instruction.IsPrintInstruction {
			continue
		}
		nodeLabel := escapeLabel(instruction.ShortDescriptor)
		nodeId := sanitizeNodeId(string(instruction.InstructionUuid))
		mermaidGraph.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", nodeId, nodeLabel))
	}

	mermaidGraph.WriteString("\n")

	// Add edges (dependencies) - skip print instructions
	for _, instruction := range instructions {
		if instruction.IsPrintInstruction {
			continue
		}
		targetNodeId := sanitizeNodeId(string(instruction.InstructionUuid))
		for _, dependency := range instruction.Dependencies {
			sourceNodeId := sanitizeNodeId(string(dependency))
			mermaidGraph.WriteString(fmt.Sprintf("  %s --> %s\n", sourceNodeId, targetNodeId))
		}
	}

	mermaidGraph.WriteString("```\n")

	// Write to file
	err := os.WriteFile(path, []byte(mermaidGraph.String()), graphFilePermissions)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to write Mermaid graph to file '%s'", path)
	}

	return nil
}

// escapeLabel escapes special characters in labels for graph formats
func escapeLabel(label string) string {
	// Replace quotes with escaped quotes
	label = strings.ReplaceAll(label, "\"", "\\\"")
	// Replace newlines with spaces
	label = strings.ReplaceAll(label, "\n", " ")
	return label
}

// sanitizeNodeId converts a UUID or identifier to a valid Mermaid node ID
// Mermaid node IDs should not contain hyphens or special characters
func sanitizeNodeId(id string) string {
	// Replace hyphens and other special characters with underscores
	id = strings.ReplaceAll(id, "-", "_")
	id = strings.ReplaceAll(id, ".", "_")
	id = strings.ReplaceAll(id, ":", "_")
	return id
}
