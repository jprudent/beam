// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package primitives

import (
	"fmt"

	"github.com/apache/beam/sdks/go/pkg/beam"
	"github.com/apache/beam/sdks/go/pkg/beam/testing/passert"
)

func genA(_ []byte, emit func(string, int)) {
	emit("a", 1)
	emit("a", 2)
	emit("a", 3)
	emit("b", 4)
	emit("b", 5)
	emit("c", 6)
}

func genB(_ []byte, emit func(string, int)) {
	emit("a", 7)
	emit("b", 8)
	emit("d", 9)
}

func genC(_ []byte, emit func(string, string)) {
	emit("a", "alpha")
	emit("c", "charlie")
	emit("d", "delta")
}

func sum(nums func(*int) bool) int {
	var ret, i int
	for nums(&i) {
		ret += i
	}
	return ret
}

func lenSum(strings func(*string) bool) int {
	var ret int
	var s string
	for strings(&s) {
		ret += len(s)
	}
	return ret
}

func joinFn(key string, as, bs func(*int) bool, cs func(*string) bool, emit func(string, int)) {
	emit(key, sum(as)+sum(bs)+lenSum(cs))
}

func splitFn(key string, v int, a, b, c, d func(int)) {
	switch key {
	case "a":
		a(v)
	case "b":
		b(v)
	case "c":
		c(v)
	case "d":
		d(v)
	default:
		panic(fmt.Sprintf("bad key: %v", key))
	}
}

// CoGBK tests CoGBK.
func CoGBK() *beam.Pipeline {
	p, s := beam.NewPipelineWithRoot()

	as := beam.ParDo(s, genA, beam.Impulse(s))
	bs := beam.ParDo(s, genB, beam.Impulse(s))
	cs := beam.ParDo(s, genC, beam.Impulse(s))

	grouped := beam.CoGroupByKey(s, as, bs, cs)
	joined := beam.ParDo(s, joinFn, grouped)
	a, b, c, d := beam.ParDo4(s, splitFn, joined)

	passert.Sum(s, a, "a", 1, 18)
	passert.Sum(s, b, "b", 1, 17)
	passert.Sum(s, c, "c", 1, 13)
	passert.Sum(s, d, "d", 1, 14)

	return p
}
