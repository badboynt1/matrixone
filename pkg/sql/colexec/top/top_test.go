// Copyright 2021 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package top

import (
	"bytes"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/value_scan"
	"github.com/matrixorigin/matrixone/pkg/testutil"
	"github.com/matrixorigin/matrixone/pkg/vm"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
	"github.com/stretchr/testify/require"
)

const (
	Rows          = 10     // default rows
	BenchmarkRows = 100000 // default rows for benchmark
)

// add unit tests for cases
type topTestCase struct {
	arg   *Argument
	types []types.Type
	proc  *process.Process
}

var (
	tcs []topTestCase
)

func init() {
	tcs = []topTestCase{
		newTestCase(mpool.MustNewZero(), []types.Type{types.T_int8.ToType()}, 3, []*plan.OrderBySpec{{Expr: newExpression(0), Flag: 0}}),
		newTestCase(mpool.MustNewZero(), []types.Type{types.T_int8.ToType()}, 3, []*plan.OrderBySpec{{Expr: newExpression(0), Flag: 2}}),
		newTestCase(mpool.MustNewZero(), []types.Type{types.T_int8.ToType(), types.T_int64.ToType()}, 3, []*plan.OrderBySpec{{Expr: newExpression(0), Flag: 2}, {Expr: newExpression(1), Flag: 0}}),
	}
}

func TestString(t *testing.T) {
	buf := new(bytes.Buffer)
	for _, tc := range tcs {
		tc.arg.String(buf)
	}
}

func TestPrepare(t *testing.T) {
	for _, tc := range tcs {
		err := tc.arg.Prepare(tc.proc)
		require.NoError(t, err)
	}
}

func TestTop(t *testing.T) {
	for _, tc := range tcs {
		err := tc.arg.Prepare(tc.proc)
		require.NoError(t, err)

		testBatch := newBatch(t, tc.types, tc.proc, Rows)
		resetChildren(tc.arg, testBatch)
		result, _ := tc.arg.Call(tc.proc)
		cleanResult(&result, tc.proc)

		testBatch = newBatch(t, tc.types, tc.proc, Rows)
		resetChildren(tc.arg, testBatch)
		result, _ = tc.arg.Call(tc.proc)
		cleanResult(&result, tc.proc)

		testBatch = batch.EmptyBatch
		resetChildren(tc.arg, testBatch)
		result, _ = tc.arg.Call(tc.proc)
		cleanResult(&result, tc.proc)

		testBatch = nil
		resetChildren(tc.arg, testBatch)
		result, _ = tc.arg.Call(tc.proc)
		cleanResult(&result, tc.proc)

		testBatch = nil
		resetChildren(tc.arg, testBatch)
		result, _ = tc.arg.Call(tc.proc)
		cleanResult(&result, tc.proc)

		tc.proc.FreeVectors()
		require.Equal(t, int64(0), tc.proc.Mp().CurrNB())
	}
}

func BenchmarkTop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tcs = []topTestCase{
			newTestCase(mpool.MustNewZero(), []types.Type{types.T_int8.ToType()}, 3, []*plan.OrderBySpec{{Expr: newExpression(0), Flag: 0}}),
			newTestCase(mpool.MustNewZero(), []types.Type{types.T_int8.ToType()}, 3, []*plan.OrderBySpec{{Expr: newExpression(0), Flag: 2}}),
		}
		t := new(testing.T)
		for _, tc := range tcs {
			err := tc.arg.Prepare(tc.proc)
			require.NoError(t, err)

			testBatch := newBatch(t, tc.types, tc.proc, BenchmarkRows)
			resetChildren(tc.arg, testBatch)
			result, _ := tc.arg.Call(tc.proc)
			cleanResult(&result, tc.proc)

			testBatch = newBatch(t, tc.types, tc.proc, BenchmarkRows)
			resetChildren(tc.arg, testBatch)
			result, _ = tc.arg.Call(tc.proc)
			cleanResult(&result, tc.proc)

			testBatch = batch.EmptyBatch
			resetChildren(tc.arg, testBatch)
			result, _ = tc.arg.Call(tc.proc)
			cleanResult(&result, tc.proc)

			testBatch = nil
			resetChildren(tc.arg, testBatch)
			result, _ = tc.arg.Call(tc.proc)
			cleanResult(&result, tc.proc)
		}
	}
}

func newTestCase(m *mpool.MPool, ts []types.Type, limit int64, fs []*plan.OrderBySpec) topTestCase {
	return topTestCase{
		types: ts,
		proc:  testutil.NewProcessWithMPool(m),
		arg: &Argument{
			Fs:    fs,
			Limit: limit,
			info: &vm.OperatorInfo{
				Idx:     0,
				IsFirst: false,
				IsLast:  false,
			},
		},
	}
}

func newExpression(pos int32) *plan.Expr {
	return &plan.Expr{
		Expr: &plan.Expr_Col{
			Col: &plan.ColRef{
				ColPos: pos,
			},
		},
		Typ: &plan.Type{},
	}
}

// create a new block based on the type information
func newBatch(t *testing.T, ts []types.Type, proc *process.Process, rows int64) *batch.Batch {
	return testutil.NewBatch(ts, false, int(rows), proc.Mp())
}

func resetChildren(arg *Argument, bat *batch.Batch) {
	if len(arg.children) == 0 {
		arg.AppendChild(&value_scan.Argument{
			Batchs: []*batch.Batch{bat},
		})

	} else {
		arg.children = arg.children[:0]
		arg.AppendChild(&value_scan.Argument{
			Batchs: []*batch.Batch{bat},
		})
	}
	arg.ctr.state = vm.Build
}

func cleanResult(result *vm.CallResult, proc *process.Process) {
	if result.Batch != nil {
		result.Batch.Clean(proc.Mp())
		result.Batch = nil
	}
}
