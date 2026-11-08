package nozzle

import (
	"fmt"
	"os"
	"testing"
)

// These exercise real same-process Metal metadata, not pixel transfer or
// cross-process communication. Required native setup failures are not skipped.
func metadataPair(t *testing.T, nameSuffix string) (*Sender, *Receiver, string, string) {
	t.Helper()
	name := fmt.Sprintf("go-info-%d-%s-\U0001f680", os.Getpid(), nameSuffix)
	application := "snapshot-\u65e5\u672c-\U0001f3a5"
	sender, err := NewSender(SenderDesc{Name: name, ApplicationName: application})
	if err != nil {
		t.Fatalf("real Metal sender required: %v", err)
	}
	t.Cleanup(sender.Close)
	info, err := sender.Info()
	if err != nil {
		t.Fatalf("sender metadata: %v", err)
	}
	if info.Backend != BackendMetal {
		t.Fatalf("backend = %v, want Metal", info.Backend)
	}
	receiver, err := NewReceiver(ReceiverDesc{Name: name})
	if err != nil {
		t.Fatalf("connect to real sender: %v", err)
	}
	t.Cleanup(receiver.Close)
	return sender, receiver, name, application
}

func TestObservedSenderClosureReportsDisconnected(t *testing.T) {
	sender, receiver, name, application := metadataPair(t, "closed")
	retained, err := receiver.ConnectedInfo()
	if err != nil {
		t.Fatalf("initial metadata: %v", err)
	}
	if !receiver.IsConnected() || retained.ID == "" {
		t.Fatal("receiver must initially be connected with a sender ID")
	}
	// No writable frames retain the sender. Acquire observes the closed marker
	// before waiting for a frame, even with a zero timeout.
	sender.Close()
	frame, err := receiver.AcquireFrame(0)
	if frame != nil {
		frame.Release()
		t.Fatal("destroyed sender unexpectedly supplied a frame")
	}
	t.Logf("GO_INFO_CLOSURE acquire_error=%v", err)
	if err != ErrorSenderClosed {
		t.Fatalf("acquire error = %v, want SenderClosed", err)
	}
	lastKnown, err := receiver.ConnectedInfo()
	if err != nil {
		t.Fatalf("last-known metadata: %v", err)
	}
	if lastKnown.Name != name || lastKnown.ApplicationName != application ||
		lastKnown.ID != retained.ID || lastKnown.Backend != BackendMetal {
		t.Fatalf("last-known identity changed: %+v", lastKnown)
	}
	connected := receiver.IsConnected()
	t.Logf("GO_INFO_CLOSURE last_known_id=%s connected=%t", lastKnown.ID, connected)
	if connected {
		t.Fatal("observed SenderClosed must report disconnected")
	}
}

func TestMetadataRemainsOwnedAfterHandleClose(t *testing.T) {
	sender, receiver, name, application := metadataPair(t, "owned")
	senderInfo, err := sender.Info()
	if err != nil {
		t.Fatal(err)
	}
	receiverInfo, err := receiver.ConnectedInfo()
	if err != nil {
		t.Fatal(err)
	}
	laterSender, err := sender.Info()
	if err != nil || laterSender != senderInfo {
		t.Fatalf("later sender metadata = %+v, %v", laterSender, err)
	}
	laterReceiver, err := receiver.ConnectedInfo()
	if err != nil || laterReceiver != receiverInfo {
		t.Fatalf("later receiver metadata = %+v, %v", laterReceiver, err)
	}
	receiver.Close()
	sender.Close()
	if senderInfo.Name != name || senderInfo.ApplicationName != application ||
		receiverInfo.Name != name || receiverInfo.ApplicationName != application ||
		senderInfo.ID == "" || receiverInfo.ID != senderInfo.ID ||
		senderInfo.Backend != BackendMetal || receiverInfo.Backend != BackendMetal {
		t.Fatalf("retained UTF-8 identity changed: sender=%+v receiver=%+v", senderInfo, receiverInfo)
	}
	t.Log("GO_INFO_OWNED PASS retained UTF-8 identity after both handles closed")
}
