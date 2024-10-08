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

package mark

import (
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/common/reuse"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/nulls"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec"
	"github.com/matrixorigin/matrixone/pkg/sql/plan"
	"github.com/matrixorigin/matrixone/pkg/vm"
	"github.com/matrixorigin/matrixone/pkg/vm/message"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

var _ vm.Operator = new(MarkJoin)

const (
	Build = iota
	Probe
	End
)

type otyp int

const (
	condFalse otyp = iota
	condTrue
	condUnkown
)

// note that, different from other joins,the result vector is like below:
// Result[0],Result[1],......,Result[n-1],bool
// we will give more one bool type vector as a marker col
// so if you use mark join result, remember to get the last vector,that's what you want
type container struct {
	// here, we will have three states:
	// Build：we will use the right table to build a hashtable
	// Probe: we will use the left table data to probe the hashtable
	// End: Join working is over
	state int

	// store the all batch from the build table
	bat     *batch.Batch
	rbat    *batch.Batch
	joinBat *batch.Batch
	expr    colexec.ExpressionExecutor
	cfs     []func(*vector.Vector, *vector.Vector, int64, int) error

	joinBat1 *batch.Batch
	cfs1     []func(*vector.Vector, *vector.Vector, int64, int) error

	joinBat2 *batch.Batch
	cfs2     []func(*vector.Vector, *vector.Vector, int64, int) error

	// records the eval result of the batch from the probe table
	executor []colexec.ExpressionExecutor
	// vecs is same as the evecs.vec's union, we need this because
	// when we use the Insert func to build the hashtable, we need
	// vecs not evecs
	vecs []*vector.Vector

	// record those tuple that contain null value in build table
	nullSels []int64
	// record those tuple that contain normal value in build table
	sels []int64
	// the result of eval join condtion for conds[1] and cons[0], those two vectors is used to
	// check equal condition when zval == 0 or condState is False from JoinMap
	buildEqVec      []*vector.Vector
	buildEqExecutor []colexec.ExpressionExecutor

	markVals  []bool
	markNulls *nulls.Nulls

	mp *message.JoinMap

	nullWithBatch *batch.Batch
	rewriteCond   *plan.Expr

	maxAllocSize int64
}

// // for join operator, it's a two-ary operator, we will reference to two table
// // so we need this info to determine the columns to output
// // (rel,pos) gives which table and which column
// type ResultPos struct {
// 	Rel int32
// 	Pos int32
// }

// remember that we may use partition stragey, for example, if the origin table has data squence
// like 1,2,3,4. If we use the hash method, after using hash function,assume that we get 13,14,15,16.
// and we divide them into 3 buckets. so 13%3 = 1,so 3 is in the 1-th bucket and so on like this
type MarkJoin struct {
	// container means the local parameters defined by the operator constructor
	ctr container
	// the five attributes below are passed by the outside

	// // the input batch's columns' type
	// Typs []types.Type

	// records the result cols' position that the build table needs to return
	Result []int32
	// because we have two tables here,so it's a slice
	// Conditions[i] stands for the table_i's expression
	// we ned this to eval first, if there is a expression
	// like  t.a+1, we will eval it first and then to build
	// hashtable and probe to get the result. note that,though
	// we have two tables, but len(Conditions[0]) == len(Conditions[1])
	// for example, select * from t1 join t2 on t1.a = t2.d+t2.e+t2.f;
	// We will get Condition below:
	// for t1: Expr_Col --> t1.a
	// for t2: Expr_F(arg0,arg1)
	// and the arg0 is Expr_F(t2.d,t2.e)
	// and the arg1 is Expr_Col --> t2.f
	// so from the view of above, the len of two Conditions is the same
	// then you will see I just make the evals with len(Conditions[0]),but
	// I will use evalJoinConditions with parameters Conditions[0] or Conditions[1]
	// they are both ok
	Conditions [][]*plan.Expr

	Cond *plan.Expr

	OnList     []*plan.Expr
	HashOnPK   bool
	JoinMapTag int32

	vm.OperatorBase
	colexec.Projection
}

func (markJoin *MarkJoin) GetOperatorBase() *vm.OperatorBase {
	return &markJoin.OperatorBase
}

func init() {
	reuse.CreatePool[MarkJoin](
		func() *MarkJoin {
			return &MarkJoin{}
		},
		func(a *MarkJoin) {
			*a = MarkJoin{}
		},
		reuse.DefaultOptions[MarkJoin]().
			WithEnableChecker(),
	)
}

func (markJoin MarkJoin) TypeName() string {
	return opName
}

func NewArgument() *MarkJoin {
	return reuse.Alloc[MarkJoin](nil)
}

func (markJoin *MarkJoin) Release() {
	if markJoin != nil {
		reuse.Free[MarkJoin](markJoin, nil)
	}
}

func (markJoin *MarkJoin) Reset(proc *process.Process, pipelineFailed bool, err error) {
	ctr := &markJoin.ctr

	ctr.resetExecutor()
	ctr.cleanHashMap()
	ctr.resetBatch(proc.Mp())
	ctr.state = Build
	ctr.sels = nil
	ctr.nullSels = nil
	ctr.markVals = nil
	ctr.markNulls = nil

	if markJoin.ProjectList != nil {
		if markJoin.OpAnalyzer != nil {
			markJoin.OpAnalyzer.Alloc(markJoin.ProjectAllocSize + markJoin.ctr.maxAllocSize)
		}
		markJoin.ctr.maxAllocSize = 0
		markJoin.ResetProjection(proc)
	} else {
		if markJoin.OpAnalyzer != nil {
			markJoin.OpAnalyzer.Alloc(markJoin.ctr.maxAllocSize)
		}
		markJoin.ctr.maxAllocSize = 0
	}
}

func (markJoin *MarkJoin) Free(proc *process.Process, pipelineFailed bool, err error) {
	ctr := &markJoin.ctr

	ctr.cleanBatch(proc.Mp())
	ctr.cleanExecutor()

	markJoin.FreeProjection(proc)
}

func (ctr *container) resetBatch(mp *mpool.MPool) {
	if ctr.bat != nil {
		ctr.bat.Clean(mp)
		ctr.bat = nil
	}
	if ctr.nullWithBatch != nil {
		ctr.nullWithBatch.Clean(mp)
		ctr.nullWithBatch = nil
	}
}

func (ctr *container) cleanBatch(mp *mpool.MPool) {
	if ctr.rbat != nil {
		ctr.rbat.Clean(mp)
		ctr.rbat = nil
	}
	if ctr.joinBat != nil {
		ctr.joinBat.Clean(mp)
		ctr.joinBat = nil
	}
	if ctr.joinBat1 != nil {
		ctr.joinBat1.Clean(mp)
		ctr.joinBat1 = nil
	}
	if ctr.joinBat2 != nil {
		ctr.joinBat2.Clean(mp)
		ctr.joinBat2 = nil
	}
}

func (ctr *container) cleanHashMap() {
	if ctr.mp != nil {
		ctr.mp.Free()
		ctr.mp = nil
	}
}

func (ctr *container) resetExecutor() {
	if ctr.expr != nil {
		ctr.expr.ResetForNextQuery()
	}
	for i := range ctr.executor {
		if ctr.executor[i] != nil {
			ctr.executor[i].ResetForNextQuery()
		}
	}
	for i := range ctr.buildEqExecutor {
		if ctr.buildEqExecutor[i] != nil {
			ctr.buildEqExecutor[i].ResetForNextQuery()
		}
	}
}

func (ctr *container) cleanExecutor() {
	if ctr.expr != nil {
		ctr.expr.Free()
	}
	for i := range ctr.executor {
		if ctr.executor[i] != nil {
			ctr.executor[i].Free()
		}
	}
	for i := range ctr.buildEqExecutor {
		if ctr.buildEqExecutor[i] != nil {
			ctr.buildEqExecutor[i].Free()
		}
	}
}
