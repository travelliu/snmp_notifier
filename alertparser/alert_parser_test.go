// Copyright 2018 Maxime Wojtczak
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package alertparser

import (
	"bytes"
	"encoding/json"
	"github.com/prometheus/alertmanager/template"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/maxwo/snmp_notifier/types"

	"github.com/go-test/deep"
)

func TestParse(t *testing.T) {
	var tests = []struct {
		DefaultOid      string
		OidLabel        string
		DefaultSeverity string
		Severities      []string
		SeverityLabel   string
		AlertsFileName  string
		BucketFileName  string
		ExpectError     bool
	}{
		{
			"1.1",
			"oid",
			"critical",
			strings.Split("critical,warning,info", ","),
			"severity",
			"test_mixed_alerts.json",
			"test_mixed_bucket.json",
			false,
		},
		{
			"1.1",
			"oid",
			"critical",
			strings.Split("critical,warning,info", ","),
			"severity",
			"test_wrong_oid_alerts.json",
			"",
			true,
		},
		{
			"1.1",
			"oid",
			"critical",
			strings.Split("critical,warning,info", ","),
			"severity",
			"test_wrong_severity_alerts.json",
			"",
			true,
		},
	}

	for _, test := range tests {
		t.Log("Testing with file", test.AlertsFileName)
		alertsByteData, err := ioutil.ReadFile(test.AlertsFileName)
		if err != nil {
			t.Fatal("Error while reading alert file:", err)
		}
		alertsReader := bytes.NewReader(alertsByteData)
		alertsData := types.AlertsData{}
		err = json.NewDecoder(alertsReader).Decode(&alertsData)
		if err != nil {
			t.Fatal("Error while parsing alert file:", err)
		}

		parserConfiguration := Configuration{test.DefaultOid, test.OidLabel, test.DefaultSeverity, test.Severities, test.SeverityLabel}
		parser := New(parserConfiguration)
		bucket, err := parser.Parse(alertsData)

		if test.ExpectError && err == nil {
			t.Error("An error was expected")
			continue
		}

		if !test.ExpectError && err != nil {
			t.Error("An unexpected error occurred:", err)
			continue
		}

		if err == nil {
			bucketByteData, err := ioutil.ReadFile(test.BucketFileName)
			if err != nil {
				t.Fatal("Error while reading bucket file:", err)
				continue
			}
			bucketReader := bytes.NewReader(bucketByteData)
			bucketData := types.AlertBucket{}
			err = json.NewDecoder(bucketReader).Decode(&bucketData)
			if err != nil {
				t.Fatal("Error while parsing bucket file:", err)
				continue
			}

			if diff := deep.Equal(bucketData, *bucket); diff != nil {
				t.Error(diff)
			}
		}
	}
}

func Test_generateInstance(t *testing.T) {
	type args struct {
		alertsData types.AlertsData
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "server_192.168.1.1:2456",
			args: args{
				alertsData: types.AlertsData{
					GroupLabels: template.KV{
						"server":   "192.168.1.1:2456",
						"instance": "192.168.1.2",
					},
				},
			},
			want: "192.168.1.1",
		},
		{
			name: "server_192.168.1.1",
			args: args{
				alertsData: types.AlertsData{
					GroupLabels: template.KV{
						"server":   "192.168.1.1",
						"instance": "192.168.1.2",
					},
				},
			},
			want: "192.168.1.1",
		},
		{
			name: "instance_192.168.1.2",
			args: args{
				alertsData: types.AlertsData{
					GroupLabels: template.KV{
						"instance": "192.168.1.2",
					},
				},
			},
			want: "192.168.1.2",
		},
		{
			name: "server_localhost",
			args: args{
				alertsData: types.AlertsData{
					GroupLabels: template.KV{
						"server":   "localhost:5432",
						"instance": "192.168.1.2",
					},
				},
			},
			want: "192.168.1.2",
		},
		{
			name: "server_127.0.0.1",
			args: args{
				alertsData: types.AlertsData{
					GroupLabels: template.KV{
						"server":   "127.0.0.1:5432",
						"instance": "192.168.1.2",
					},
				},
			},
			want: "192.168.1.2",
		},
		{
			name: "no_instance",
			args: args{
				alertsData: types.AlertsData{
					GroupLabels: template.KV{},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateInstance(tt.args.alertsData); got != tt.want {
				t.Errorf("generateInstance() = %v, want %v", got, tt.want)
			}
		})
	}
}
