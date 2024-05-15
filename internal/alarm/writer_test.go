package alarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"text/template"
)

func TestValidTemplates(t *testing.T) {
	cases := []struct {
		serviceName string
		data        *alarmData
	}{
		{
			serviceName: "sqs",
			data: &alarmData{
				Resources: map[string]any{
					"QueueName": "test",
					"DLQName":   "test-dlq",
				},
			},
		},
		{
			serviceName: "events",
			data: &alarmData{
				Resources: map[string]any{
					"RuleName":        "test",
					"EventBridgeName": "test-bridge",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.serviceName, func(t *testing.T) {

			tmpls, err := template.ParseFS(content, fmt.Sprintf("templates/%s/*", tc.serviceName))
			if err != nil {
				t.Fatal(err)
			}

			for _, tmpl := range tmpls.Templates() {
				buf := new(bytes.Buffer)

				err = tmpl.Execute(buf, tc.data)
				if err != nil {
					t.Fatal(err)
				}

				b, err := io.ReadAll(buf)
				if err != nil {
					t.Fatal(err)
				}

				if !json.Valid(b) {
					t.Logf("%s - invalid JSON: %s\n", tmpl.Name(), b)
					t.Fail()
				}
			}
		})
	}
}
