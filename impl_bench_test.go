// Copyright 2022 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package event

import "testing"

func BenchmarkNew(b *testing.B) {
	benchmarkEmitter(b, New())
}

func BenchmarkNewCommon(b *testing.B) {
	benchmarkEmitter(b, NewCommon(func(matchedEvent, emittedEvent string) bool {
		return matchedEvent == emittedEvent
	}))
}

func benchmarkEmitter(b *testing.B, emitter Emitter) {
	emitter.On("event", "ln", ListenerFunc(func(string, ...interface{}) {}))

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			emitter.Emit("event")
		}
	})
}

func BenchmarkA1(b *testing.B) {
	ss := []string{"a", "b", "c", "d"}
	f := func(i int, s string) { ss[i] = s }

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			for i, s := range ss {
				f(i, s)
			}
		}
	})
}

func BenchmarkA2(b *testing.B) {
	ss := []string{"a", "b", "c", "d"}
	f := func(i int, s string) { ss[i] = s }

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			for i, _len := 0, len(ss); i < _len; i++ {
				f(i, ss[i])
			}
		}
	})
}
