package orm

import "testing"

func TestCreateFriendRequestIsUnique(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &FriendLink{})
	clearTable(t, &FriendRequest{})
	clearTable(t, &Commander{})

	if err := CreateCommanderRoot(7001, 7001, "Requester", 0, 0); err != nil {
		t.Fatalf("create requester: %v", err)
	}
	if err := CreateCommanderRoot(7002, 7002, "Target", 0, 0); err != nil {
		t.Fatalf("create target: %v", err)
	}

	created, err := CreateFriendRequest(7001, 7002, "hello")
	if err != nil {
		t.Fatalf("create friend request: %v", err)
	}
	if !created {
		t.Fatalf("expected first request to be created")
	}

	created, err = CreateFriendRequest(7001, 7002, "hello")
	if err != nil {
		t.Fatalf("create duplicate friend request: %v", err)
	}
	if created {
		t.Fatalf("expected duplicate request to be ignored")
	}

	requests, err := ListFriendRequestsForTarget(7002)
	if err != nil {
		t.Fatalf("list requests: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected one request, got %d", len(requests))
	}
}

func TestAcceptFriendRequestCreatesBidirectionalFriendship(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &FriendLink{})
	clearTable(t, &FriendRequest{})
	clearTable(t, &Commander{})

	if err := CreateCommanderRoot(7101, 7101, "Requester", 0, 0); err != nil {
		t.Fatalf("create requester: %v", err)
	}
	if err := CreateCommanderRoot(7102, 7102, "Target", 0, 0); err != nil {
		t.Fatalf("create target: %v", err)
	}

	if _, err := CreateFriendRequest(7101, 7102, "hello"); err != nil {
		t.Fatalf("seed request: %v", err)
	}
	if err := AcceptFriendRequest(7102, 7101); err != nil {
		t.Fatalf("accept request: %v", err)
	}

	first, err := AreFriends(7101, 7102)
	if err != nil {
		t.Fatalf("check first direction: %v", err)
	}
	second, err := AreFriends(7102, 7101)
	if err != nil {
		t.Fatalf("check second direction: %v", err)
	}
	if !first || !second {
		t.Fatalf("expected bidirectional friendship")
	}

	requests, err := ListFriendRequestsForTarget(7102)
	if err != nil {
		t.Fatalf("list requests: %v", err)
	}
	if len(requests) != 0 {
		t.Fatalf("expected no pending requests after accept")
	}
}
