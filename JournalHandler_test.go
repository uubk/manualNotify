package manualNotify

import (
	"testing"
	"time"
)

func TestSingleMessage(t *testing.T) {
	src := NewChannelWriter()
	dnsHandlerChan := make(chan int, 1)
	defer close(dnsHandlerChan)
	defer src.Close()
	uut := NewJournalHandler(src.Get(), "google.ch.", false, dnsHandlerChan)
	go uut.readJournalMsg()
	ctr, _ := src.Write([]byte("zone google.ch./IN/external: sending notifies (serial 123)\n"))
	if ctr != 59 {
		t.Fail()
		t.Error("Didn't write expected number of bytes: ", ctr, ", expected: 59")
	}

	select {
	case res := <-dnsHandlerChan:
		if res != 123 {
			t.Error("Zone serial wrong: ", res, ", expected: 123")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout")
	}
}
