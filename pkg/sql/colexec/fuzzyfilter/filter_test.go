// Copyright 2023 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package fuzzyfilter

import (
	"bytes"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/value_scan"
	"github.com/matrixorigin/matrixone/pkg/testutil"
	"github.com/matrixorigin/matrixone/pkg/vm"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
	"github.com/stretchr/testify/require"
)

// can be used to check hash collision rate
// when the number of tests is below 200,000, the rate of false positives should be very low
const (
	rowCnt = 200000
	fkCnt  = 1
)

// add unit tests for cases
type fuzzyTestCase struct {
	arg   *Argument
	types []types.Type
	proc  *process.Process
}

var (
	tcs []fuzzyTestCase
)

func init() {
	tcs = []fuzzyTestCase{
		{
			arg: &Argument{
				info: &vm.OperatorInfo{
					Idx:     1,
					IsFirst: false,
					IsLast:  false,
				},
			},
			proc: testutil.NewProcessWithMPool(mpool.MustNewZero()),
			types: []types.Type{
				types.T_int32.ToType(),
			},
		},
		{
			arg: &Argument{
				info: &vm.OperatorInfo{
					Idx:     1,
					IsFirst: false,
					IsLast:  false,
				},
			},
			proc: testutil.NewProcessWithMPool(mpool.MustNewZero()),
			types: []types.Type{
				types.T_date.ToType(),
			},
		},
		{
			arg: &Argument{
				info: &vm.OperatorInfo{
					Idx:     1,
					IsFirst: false,
					IsLast:  false,
				},
			},
			proc: testutil.NewProcessWithMPool(mpool.MustNewZero()),
			types: []types.Type{
				types.T_float32.ToType(),
			},
		},
		{
			arg: &Argument{
				info: &vm.OperatorInfo{
					Idx:     1,
					IsFirst: false,
					IsLast:  false,
				},
			},
			proc: testutil.NewProcessWithMPool(mpool.MustNewZero()),
			types: []types.Type{
				types.T_varchar.ToType(),
			},
		},
		{
			arg: &Argument{
				info: &vm.OperatorInfo{
					Idx:     1,
					IsFirst: false,
					IsLast:  false,
				},
			},
			proc: testutil.NewProcessWithMPool(mpool.MustNewZero()),
			types: []types.Type{
				types.T_binary.ToType(),
			},
		},
	}
}

func TestString(t *testing.T) {
	for _, tc := range tcs {
		buf := new(bytes.Buffer)
		tc.arg.String(buf)
		require.Equal(t, " fuzzy check duplicate constraint", buf.String())
	}
}

func TestPrepare(t *testing.T) {
	for _, tc := range tcs {
		err := tc.arg.Prepare(tc.proc)
		require.NoError(t, err)
	}
}

func TestFuzzyFilter(t *testing.T) {
	for _, tc := range tcs {
		err := tc.arg.Prepare(tc.proc)
		require.NoError(t, err)

		testBatch := newBatch(t, tc.types, tc.proc, rowCnt)
		resetChildren(tc.arg, testBatch)
		result, err := tc.arg.Call(tc.proc)
		cleanResult(&result, tc.proc)

		require.NoError(t, err)
		require.LessOrEqual(t, tc.arg.collisionCnt, fkCnt, "collision cnt is too high, sth must went wrong")
	}
}

// create a new block based on the type information
func newBatch(t *testing.T, ts []types.Type, proc *process.Process, rows int64) *batch.Batch {
	// not random
	bat := testutil.NewBatch(ts, false, int(rows), proc.Mp())
	pkAttr := make([]string, 1)
	pkAttr[0] = "pkCol"
	bat.SetAttributes(pkAttr)
	return bat
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
	arg.state = vm.Build
}

func cleanResult(result *vm.CallResult, proc *process.Process) {
	if result.Batch != nil {
		result.Batch.Clean(proc.Mp())
		result.Batch = nil
	}
}
