package model

import "testing"

func TestIsEditableMessageType(t *testing.T) {
	if !IsEditableMessageType(MessageTypeChat) {
		t.Fatal("chat messages must remain editable")
	}
	if !IsEditableMessageType(MessageTypeReply) {
		t.Fatal("reply messages must remain editable")
	}
	if IsEditableMessageType(MessageTypeThreadCreated) {
		t.Fatal("thread-created messages must not be editable")
	}
	if IsEditableMessageType(MessageTypeThreadInitial) {
		t.Fatal("thread-initial messages must not be editable")
	}
}
