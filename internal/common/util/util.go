// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package util

import (
	"fmt"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common/persistence"
	"sync"
	"time"
)

// MergeDictoRight copies the contents of src to dest
func MergeDictoRight(src map[string]string, dest map[string]string) {
	for k, v := range src {
		dest[k] = v
	}
}

// MergeDicts creates a union of the two dicts
func MergeDicts(dic1 map[string]string, dic2 map[string]string) (resultDict map[string]string) {
	resultDict = make(map[string]string)
	MergeDictoRight(dic1, resultDict)
	MergeDictoRight(dic2, resultDict)
	return
}

// AwaitWaitGroup calls Wait on the given wait
// Returns true if the Wait() call succeeded before the timeout
// Returns false if the Wait() did not return before the timeout
func AwaitWaitGroup(wg *sync.WaitGroup, timeout time.Duration) bool {

	doneC := make(chan struct{})

	go func() {
		wg.Wait()
		close(doneC)
	}()

	select {
	case <-doneC:
		return true
	case <-time.After(timeout):
		return false
	}
}

// DeserializeBlobDataToHistoryEvents deserialize the blob data to history event data
func DeserializeBlobDataToHistoryEvents(
	dataBlobs []*s.DataBlob, filterType s.HistoryEventFilterType,
) (*s.History, error) {

	var historyEvents []*s.HistoryEvent

	for _, batch := range dataBlobs {
		events, err := persistence.DeserializeBatchEvents(batch)
		if err != nil {
			return nil, err
		}
		if len(events) == 0 {
			return nil, &s.InternalServiceError{
				Message: fmt.Sprintf("corrupted history event batch, empty events"),
			}
		}

		historyEvents = append(historyEvents, events...)
	}

	if filterType == s.HistoryEventFilterTypeCloseEvent {
		historyEvents = []*s.HistoryEvent{historyEvents[len(historyEvents)-1]}
	}
	return &s.History{Events: historyEvents}, nil
}
