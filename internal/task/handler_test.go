package task

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const (
	testSQSMessageID        = "id-123"
	testChangeDetailService = "test-service"
)

type MockAlarmCreator struct {
	Called int
	Error  error
}

func (m *MockAlarmCreator) Create(_ context.Context, _ *alarmCreationRequest) error {
	m.Called++
	return m.Error
}

func TestAlarmHandler_Handle(t *testing.T) {
	t.Parallel()

	validChangeDetail := &tagChangeDetail{Service: testChangeDetailService}

	cases := map[string]struct {
		event        func(*testing.T) *events.SQSEvent
		resp         func(*testing.T) *events.SQSEventResponse
		alarmCreator func(*testing.T) AlarmCreator
		wantError    bool
	}{
		"invalid json in sqs body returns batch failure": {
			event: func(t *testing.T) *events.SQSEvent {
				t.Helper()

				msg := events.SQSMessage{
					MessageId: testSQSMessageID,
					Body:      "invalid json",
				}

				return &events.SQSEvent{
					Records: []events.SQSMessage{msg},
				}
			},
			resp:         sqsResponseWithTestMessageID,
			alarmCreator: defaultAlarmCreator,
		},
		"invalid eb event source returns no op": {
			event: func(t *testing.T) *events.SQSEvent {
				t.Helper()

				event := eventbridgeEventWithDetail(
					&events.EventBridgeEvent{
						Source:     "invalid",
						DetailType: eventbridgeEventDetailType,
						Resources:  []string{"foo"},
					},
					validChangeDetail,
				)

				msg := sqsMessageWithEventbridgeEvent(event)

				return &events.SQSEvent{
					Records: []events.SQSMessage{msg},
				}
			},
			resp:         emptySQSResponse,
			alarmCreator: defaultAlarmCreator,
		},
		"invalid eb event detail type returns no op": {
			event: func(t *testing.T) *events.SQSEvent {
				t.Helper()

				event := eventbridgeEventWithDetail(
					&events.EventBridgeEvent{
						Source:     eventbridgeEventSource,
						DetailType: "invalid",
						Resources:  []string{"foo"},
					},
					validChangeDetail,
				)

				msg := sqsMessageWithEventbridgeEvent(event)

				return &events.SQSEvent{
					Records: []events.SQSMessage{msg},
				}
			},
			resp:         emptySQSResponse,
			alarmCreator: defaultAlarmCreator,
		},
		"invalid eb event detail returns batch failure": {
			event: func(t *testing.T) *events.SQSEvent {
				t.Helper()

				event := &events.EventBridgeEvent{
					Source:     eventbridgeEventSource,
					DetailType: eventbridgeEventDetailType,
					Resources:  []string{"foo"},
					Detail:     []byte(`{}`),
				}

				msg := sqsMessageWithEventbridgeEvent(event)

				return &events.SQSEvent{
					Records: []events.SQSMessage{msg},
				}
			},
			resp:         sqsResponseWithTestMessageID,
			alarmCreator: defaultAlarmCreator,
		},
		"invalid eb event detail service returns batch failure": {
			event: func(t *testing.T) *events.SQSEvent {
				t.Helper()

				event := eventbridgeEventWithDetail(
					&events.EventBridgeEvent{
						Source:     eventbridgeEventSource,
						DetailType: eventbridgeEventDetailType,
						Resources:  []string{"foo"},
					},
					&tagChangeDetail{Service: "invalid"},
				)

				msg := sqsMessageWithEventbridgeEvent(event)

				return &events.SQSEvent{
					Records: []events.SQSMessage{msg},
				}
			},
			resp:         sqsResponseWithTestMessageID,
			alarmCreator: defaultAlarmCreator,
		},
		"alarm creator failure returns batch failure": {
			event: func(t *testing.T) *events.SQSEvent {
				t.Helper()

				msg := sqsMessageWithEventbridgeEvent(
					validEventbridgeEventWithDetail(validChangeDetail))

				return &events.SQSEvent{
					Records: []events.SQSMessage{msg},
				}
			},
			resp: sqsResponseWithTestMessageID,
			alarmCreator: func(t *testing.T) AlarmCreator {
				t.Helper()

				return &MockAlarmCreator{Error: errors.New("alarm creation error")}
			},
		},
		"valid payload returns success": {
			event: func(t *testing.T) *events.SQSEvent {
				t.Helper()

				msg := sqsMessageWithEventbridgeEvent(
					validEventbridgeEventWithDetail(validChangeDetail))

				return &events.SQSEvent{
					Records: []events.SQSMessage{msg},
				}
			},
			resp:         emptySQSResponse,
			alarmCreator: defaultAlarmCreator,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			ctx := zerolog.New(zerolog.NewConsoleWriter(zerolog.ConsoleTestWriter(t))).WithContext(context.Background())

			creators := map[string]AlarmCreator{
				testChangeDetailService: tc.alarmCreator(t),
			}

			assert := assert.New(t)

			handler := &AlarmHandler{alarmCreators: creators}
			actual, err := handler.Handle(ctx, tc.event(t))

			if tc.wantError {
				assert.Error(err)
			}

			assert.EqualValues(tc.resp(t).BatchItemFailures, actual.BatchItemFailures)
		})
	}
}

func emptySQSResponse(t *testing.T) *events.SQSEventResponse {
	t.Helper()

	return &events.SQSEventResponse{
		BatchItemFailures: []events.SQSBatchItemFailure{},
	}
}

func sqsResponseWithTestMessageID(t *testing.T) *events.SQSEventResponse {
	t.Helper()

	return &events.SQSEventResponse{
		BatchItemFailures: []events.SQSBatchItemFailure{
			{ItemIdentifier: testSQSMessageID},
		},
	}
}

func eventbridgeEventWithDetail(eb *events.EventBridgeEvent, detail *tagChangeDetail) *events.EventBridgeEvent {
	d, err := json.Marshal(detail)
	if err != nil {
		panic(err)
	}

	return &events.EventBridgeEvent{
		Source:     eb.Source,
		DetailType: eb.DetailType,
		Resources:  eb.Resources,
		Detail:     d,
	}
}

func validEventbridgeEventWithDetail(detail *tagChangeDetail) *events.EventBridgeEvent {
	d, err := json.Marshal(detail)
	if err != nil {
		panic(err)
	}

	return &events.EventBridgeEvent{
		Source:     eventbridgeEventSource,
		DetailType: eventbridgeEventDetailType,
		Resources:  []string{"fooarn"},
		Detail:     d,
	}
}

func sqsMessageWithEventbridgeEvent(ebevent *events.EventBridgeEvent) events.SQSMessage {
	eb, err := json.Marshal(ebevent)
	if err != nil {
		panic(err)
	}

	return events.SQSMessage{
		MessageId: testSQSMessageID,
		Body:      string(eb),
	}
}

func defaultAlarmCreator(t *testing.T) AlarmCreator {
	t.Helper()

	return new(MockAlarmCreator)
}
