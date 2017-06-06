package render

import (
	"testing"
)

func TestMetricRequest_Encode(t *testing.T) {
	tests := []struct {
		name string
		m    *MetricRequest
		want string
	}{
		{
			"MetricRequest blank",
			&MetricRequest{},
			"format=json",
		},
		{
			"MetricRequest from",
			&MetricRequest{
				From: "-5min",
			},
			"format=json&from=-5min",
		},
		{
			"MetricRequest from/until",
			&MetricRequest{
				From:  "-5min",
				Until: "-19years",
			},
			"format=json&from=-5min&until=-19years",
		},
		{
			"MetricRequest Target",
			&MetricRequest{
				Target: []string{},
			},
			"format=json",
		},
		{
			"MetricRequest Target One",
			&MetricRequest{
				Target: []string{"Fooo"},
			},
			"format=json&target=Fooo",
		},
		{
			"MetricRequest Target Manu",
			&MetricRequest{
				Target: []string{"Fooo", "Bar"},
			},
			"format=json&target=Fooo&target=Bar",
		},
		{
			"MetricRequest Full",
			&MetricRequest{
				From:   "-1year",
				Target: []string{"One", "Two"},
			},
			"format=json&from=-1year&target=One&target=Two",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Encode(); got != tt.want {
				t.Errorf("MetricRequest.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}
