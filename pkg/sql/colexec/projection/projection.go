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

package projection

import (
	"bytes"

	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

func (arg *Argument) String(buf *bytes.Buffer) {
	n := arg
	buf.WriteString("projection(")
	for i, e := range n.Es {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(e.String())
	}
	buf.WriteString(")")
}

func (arg *Argument) Prepare(proc *process.Process) (err error) {
	ap := arg
	ap.ctr = new(container)
	ap.ctr.projExecutors, err = colexec.NewExpressionExecutorsFromPlanExpressions(proc, ap.Es)

	return err
}

func (arg *Argument) Call(idx int, proc *process.Process, isFirst bool, isLast bool) (process.ExecStatus, error) {
	anal := proc.GetAnalyze(idx)
	anal.Start()
	defer anal.Stop()

	bat := proc.InputBatch()
	if bat == nil {
		proc.SetInputBatch(nil)
		return process.ExecStop, nil
	}
	if bat.Last() {
		proc.SetInputBatch(bat)
		return process.ExecNext, nil
	}
	if bat.IsEmpty() {
		proc.PutBatch(bat)
		proc.SetInputBatch(batch.EmptyBatch)
		return process.ExecNext, nil
	}

	anal.Input(bat, isFirst)
	ap := arg
	rbat := batch.NewWithSize(len(ap.Es))

	// do projection.
	for i := range ap.ctr.projExecutors {
		vec, err := ap.ctr.projExecutors[i].Eval(proc, []*batch.Batch{bat})
		if err != nil {
			return process.ExecNext, err
		}
		rbat.Vecs[i] = vec
	}

	newAlloc, err := colexec.FixProjectionResult(proc, ap.ctr.projExecutors, rbat, bat)
	if err != nil {
		bat.Clean(proc.Mp())
		return process.ExecNext, err
	}
	anal.Alloc(int64(newAlloc))

	rbat.SetRowCount(bat.RowCount())

	proc.PutBatch(bat)
	anal.Output(rbat, isLast)
	proc.SetInputBatch(rbat)
	return process.ExecNext, nil
}
