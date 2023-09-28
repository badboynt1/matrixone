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

	"github.com/bits-and-blooms/bloom"
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/vm"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

/*
This operator is used to implement a way to ensure primary keys/unique keys are not duplicate in `INSERT` and `LOAD` statements

the BIG idea is to store
    pk columns to be loaded
    pk columns already exist
both in a bitmap-like data structure, let's say bloom filter below

if the final bloom filter claim that
    case 1: have no duplicate keys
        pass duplicate constraint directly
    case 2: Not sure if there are duplicate keys because of hash collision
        start a background SQL to double check

Note:
1. backgroud SQL may slow, so some optimizations could be applied
    1. manually check whether collision keys duplicate or not,
        if duplicate, then return error timely
    2. shuffle between operations
2. there is a corner case that no need to run background SQL
    on duplicate key update

*/

const (
	// one million item with 10 MB(byte, not bit) and 3 hash function cnt
	// the probability of false positives  as low as 0.005%
	bloomSize = 8 * 10 * mpool.MB
	hashCnt   = 3
)

func (arg *Argument) String(buf *bytes.Buffer) {
	buf.WriteString(" fuzzy check duplicate constraint")
}

func (arg *Argument) Prepare(proc *process.Process) (err error) {
	arg.filter = bloom.New(bloomSize, hashCnt)
	return nil
}

func (arg *Argument) Call(proc *process.Process) (vm.CallResult, error) {
	anal := proc.GetAnalyze(arg.info.Idx)
	anal.Start()
	defer anal.Stop()

	if arg.state == vm.Build {
		for {
			result, err := arg.children[0].Call(proc)
			if err != nil {
				return result, err
			}
			if result.Batch == nil {
				arg.state = vm.Eval
				break
			}
			if result.Batch.IsEmpty() {
				continue
			}
			bat := result.Batch

			if arg.rbat == nil {
				if err := generateRbat(proc, arg, bat); err != nil {
					return result, err
				}
			}

			pkCol := bat.GetVector(0)
			rowCnt := bat.RowCount()
			for i := 0; i < rowCnt; i++ {
				var bytes = pkCol.GetRawBytesAt(i)
				if arg.filter.Test(bytes) {
					appendCollisionKey(proc, arg, i, bat)
					arg.collisionCnt++
				} else {
					arg.filter.Add(bytes)
				}
			}
		}
	}

	result := vm.NewCallResult()
	if arg.state == vm.Eval {
		// this will happen in such case:create unique index from a table that unique col have no data
		if arg.rbat != nil {
			arg.rbat.SetRowCount(arg.collisionCnt)
			if arg.collisionCnt > 0 {
				result.Batch = arg.rbat
				arg.rbat = nil
			}
		}
		arg.state = vm.End
		return result, nil
	}

	if arg.state == vm.End {
		return result, nil
	}

	panic("bug")
}

// appendCollisionKey will append collision key into rbat
func appendCollisionKey(proc *process.Process, arg *Argument, idx int, bat *batch.Batch) {
	pkCol := bat.GetVector(0)
	arg.rbat.GetVector(0).UnionOne(pkCol, int64(idx), proc.GetMPool())
}

// rbat will contain the keys that have hash collisions
func generateRbat(proc *process.Process, arg *Argument, bat *batch.Batch) error {
	rbat := batch.NewWithSize(1)
	rbat.SetVector(0, vector.NewVec(*bat.GetVector(0).GetType()))
	arg.rbat = rbat
	return nil
}
