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

package loopsingle

import (
	"bytes"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec"
	"github.com/matrixorigin/matrixone/pkg/vm"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

func (arg *Argument) String(buf *bytes.Buffer) {
	buf.WriteString(" loop single join ")
}

func (arg *Argument) Prepare(proc *process.Process) error {
	var err error

	arg.ctr = new(container)
	arg.ctr.InitReceiver(proc, false)
	arg.ctr.bat = batch.NewWithSize(len(arg.Typs))
	for i, typ := range arg.Typs {
		arg.ctr.bat.Vecs[i] = vector.NewVec(typ)
	}

	if arg.Cond != nil {
		arg.ctr.expr, err = colexec.NewExpressionExecutor(proc, arg.Cond)
	}
	return err
}

func (arg *Argument) Call(proc *process.Process) (vm.CallResult, error) {
	anal := proc.GetAnalyze(arg.info.Idx)
	anal.Start()
	defer anal.Stop()
	ctr := arg.ctr
	result := vm.NewCallResult()
	for {
		switch ctr.state {
		case Build:
			if err := ctr.build(arg, proc, anal); err != nil {
				return result, err
			}
			ctr.state = Probe

		case Probe:
			var err error
			bat, _, err := ctr.ReceiveFromSingleReg(0, anal)
			if err != nil {
				return result, err
			}

			if bat == nil {
				ctr.state = End
				continue
			}
			if bat.IsEmpty() {
				proc.PutBatch(bat)
				continue
			}
			if ctr.bat.RowCount() == 0 {
				err = ctr.emptyProbe(bat, arg, proc, anal, arg.info.IsFirst, arg.info.IsLast, &result)
			} else {
				err = ctr.probe(bat, arg, proc, anal, arg.info.IsFirst, arg.info.IsLast, &result)
			}
			if err != nil {
				bat.Clean(proc.Mp())
			} else {
				proc.PutBatch(bat)
			}

			return result, err

		default:
			result.Batch = nil
			result.Status = vm.ExecStop
			return result, nil
		}
	}
}

func (ctr *container) build(ap *Argument, proc *process.Process, anal process.Analyze) error {
	bat, _, err := ctr.ReceiveFromSingleReg(1, anal)
	if err != nil {
		return err
	}

	if bat != nil {
		ctr.bat = bat
	}
	return nil
}

func (ctr *container) emptyProbe(bat *batch.Batch, ap *Argument, proc *process.Process, anal process.Analyze, isFirst bool, isLast bool, result *vm.CallResult) error {
	anal.Input(bat, isFirst)
	rbat := batch.NewWithSize(len(ap.Result))
	for i, rp := range ap.Result {
		if rp.Rel == 0 {
			rbat.Vecs[i] = bat.Vecs[rp.Pos]
			bat.Vecs[rp.Pos] = nil
		} else {
			rbat.Vecs[i] = vector.NewConstNull(ap.Typs[rp.Pos], bat.RowCount(), proc.Mp())
		}
	}
	rbat.SetRowCount(rbat.RowCount() + bat.RowCount())
	anal.Output(rbat, isLast)
	result.Batch = rbat
	return nil
}

func (ctr *container) probe(bat *batch.Batch, ap *Argument, proc *process.Process, anal process.Analyze, isFirst bool, isLast bool, result *vm.CallResult) error {
	anal.Input(bat, isFirst)
	rbat := batch.NewWithSize(len(ap.Result))
	for i, rp := range ap.Result {
		if rp.Rel != 0 {
			rbat.Vecs[i] = vector.NewVec(ap.Typs[rp.Pos])
		}
	}
	count := bat.RowCount()
	if ctr.expr == nil {
		switch ctr.bat.RowCount() {
		case 0:
			for i, rp := range ap.Result {
				if rp.Rel != 0 {
					err := vector.AppendMultiFixed(rbat.Vecs[i], 0, true, count, proc.Mp())
					if err != nil {
						rbat.Clean(proc.Mp())
						return err
					}
				}
			}
		case 1:
			for i, rp := range ap.Result {
				if rp.Rel != 0 {
					err := rbat.Vecs[i].UnionMulti(ctr.bat.Vecs[rp.Pos], 0, count, proc.Mp())
					if err != nil {
						rbat.Clean(proc.Mp())
						return err
					}
				}
			}
		default:
			return moerr.NewInternalError(proc.Ctx, "scalar subquery returns more than 1 row")
		}
	} else {
		if ctr.joinBat == nil {
			ctr.joinBat, ctr.cfs = colexec.NewJoinBatch(bat, proc.Mp())
		}
		for i := 0; i < count; i++ {
			if err := colexec.SetJoinBatchValues(ctr.joinBat, bat, int64(i),
				ctr.bat.RowCount(), ctr.cfs); err != nil {
				rbat.Clean(proc.Mp())
				return err
			}
			unmatched := true
			vec, err := ctr.expr.Eval(proc, []*batch.Batch{ctr.joinBat, ctr.bat})
			if err != nil {
				rbat.Clean(proc.Mp())
				return err
			}
			defer vec.Free(proc.Mp())

			rs := vector.GenerateFunctionFixedTypeParameter[bool](vec)
			if vec.IsConst() {
				b, null := rs.GetValue(0)
				if !null && b {
					if ctr.bat.RowCount() > 1 {
						return moerr.NewInternalError(proc.Ctx, "scalar subquery returns more than 1 row")
					}
					unmatched = false
					for k, rp := range ap.Result {
						if rp.Rel != 0 {
							if err := rbat.Vecs[k].UnionOne(ctr.bat.Vecs[rp.Pos], 0, proc.Mp()); err != nil {
								rbat.Clean(proc.Mp())
								return err
							}
						}
					}
				}
			} else {
				l := vec.Length()
				for j := uint64(0); j < uint64(l); j++ {
					b, null := rs.GetValue(j)
					if !null && b {
						if !unmatched {
							return moerr.NewInternalError(proc.Ctx, "scalar subquery returns more than 1 row")
						}
						unmatched = false
						for k, rp := range ap.Result {
							if rp.Rel != 0 {
								if err := rbat.Vecs[k].UnionOne(ctr.bat.Vecs[rp.Pos], int64(j), proc.Mp()); err != nil {
									rbat.Clean(proc.Mp())
									return err
								}
							}
						}
					}
				}
			}
			if unmatched {
				for k, rp := range ap.Result {
					if rp.Rel != 0 {
						if err := rbat.Vecs[k].UnionNull(proc.Mp()); err != nil {
							rbat.Clean(proc.Mp())
							return err
						}
					}
				}
			}
		}
	}
	for i, rp := range ap.Result {
		if rp.Rel == 0 {
			// rbat.Vecs[i] = bat.Vecs[rp.Pos]
			// bat.Vecs[rp.Pos] = nil
			typ := *bat.Vecs[rp.Pos].GetType()
			rbat.Vecs[i] = vector.NewVec(typ)
			if err := vector.GetUnionAllFunction(typ, proc.Mp())(rbat.Vecs[i], bat.Vecs[rp.Pos]); err != nil {
				return err
			}
		}
	}
	rbat.AddRowCount(bat.RowCount())
	anal.Output(rbat, isLast)
	result.Batch = rbat
	return nil
}
