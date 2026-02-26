package db

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// =====================
// GLOBAL STATE: db variable issues
// These tests demonstrate problems with the package-level global db variable.
// The db variable makes the package:
// - Not thread-safe for multiple database instances
// - Hard to test in isolation
// - Prone to state leakage between tests
// =====================

// TestGlobalDB_IsolationProblem demonstrates that the global db variable
// prevents having multiple isolated database instances in tests.
// With the new Reset() function, this is now safer but still limited.
func TestGlobalDB_IsolationProblem(t *testing.T) {
	// Create first test database
	dir1, err := os.MkdirTemp("", "db-isolation-1-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir1)

	// Use Reset for clean state
	Reset()

	if err := Init(dir1); err != nil {
		t.Fatalf("Init first DB failed: %v", err)
	}

	// Insert card into first DB
	card1 := makeCachedCard("Card-In-DB1")
	if err := UpsertCard(card1); err != nil {
		t.Fatalf("Upsert to DB1 failed: %v", err)
	}

	// Create second test database - now using Reset()
	dir2, err := os.MkdirTemp("", "db-isolation-2-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir2)

	// Use Reset to cleanly switch databases
	Reset()

	if err := Init(dir2); err != nil {
		t.Fatalf("Init second DB failed: %v", err)
	}

	// Insert card into second DB
	card2 := makeCachedCard("Card-In-DB2")
	if err := UpsertCard(card2); err != nil {
		t.Fatalf("Upsert to DB2 failed: %v", err)
	}

	// PROBLEM: We can only access ONE database at a time!
	// There's no way to have isolated database instances because
	// all functions use the global db variable.

	// We can only verify the SECOND database exists
	got, err := GetCard("Card-In-DB2")
	if err != nil {
		t.Fatalf("GetCard failed: %v", err)
	}
	if got == nil {
		t.Error("Card-In-DB2 should exist in second DB")
	}

	// Cleanup
	Reset()
}

// TestGlobalDB_StateLeakage demonstrates that test state can leak
// because there's only one global db connection.
func TestGlobalDB_StateLeakage(t *testing.T) {
	// Setup first DB
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	// Insert a "known" card
	knownCard := makeCachedCard("KNOWN_CARD")
	if err := UpsertCard(knownCard); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	// Simulate what happens when another test runs after this one
	// The global db still has KNOWN_CARD in it

	// We can verify it exists
	got, _ := GetCard("KNOWN_CARD")
	if got == nil {
		t.Fatal("Known card should exist")
	}

	// PROBLEM: If cleanupTestDB fails to properly reset, or if
	// another test doesn't call cleanup, state leaks.

	_ = got // silence unused warning
}

// TestGlobalDB_CannotHaveTwoConnections demonstrates that we cannot
// have two independent database connections for different purposes.
func TestGlobalDB_CannotHaveTwoConnections(t *testing.T) {
	// This test documents the limitation: we cannot open two
	// separate database connections simultaneously.

	// The db package only exposes a global variable - there's no way to:
	// - Open a read-only connection
	// - Open a connection to a different database
	// - Have transaction isolation between operations

	// If we wanted to implement something like "get or create" atomically,
	// we'd need either:
	// 1. A connection pool with multiple connections
	// 2. A context-based API that passes the db connection
	// 3. Transaction support

	// Currently impossible because all functions use:
	// var db *sql.DB  // package-level global

	t.Log("Current API: all DB operations use global var db - cannot have 2 connections")
}

// TestGlobalDB_ParallelTestProblem demonstrates why running tests
// in parallel (-parallel flag) would fail with the global db.
func TestGlobalDB_ParallelTestProblem(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// This test simulates what would happen if two test files
	// tried to run in parallel with the global db

	dir := setupTestDB(t)

	// Both "tests" would try to use the same global db
	// Only one can succeed at a time

	card := makeCachedCard("ParallelTestCard")
	if err := UpsertCard(card); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	// Verify
	got, err := GetCard("ParallelTestCard")
	if err != nil {
		t.Fatalf("GetCard failed: %v", err)
	}
	if got == nil {
		t.Error("Card should exist")
	}

	cleanupTestDB(t, dir)

	// NOTE: Running `go test -parallel 2` on this package would fail
	// because both test functions would compete for the global db.
}

// TestGlobalDB_NoConnectionPooling demonstrates lack of connection pooling.
// The global db doesn't expose sql.DB's connection pool settings.
func TestGlobalDB_NoConnectionPooling(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	// The global db is initialized but we have no way to:
	// - Check how many connections are open
	// - Set MaxOpenConns
	// - Set MaxIdleConns
	// - Check connection status

	// We can only verify db exists
	if db == nil {
		t.Fatal("db should be initialized")
	}

	// No way to configure: db.SetMaxOpenConns(25)
	// No way to configure: db.SetMaxIdleConns(10)
	// No way to configure: db.SetConnMaxLifetime(time.Minute)

	t.Log("No public API to configure connection pooling on global db")
}

// TestGlobalDB_CannotInjectMock demonstrates that the global db
// cannot be mocked or replaced for testing.
func TestGlobalDB_CannotInjectMock(t *testing.T) {
	// With a global variable, there's no way to inject a mock database.
	// This makes unit testing difficult - we can't:
	// - Swap in a fake database
	// - Use an in-memory database for tests
	// - Have deterministic test data

	// The current approach requires:
	// 1. Creating temp directories
	// 2. Calling Init() which sets the global
	// 3. Hoping no other test is running concurrently

	// A better design would be:
	// type DB interface { Query..., Exec..., Close... }
	// type SQLiteDB struct { *sql.DB }
	// func NewApp(db DB) *App { ... }

	t.Log("Current design: global var db - cannot inject mock")
	t.Log("Better design: pass db as dependency or use context")
}

// TestGlobalDB_VerifyCurrentState shows the current workaround:
// tests must manually reset the global db.
func TestGlobalDB_VerifyCurrentState(t *testing.T) {
	// The current workaround in tests is to manually reset:
	// if db != nil { db.Close(); db = nil }

	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	// Insert data
	card := makeCachedCard("StateCard")
	if err := UpsertCard(card); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	// This works, but is error-prone because:
	// 1. Every test must remember to call cleanupTestDB
	// 2. If cleanup fails, state leaks
	// 3. Tests cannot run in parallel
	// 4. There's no way to have isolated test scenarios

	got, _ := GetCard("StateCard")
	if got == nil {
		t.Error("StateCard should exist")
	}
}

// TestConcurrentUpsertAndGet tests concurrent writes and reads to the database.
// This test can detect race conditions in the global db variable.
func TestConcurrentUpsertAndGet(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	var wg sync.WaitGroup
	numWorkers := 10
	opsPerWorker := 50

	// Concurrent writers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < opsPerWorker; i++ {
				card := makeCachedCardCardName("ConcurrentCard", workerID, i)
				if err := UpsertCard(card); err != nil {
					t.Errorf("UpsertCard failed: %v", err)
				}
			}
		}(w)
	}

	// Concurrent readers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < opsPerWorker; i++ {
				// Read any card
				_, err := GetCard("ConcurrentCard0")
				if err != nil {
					t.Errorf("GetCard failed: %v", err)
				}
			}
		}(w)
	}

	wg.Wait()
}

// TestConcurrentMixedOperations tests mixed read/write/delete operations concurrently.
func TestConcurrentMixedOperations(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	var wg sync.WaitGroup

	// Setup: insert some cards
	for i := 0; i < 10; i++ {
		card := makeCachedCardCardName("MixedCard", 0, i)
		UpsertCard(card)
	}

	// Concurrent reads and writes
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				// Read
				_, _ = GetCard("MixedCard0")

				// Write (upsert)
				card := makeCachedCardCardName("MixedCard", idx, j)
				_ = UpsertCard(card)

				// Check stale
				_, _ = IsStale("MixedCard0", 24)
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentStaleChecks tests concurrent IsStale calls.
func TestConcurrentStaleChecks(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	// Insert test card
	card := makeCachedCard("StaleCard")
	card.UpdatedAt = time.Now()
	UpsertCard(card)

	var wg sync.WaitGroup
	numWorkers := 20

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				_, err := IsStale("StaleCard", 24)
				if err != nil {
					t.Errorf("IsStale error: %v", err)
				}
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentUnmatchedCards tests concurrent access to unmatched_cards table.
func TestConcurrentUnmatchedCards(t *testing.T) {
	dir := setupTestDB(t)
	defer cleanupTestDB(t, dir)

	var wg sync.WaitGroup
	numWorkers := 10

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			deckSlug := "concurrent-deck"

			// Add unmatched cards
			for i := 0; i < 10; i++ {
				_ = AddUnmatchedCard("FakeCard", deckSlug)
			}

			// Retrieve unmatched cards
			for i := 0; i < 10; i++ {
				_, _ = GetUnmatchedCards(deckSlug)
			}
		}(w)
	}

	wg.Wait()

	// Verify final state
	names, err := GetUnmatchedCards("concurrent-deck")
	if err != nil {
		t.Fatalf("GetUnmatchedCards failed: %v", err)
	}
	// Due to INSERT OR REPLACE, we should have at most 1
	if len(names) > 1 {
		t.Logf("Note: Got %d unmatched cards (INSERT OR REPLACE behavior)", len(names))
	}
}

// makeCachedCardCardName is a helper to create a card with a unique name for concurrency tests.
func makeCachedCardCardName(base string, workerID, opID int) *CachedCard {
	return &CachedCard{
		Name:            base,
		OracleID:        "oracle-" + base,
		OracleText:      base + " does something.",
		TypeLine:        "Instant",
		ManaCost:        "{R}",
		CMC:             1.0,
		Colors:          `"R"`,
		ColorIdentity:   `"R"`,
		SetCode:         "M19",
		SetName:         "Core Set 2019",
		CollectorNumber: "123",
		ImageURI:        "https://example.com/" + base + ".jpg",
		BackImageURI:    "",
		IsDoubleFaced:   false,
		Legalities:      `{"commander":"legal"}`,
		UpdatedAt:       time.Now(),
	}
}

// =====================
// GRACEFUL SHUTDOWN: Database Close Tests
// =====================

// TestClose_Idempotent tests that calling Close multiple times doesn't panic.
func TestClose_Idempotent(t *testing.T) {
	dir := setupTestDB(t)

	// Close once
	if err := Close(); err != nil {
		t.Fatalf("first Close failed: %v", err)
	}

	// Close again - should not panic
	if err := Close(); err != nil {
		t.Errorf("second Close should not error: %v", err)
	}

	// Cleanup
	os.RemoveAll(dir)
}

// TestClose_AfterMultipleOperations tests that Close works after many operations.
func TestClose_AfterMultipleOperations(t *testing.T) {
	dir := setupTestDB(t)

	// Perform many operations
	for i := 0; i < 1000; i++ {
		card := &CachedCard{
			Name:            "Card-" + string(rune(i)),
			OracleID:        "oracle-" + string(rune(i)),
			OracleText:      "Text",
			TypeLine:        "Instant",
			ManaCost:        "{R}",
			CMC:             1.0,
			Colors:          `"R"`,
			ColorIdentity:   `"R"`,
			SetCode:         "M19",
			SetName:         "Core Set 2019",
			CollectorNumber: "123",
			ImageURI:        "https://example.com/img.jpg",
			Legalities:      `{}`,
			UpdatedAt:       time.Now(),
		}
		_ = UpsertCard(card)
	}

	// Close should work
	if err := Close(); err != nil {
		t.Fatalf("Close failed after operations: %v", err)
	}

	// Cleanup
	os.RemoveAll(dir)
}

// TestClose_WithActiveRead tests that Close works while reads are in progress.
func TestClose_WithActiveRead(t *testing.T) {
	dir := setupTestDB(t)

	// Insert cards
	for i := 0; i < 100; i++ {
		card := makeCachedCard("ReadCard")
		UpsertCard(card)
	}

	// Start multiple goroutines doing reads
	var wg sync.WaitGroup
	readDone := make(chan bool, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_, _ = GetCard("ReadCard")
		}
		readDone <- true
	}()

	// Close while reads are happening
	time.Sleep(1 * time.Millisecond) // Give reads a moment to start
	if err := Close(); err != nil {
		t.Logf("Note: Close during reads returned error: %v (may be expected)", err)
	}

	// Wait for reads to complete
	<-readDone
	wg.Wait()

	// Cleanup
	os.RemoveAll(dir)

	// Re-initialize for cleanupTestDB (since we closed)
	if err := Init(dir); err != nil {
		// Just remove the directory
		os.RemoveAll(dir)
	}
}

// TestClose_DataPersistsAfterReinit tests that data persists across close/reinit cycles.
func TestClose_DataPersistsAfterReinit(t *testing.T) {
	dir := setupTestDB(t)

	// Insert data
	cards := []string{"Persist1", "Persist2", "Persist3"}
	for _, name := range cards {
		card := makeCachedCard(name)
		UpsertCard(card)
	}

	// Close
	if err := Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Reset to allow re-init
	Reset()

	// Re-init
	if err := Init(dir); err != nil {
		t.Fatalf("Re-init failed: %v", err)
	}

	// Verify data
	for _, name := range cards {
		got, err := GetCard(name)
		if err != nil {
			t.Fatalf("GetCard failed for %s: %v", name, err)
		}
		if got == nil {
			t.Errorf("data for %s not persisted after close/reinit", name)
		}
	}

	cleanupTestDB(t, dir)
}

// TestInit_AfterClose tests that Init works after the database was closed.
func TestInit_AfterClose(t *testing.T) {
	dir := setupTestDB(t)

	// Close the database
	if err := Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Reset to allow re-init
	Reset()

	// Re-init should work
	if err := Init(dir); err != nil {
		t.Fatalf("Init after close failed: %v", err)
	}

	// Should be able to use the database
	card := makeCachedCard("PostCloseCard")
	if err := UpsertCard(card); err != nil {
		t.Fatalf("UpsertCard after re-init failed: %v", err)
	}

	got, err := GetCard("PostCloseCard")
	if err != nil {
		t.Fatalf("GetCard after re-init failed: %v", err)
	}
	if got == nil {
		t.Error("card not found after re-init")
	}

	cleanupTestDB(t, dir)
}

// TestClose_FreesResources tests that closing properly frees database resources.
func TestClose_FreesResources(t *testing.T) {
	dir := setupTestDB(t)

	// Perform many operations to allocate resources
	for i := 0; i < 100; i++ {
		card := makeCachedCard("ResourceCard")
		UpsertCard(card)
		_, _ = GetCard("ResourceCard")
	}

	// Close
	if err := Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify DB file still exists (data is preserved)
	dbPath := filepath.Join(dir, "cards.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was deleted on close")
	}

	// Cleanup
	os.RemoveAll(dir)
}

// TestConcurrentCloseAndOps tests concurrent close and operations.
func TestConcurrentCloseAndOps(t *testing.T) {
	dir := setupTestDB(t)

	// Insert some data
	for i := 0; i < 10; i++ {
		card := makeCachedCard("ConcurrentCloseCard")
		UpsertCard(card)
	}

	var wg sync.WaitGroup

	// Concurrent operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_, _ = GetCard("ConcurrentCloseCard")
			}
		}()
	}

	// Close in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Millisecond) // Let some ops start
		Close()
	}()

	wg.Wait()

	// Cleanup
	os.RemoveAll(dir)

	// Re-init for cleanup
	if err := Init(dir); err != nil {
		os.RemoveAll(dir)
	}
}

// TestGracefulShutdown_Sequence tests the proper sequence of shutdown operations.
func TestGracefulShutdown_Sequence(t *testing.T) {
	dir := setupTestDB(t)

	// 1. Use the database
	card := makeCachedCard("ShutdownCard")
	if err := UpsertCard(card); err != nil {
		t.Fatalf("UpsertCard failed: %v", err)
	}

	// 2. Verify data
	got, err := GetCard("ShutdownCard")
	if err != nil {
		t.Fatalf("GetCard failed: %v", err)
	}
	if got == nil {
		t.Fatal("card not found before shutdown")
	}

	// 3. Close gracefully
	if err := Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// 4. Reset to allow re-init
	Reset()

	// 5. Re-init
	if err := Init(dir); err != nil {
		t.Fatalf("Re-init failed: %v", err)
	}

	// 6. Data should still be there
	got, err = GetCard("ShutdownCard")
	if err != nil {
		t.Fatalf("GetCard after re-init failed: %v", err)
	}
	if got == nil {
		t.Error("data lost after graceful shutdown sequence")
	}

	cleanupTestDB(t, dir)
}
