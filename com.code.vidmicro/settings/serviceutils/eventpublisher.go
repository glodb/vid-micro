package serviceutils

import (
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/topics"
	"com.code.vidmicro/com.code.vidmicro/settings/utils"
	"com.code.vidmicro/com.code.vidmicro/settings/utilsdatatypes"
	"github.com/bytedance/sonic"
	"github.com/rs/xid"
)

type FailSave struct {
	msg Message
	key string
}

type EventPublisher struct {
	ticker        *time.Ticker
	eventMutex    sync.Mutex
	failedMutex   sync.Mutex
	channelsQueue map[string]*utilsdatatypes.Queue
	failedQueue   utilsdatatypes.Queue
}

func (ts *EventPublisher) New() {
	ts.failedQueue.New()
	ts.channelsQueue = make(map[string]*utilsdatatypes.Queue)
	ts.startTimer()
}

func (ts *EventPublisher) startTimer() {
	ts.ticker = time.NewTicker(1 * time.Millisecond)
	go func() {
		for range ts.ticker.C {
			go ts.sendEvents()
		}
	}()
}

func (ts *EventPublisher) StopTimer() {
	ts.ticker.Stop()
	ts.sendEvents()
}

func (ts *EventPublisher) clearFailedQueue() {
	ts.failedMutex.Lock()
	tempFailedQueue := ts.failedQueue.Copy()
	ts.failedQueue.New()
	ts.failedMutex.Unlock()

	for i := range tempFailedQueue {
		faileSave := tempFailedQueue[i].(FailSave)
		if err := GetInstance().nat.Publish(faileSave.key, faileSave.msg.Body); err != nil {
			ts.failedMutex.Lock()
			ts.failedQueue.Enqueue(FailSave{msg: faileSave.msg, key: faileSave.key})
			ts.failedMutex.Unlock()
			if configmanager.GetInstance().PrintInfo {
				log.Println("("+configmanager.GetInstance().MicroServiceName+") {ClearFailedQueue}, Publish Error:", faileSave.key, ",PublishError:", err)
			}
		}
	}
}

func (ts *EventPublisher) sendBatches(element []interface{}, dedupId string, key string) {
	length := len(element)
	batches := math.Ceil(float64(length) / float64(configmanager.GetInstance().PublisherBatchSize))
	values := strings.Split(key, ":")
	for i := 0; i < int(batches); i++ {
		batchStart := i * int(configmanager.GetInstance().PublisherBatchSize)
		batchEnd := (i + 1) * int(configmanager.GetInstance().PublisherBatchSize)
		if batchEnd > length {
			batchEnd = length
		}
		body, _ := sonic.Marshal(element[batchStart:batchEnd])
		msg := &Message{
			Header: map[string]string{
				"id":      values[0],
				"dedupid": dedupId,
				"groupid": configmanager.GetInstance().MicroServiceName,
			},
			Body: body,
		}
		if err := GetInstance().nat.Publish(values[1], msg.Body); err != nil {
			ts.failedMutex.Lock()
			ts.failedQueue.Enqueue(FailSave{msg: *msg, key: values[1]})
			ts.failedMutex.Unlock()
			if configmanager.GetInstance().PrintInfo {
				log.Println("("+configmanager.GetInstance().MicroServiceName+") {sendBatches}, Publish Error:", values[0], ",PublishError:", err, ",BatchStart:", batchStart, ",BatchEnd:", batchEnd, ",Data:", element)
			}
		}
	}
}

func (ts *EventPublisher) sendEvents() {

	ts.clearFailedQueue()

	ts.eventMutex.Lock()
	newMap := utils.CopyMap(ts.channelsQueue)
	ts.channelsQueue = make(map[string]*utilsdatatypes.Queue)
	ts.eventMutex.Unlock()

	for key, element := range newMap {
		dedupId := xid.New().String()
		ts.sendBatches(element, dedupId, key)
	}
}

func (ts *EventPublisher) publishEvent(data interface{}, serviceName string, topic string) error {
	// Marshal to JSON string
	if topics.GetInstance().ValidatePublishableTopics(topic) {

		key := serviceName + ":" + topic
		ts.eventMutex.Lock()
		_, ok := ts.channelsQueue[key]

		if !ok {
			ts.channelsQueue[key] = &utilsdatatypes.Queue{}
			ts.channelsQueue[key].New()
			ts.channelsQueue[key].Enqueue(data)
		} else {
			ts.channelsQueue[key].Enqueue(data)
		}
		ts.eventMutex.Unlock()
	} else {
		if configmanager.GetInstance().PrintInfo {
			log.Println("("+configmanager.GetInstance().MicroServiceName+") {PublishEvent}, PublishError:", serviceName, " is not allowed to publish this topic", topic)
		}
	}
	return nil
}
