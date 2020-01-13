package metrics

import (
	"github.com/DataDog/datadog-go/statsd"
)

var Statsd = mustInitStatsd()

func mustInitStatsd() *statsd.Client {
	statsd, err := statsd.New("127.0.0.1:8125")
	if err != nil {
		panic(err)
	}

	return statsd
}
