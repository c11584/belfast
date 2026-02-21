package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestCollectFileResponsePacketsDetectsLocalPacketIDBindings(t *testing.T) {
	file := parseTestFile(t, `
package example

func handle(client any) {
	const packetId = 10701
	client.SendMessage(packetId, nil)
}
`)

	responses := collectFileResponsePackets(file, buildImportMap(file), "example", map[string]int{})
	if !responses[10701] {
		t.Fatalf("expected packet 10701 to be detected")
	}
}

func TestCollectFileResponsePacketsDetectsBroadcastPacketIDs(t *testing.T) {
	file := parseTestFile(t, `
package example

func handle(server any, islandID uint32) {
	broadcastIslandPacket(server, islandID, 21325, nil)
	broadcastIslandPacketExcept(server, islandID, islandID, 21309, nil)
}
`)

	responses := collectFileResponsePackets(file, buildImportMap(file), "example", map[string]int{})
	if !responses[21325] {
		t.Fatalf("expected packet 21325 to be detected")
	}
	if !responses[21309] {
		t.Fatalf("expected packet 21309 to be detected")
	}
}

func TestCollectFileResponsePacketsIgnoresNonPacketBroadcastInts(t *testing.T) {
	file := parseTestFile(t, `
package example

func handle(server any) {
	broadcastSomething(server, 1, 2, 3)
}
`)

	responses := collectFileResponsePackets(file, buildImportMap(file), "example", map[string]int{})
	if len(responses) != 0 {
		t.Fatalf("expected no packet IDs, got %v", responses)
	}
}

func parseTestFile(t *testing.T, source string) *ast.File {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "test.go", source, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}
	return file
}
