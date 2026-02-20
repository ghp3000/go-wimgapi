//go:build windows

package wimgapi

import "testing"

func TestProgressDecoderState(t *testing.T) {
	d := NewProgressDecoder()

	ev := d.Decode(ProgressEvent{MessageID: WIMMessageSetRange, LParam: 10})
	if ev.Total != 10 {
		t.Fatalf("total=%d want 10", ev.Total)
	}

	ev = d.Decode(ProgressEvent{MessageID: WIMMessageSetPos, WParam: 3})
	if ev.Current != 3 {
		t.Fatalf("current=%d want 3", ev.Current)
	}

	ev = d.Decode(ProgressEvent{MessageID: WIMMessageStepIt})
	if ev.Current != 4 {
		t.Fatalf("current=%d want 4", ev.Current)
	}
	if ev.Percent <= 39 || ev.Percent >= 41 {
		t.Fatalf("percent=%f want around 40", ev.Percent)
	}
}

func TestProgressDecoderObservedSequence(t *testing.T) {
	d := NewProgressDecoder()

	ev := d.Decode(ProgressEvent{MessageID: WIMMessageSetRange, LParam: 5})
	if ev.Total != 5 {
		t.Fatalf("total=%d want 5", ev.Total)
	}

	for i := 0; i < 5; i++ {
		ev = d.Decode(ProgressEvent{MessageID: WIMMessageStepIt})
	}
	if ev.Current != 5 {
		t.Fatalf("current=%d want 5", ev.Current)
	}
	if ev.Percent < 99.9 {
		t.Fatalf("percent=%f want 100", ev.Percent)
	}

	ev = d.Decode(ProgressEvent{MessageID: WIMMessageSetPos, LParam: 5})
	if ev.Current != 5 || ev.Total != 5 {
		t.Fatalf("final state current=%d total=%d want 5/5", ev.Current, ev.Total)
	}
}
