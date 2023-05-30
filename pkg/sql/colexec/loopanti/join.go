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

package loopanti

import (
	"bytes"

	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/nulls"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

func String(_ any, buf *bytes.Buffer) {
	buf.WriteString(" loop anti join ")
}

func Prepare(proc *process.Process, arg any) error {
	var err error

	ap := arg.(*Argument)
	ap.ctr = new(container)
	ap.ctr.InitReceiver(proc, false)

	if ap.Cond != nil {
		ap.ctr.expr, err = colexec.NewExpressionExecutor(proc, ap.Cond)
	}
	return err
}

func Call(idx int, proc *process.Process, arg any, isFirst bool, isLast bool) (bool, error) {
	anal := proc.GetAnalyze(idx)
	anal.Start()
	defer anal.Stop()
	ap := arg.(*Argument)
	ctr := ap.ctr
	for {
		switch ctr.state {
		case Build:
			if err := ctr.build(ap, proc, anal); err != nil {
				return false, err
			}
			ctr.state = Probe

		case Probe:
			var err error
			bat, _, err := ctr.ReceiveFromSingleReg(0, anal)
			if err != nil {
				return false, err
			}

			if bat == nil {
				ctr.state = End
				continue
			}
			if bat.Length() == 0 {
				proc.PutBatch(bat)
				continue
			}
			if ctr.bat == nil || ctr.bat.Length() == 0 {
				err = ctr.emptyProbe(bat, ap, proc, anal, isFirst, isLast)
			} else {
				err = ctr.probe(bat, ap, proc, anal, isFirst, isLast)
			}
			proc.PutBatch(bat)
			return false, err
		default:
			proc.SetInputBatch(nil)
			return true, nil
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

func (ctr *container) emptyProbe(bat *batch.Batch, ap *Argument, proc *process.Process, anal process.Analyze, isFirst bool, isLast bool) error {
	anal.Input(bat, isFirst)
	rbat := batch.NewWithSize(len(ap.Result))
	for i, pos := range ap.Result {
		// rbat.Vecs[i] = bat.Vecs[pos]
		// bat.Vecs[pos] = nil
		typ := *bat.Vecs[pos].GetType()
		rbat.Vecs[i] = vector.NewVec(typ)
		if err := vector.GetUnionAllFunction(typ, proc.Mp())(rbat.Vecs[i], bat.Vecs[pos]); err != nil {
			return err
		}
	}
	rbat.Zs = append(rbat.Zs, bat.Zs...)
	anal.Output(rbat, isLast)
	proc.SetInputBatch(rbat)
	return nil
}

func (ctr *container) probe(bat *batch.Batch, ap *Argument, proc *process.Process, anal process.Analyze, isFirst bool, isLast bool) error {
	anal.Input(bat, isFirst)
	rbat := batch.NewWithSize(len(ap.Result))
	rbat.Zs = proc.Mp().GetSels()
	for i, pos := range ap.Result {
		rbat.Vecs[i] = proc.GetVector(*bat.Vecs[pos].GetType())
	}
	count := bat.Length()
	if ctr.joinBat == nil {
		ctr.joinBat, ctr.cfs = colexec.NewJoinBatch(bat, proc.Mp())
	}
	for i := 0; i < count; i++ {
		if err := colexec.SetJoinBatchValues(ctr.joinBat, bat, int64(i),
			ctr.bat.Length(), ctr.cfs); err != nil {
			rbat.Clean(proc.Mp())
			return err
		}
		matched := false
		vec, err := ctr.expr.Eval(proc, []*batch.Batch{ctr.joinBat, ctr.bat})
		if err != nil {
			rbat.Clean(proc.Mp())
			return err
		}

		rs := vector.GenerateFunctionFixedTypeParameter[bool](vec)
		for k := uint64(0); k < uint64(vec.Length()); k++ {
			b, null := rs.GetValue(k)
			if !null && b {
				matched = true
				break
			}
		}
		if !matched && !nulls.Any(vec.GetNulls()) {
			for k, pos := range ap.Result {
				if err := rbat.Vecs[k].UnionOne(bat.Vecs[pos], int64(i), proc.Mp()); err != nil {
					rbat.Clean(proc.Mp())
					vec.Free(proc.Mp())
					return err
				}
			}
			rbat.Zs = append(rbat.Zs, bat.Zs[i])
		}
	}
	anal.Output(rbat, isLast)
	proc.SetInputBatch(rbat)
	return nil
}
