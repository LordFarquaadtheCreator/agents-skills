package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	sepIdx := -1
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--" {
			sepIdx = i
			break
		}
	}

	if sepIdx == -1 {
		fmt.Fprintln(os.Stderr, "usage: mcp-call <command...> -- <list|call|describe> [flags]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "examples:")
		fmt.Fprintln(os.Stderr, "  mcp-call /path/to/binary -- list")
		fmt.Fprintln(os.Stderr, "  mcp-call /path/to/binary -- call <tool> --args '{\"key\":\"val\"}'")
		fmt.Fprintln(os.Stderr, "  mcp-call /path/to/binary -- describe <tool>")
		fmt.Fprintln(os.Stderr, "  mcp-call docker run --rm -i create-story -- list")
		os.Exit(1)
	}

	binaryCmd := os.Args[1:sepIdx]
	mcpCallArgs := os.Args[sepIdx+1:]

	if len(binaryCmd) == 0 || len(mcpCallArgs) == 0 {
		fmt.Fprintln(os.Stderr, "error: need both command and subcommand")
		os.Exit(1)
	}

	subcommand := mcpCallArgs[0]
	rest := mcpCallArgs[1:]

	var (
		envVars  []string
		argsJSON string
		timeout  = 120 * time.Second
		toolName string
	)

	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case "--env":
			if i+1 >= len(rest) {
				fmt.Fprintln(os.Stderr, "error: --env requires a value")
				os.Exit(1)
			}
			envVars = append(envVars, rest[i+1])
			i++
		case "--args":
			if i+1 >= len(rest) {
				fmt.Fprintln(os.Stderr, "error: --args requires a value")
				os.Exit(1)
			}
			argsJSON = rest[i+1]
			i++
		case "--timeout":
			if i+1 >= len(rest) {
				fmt.Fprintln(os.Stderr, "error: --timeout requires a value")
				os.Exit(1)
			}
			d, err := time.ParseDuration(rest[i+1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: invalid timeout: %v\n", err)
				os.Exit(1)
			}
			timeout = d
			i++
		default:
			if toolName == "" {
				toolName = rest[i]
			}
		}
	}

	cmd := exec.Command(binaryCmd[0], binaryCmd[1:]...)
	cmd.Env = append(os.Environ(), envVars...)
	cmd.Stderr = os.Stderr

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "mcp-call",
		Version: "1.0.0",
	}, nil)

	transport := &mcp.CommandTransport{Command: cmd}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: connect: %v\n", err)
		os.Exit(1)
	}
	defer session.Close()

	switch subcommand {
	case "list":
		listTools(ctx, session)
	case "call":
		if toolName == "" {
			fmt.Fprintln(os.Stderr, "error: call requires a tool name")
			os.Exit(1)
		}
		callTool(ctx, session, toolName, argsJSON)
	case "describe":
		if toolName == "" {
			fmt.Fprintln(os.Stderr, "error: describe requires a tool name")
			os.Exit(1)
		}
		describeTool(ctx, session, toolName)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func listTools(ctx context.Context, session *mcp.ClientSession) {
	res, err := session.ListTools(ctx, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: list tools: %v\n", err)
		os.Exit(1)
	}

	for _, tool := range res.Tools {
		fmt.Printf("%s\n", tool.Name)
		if tool.Description != "" {
			fmt.Printf("  %s\n", tool.Description)
		}
		fmt.Println()
	}
}

func callTool(ctx context.Context, session *mcp.ClientSession, name, argsJSON string) {
	var args map[string]any
	if argsJSON != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			fmt.Fprintf(os.Stderr, "error: parse args: %v\n", err)
			os.Exit(1)
		}
	} else {
		args = map[string]any{}
	}

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: call tool: %v\n", err)
		os.Exit(1)
	}

	if res.IsError {
		for _, c := range res.Content {
			if tc, ok := c.(*mcp.TextContent); ok {
				fmt.Fprintf(os.Stderr, "error: %s\n", tc.Text)
			}
		}
		os.Exit(1)
	}

	for _, c := range res.Content {
		switch v := c.(type) {
		case *mcp.TextContent:
			fmt.Println(v.Text)
		case *mcp.ImageContent:
			ext := ".png"
			if v.MIMEType == "image/jpeg" {
				ext = ".jpg"
			}
			f, err := os.CreateTemp("", fmt.Sprintf("mcp-call-*%s", ext))
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: create temp file: %v\n", err)
				continue
			}
			if _, err := f.Write(v.Data); err != nil {
				fmt.Fprintf(os.Stderr, "error: write temp file: %v\n", err)
				f.Close()
				continue
			}
			f.Close()
			fmt.Println(f.Name())
		}
	}
}

func describeTool(ctx context.Context, session *mcp.ClientSession, name string) {
	res, err := session.ListTools(ctx, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: list tools: %v\n", err)
		os.Exit(1)
	}

	for _, tool := range res.Tools {
		if tool.Name == name {
			b, err := json.MarshalIndent(tool, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: marshal tool: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(b))
			return
		}
	}

	fmt.Fprintf(os.Stderr, "error: tool not found: %s\n", name)
	os.Exit(1)
}
