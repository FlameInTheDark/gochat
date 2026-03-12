package main

import "testing"

func TestPlanMessagePositionBackfillFillOnlyMissing(t *testing.T) {
	plan := planMessagePositionBackfill([]messagePositionRow{
		{ID: 10, Bucket: 1, Position: 1},
		{ID: 11, Bucket: 1, Position: 0},
		{ID: 12, Bucket: 1, Position: 4},
		{ID: 13, Bucket: 1, Position: 0},
	}, 0, false)

	if plan.Cursor != 5 {
		t.Fatalf("expected cursor 5, got %d", plan.Cursor)
	}
	if len(plan.Updates) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(plan.Updates))
	}
	if plan.Updates[0].ID != 11 || plan.Updates[0].Position != 2 {
		t.Fatalf("unexpected first update: %#v", plan.Updates[0])
	}
	if plan.Updates[1].ID != 13 || plan.Updates[1].Position != 5 {
		t.Fatalf("unexpected second update: %#v", plan.Updates[1])
	}
}

func TestPlanMessagePositionBackfillRespectsExistingCursor(t *testing.T) {
	plan := planMessagePositionBackfill([]messagePositionRow{
		{ID: 10, Bucket: 1, Position: 0},
		{ID: 11, Bucket: 1, Position: 0},
	}, 100, false)

	if plan.Cursor != 102 {
		t.Fatalf("expected cursor 102, got %d", plan.Cursor)
	}
	if len(plan.Updates) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(plan.Updates))
	}
	if plan.Updates[0].Position != 101 || plan.Updates[1].Position != 102 {
		t.Fatalf("unexpected updates: %#v", plan.Updates)
	}
}

func TestPlanMessagePositionBackfillRewrite(t *testing.T) {
	plan := planMessagePositionBackfill([]messagePositionRow{
		{ID: 10, Bucket: 1, Position: 7},
		{ID: 11, Bucket: 1, Position: 0},
		{ID: 12, Bucket: 1, Position: 7},
	}, 0, true)

	if plan.Cursor != 3 {
		t.Fatalf("expected cursor 3, got %d", plan.Cursor)
	}
	if len(plan.Updates) != 3 {
		t.Fatalf("expected 3 updates, got %d", len(plan.Updates))
	}
	for i, update := range plan.Updates {
		want := int64(i + 1)
		if update.Position != want {
			t.Fatalf("expected rewrite position %d at index %d, got %#v", want, i, update)
		}
	}
}
