package answer

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
)

var islandAgoraTestDBOnce sync.Once

func initIslandAgoraTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	islandAgoraTestDBOnce.Do(func() {
		orm.InitDatabase()
	})
}

func newIslandAgoraTestClient(t *testing.T) *connection.Client {
	t.Helper()
	initIslandAgoraTestDB(t)

	commanderID := uint32(time.Now().UnixNano())
	if err := orm.CreateCommanderRoot(commanderID, commanderID, "Island Agora Theme Tester", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}

	commander := orm.Commander{CommanderID: commanderID}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	return &connection.Client{Commander: &commander}
}

func TestDeleteIslandAgoraThemeSuccess(t *testing.T) {
	client := newIslandAgoraTestClient(t)
	ctx := context.Background()

	err := orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		return orm.UpsertIslandAgoraThemeTx(ctx, tx, client.Commander.CommanderID, 3, "To Delete", []byte{1, 2, 3})
	})
	if err != nil {
		t.Fatalf("seed theme: %v", err)
	}

	request := protobuf.CS_21319{Id: proto.Uint32(3)}
	buffer, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	if _, _, err := DeleteIslandAgoraTheme(&buffer, client); err != nil {
		t.Fatalf("DeleteIslandAgoraTheme returned error: %v", err)
	}

	var response protobuf.SC_21320
	decodePacketAt(t, client, 0, 21320, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result, got %d", response.GetResult())
	}

	themes, err := orm.ListIslandAgoraThemes(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("list themes: %v", err)
	}
	if len(themes) != 0 {
		t.Fatalf("expected no themes after delete, got %d", len(themes))
	}
}

func TestDeleteIslandAgoraThemeNotFoundIsIdempotent(t *testing.T) {
	client := newIslandAgoraTestClient(t)

	request := protobuf.CS_21319{Id: proto.Uint32(99)}
	buffer, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	if _, _, err := DeleteIslandAgoraTheme(&buffer, client); err != nil {
		t.Fatalf("DeleteIslandAgoraTheme returned error: %v", err)
	}

	var response protobuf.SC_21320
	decodePacketAt(t, client, 0, 21320, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result, got %d", response.GetResult())
	}
}

func TestDeleteIslandAgoraThemeMalformedPayloadDoesNotMutate(t *testing.T) {
	client := newIslandAgoraTestClient(t)
	ctx := context.Background()

	err := orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		return orm.UpsertIslandAgoraThemeTx(ctx, tx, client.Commander.CommanderID, 7, "Keep", []byte{4, 5, 6})
	})
	if err != nil {
		t.Fatalf("seed theme: %v", err)
	}

	beforeCount := queryAnswerTestInt64(t, `SELECT COUNT(*) FROM island_agora_themes WHERE commander_id = $1`, client.Commander.CommanderID)

	malformed := []byte{0xff}
	_, packetID, err := DeleteIslandAgoraTheme(&malformed, client)
	if err == nil {
		t.Fatalf("expected unmarshal error")
	}
	if packetID != 21320 {
		t.Fatalf("expected packet id 21320, got %d", packetID)
	}

	afterCount := queryAnswerTestInt64(t, `SELECT COUNT(*) FROM island_agora_themes WHERE commander_id = $1`, client.Commander.CommanderID)
	if afterCount != beforeCount {
		t.Fatalf("expected row count unchanged, before=%d after=%d", beforeCount, afterCount)
	}
}

func TestDeleteIslandAgoraThemeDbFailureReturnsNonZero(t *testing.T) {
	client := newIslandAgoraTestClient(t)

	request := protobuf.CS_21319{Id: proto.Uint32(1)}
	buffer, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	originalStore := db.DefaultStore
	db.DefaultStore = nil
	defer func() { db.DefaultStore = originalStore }()

	if _, _, err := DeleteIslandAgoraTheme(&buffer, client); err != nil {
		t.Fatalf("DeleteIslandAgoraTheme returned error: %v", err)
	}

	var response protobuf.SC_21320
	decodePacketAt(t, client, 0, 21320, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result when db is unavailable")
	}
}
