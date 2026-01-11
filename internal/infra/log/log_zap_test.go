package log_test

import (
	"testing"

	"go.uber.org/zap"
)

func BenchmarkLogZapWith(b *testing.B) {
	var (
		msgtext = "1234567890"
	)

	ss := zap.L()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ss.Info(msgtext)
	}
}
