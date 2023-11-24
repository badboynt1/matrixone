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

package partition

import (
	"github.com/matrixorigin/matrixone/pkg/compare"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec"
	"github.com/matrixorigin/matrixone/pkg/vm"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

var _ vm.Operator = new(Argument)

const (
	receive = iota
	eval
)

type Argument struct {
	ctr *container

	OrderBySpecs []*plan.OrderBySpec

	info     *vm.OperatorInfo
	children []vm.Operator
}

func (arg *Argument) SetInfo(info *vm.OperatorInfo) {
	arg.info = info
}

func (arg *Argument) AppendChild(child vm.Operator) {
	arg.children = append(arg.children, child)
}

type container struct {
	colexec.ReceiverOperator

	// operator status
	status int

	// batchList is the data structure to store the all the received batches
	batchList []*batch.Batch
	orderCols [][]*vector.Vector
	// indexList[i] = k means the number of rows before k in batchList[i] has been merged and send.
	indexList []int64

	// expression executors for order columns.
	executors []colexec.ExpressionExecutor
	compares  []compare.Compare

	buf *batch.Batch
}

func (arg *Argument) Free(proc *process.Process, pipelineFailed bool, err error) {
	if arg.ctr != nil {
		mp := proc.Mp()
		ctr := arg.ctr
		for i := range ctr.batchList {
			if ctr.batchList[i] != nil {
				ctr.batchList[i].Clean(mp)
			}
		}
		for i := range ctr.orderCols {
			if ctr.orderCols[i] != nil {
				for j := range ctr.orderCols[i] {
					if ctr.orderCols[i][j] != nil {
						ctr.orderCols[i][j].Free(mp)
					}
				}
			}
		}
		for i := range ctr.executors {
			if ctr.executors[i] != nil {
				ctr.executors[i].Free()
			}
		}

		if ctr.buf != nil {
			ctr.buf.Clean(proc.Mp())
			ctr.buf = nil
		}
	}
}
