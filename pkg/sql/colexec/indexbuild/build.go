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

package indexbuild

import (
	"bytes"

	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/pb/pipeline"
	"github.com/matrixorigin/matrixone/pkg/vm"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

const argName = "index_build"

func (arg *Argument) String(buf *bytes.Buffer) {
	buf.WriteString(argName)
	buf.WriteString(": index build ")
}

func (arg *Argument) Prepare(proc *process.Process) (err error) {
	ap := arg
	ap.ctr = new(container)
	if len(proc.Reg.MergeReceivers) > 1 {
		ap.ctr.InitReceiver(proc, true)
		ap.ctr.isMerge = true
	} else {
		ap.ctr.InitReceiver(proc, false)
	}

	return nil
}

func (arg *Argument) Call(proc *process.Process) (vm.CallResult, error) {
	if err, isCancel := vm.CancelCheck(proc); isCancel {
		return vm.CancelResult, err
	}

	anal := proc.GetAnalyze(arg.GetIdx(), arg.GetParallelIdx(), arg.GetParallelMajor())
	anal.Start()
	defer anal.Stop()
	result := vm.NewCallResult()
	ap := arg
	ctr := ap.ctr
	for {
		switch ctr.state {
		case ReceiveBatch:
			if err := ctr.build(ap, proc, anal, arg.GetIsFirst()); err != nil {
				return result, err
			}
			ctr.state = HandleRuntimeFilter

		case HandleRuntimeFilter:
			if err := ctr.handleRuntimeFilter(ap, proc); err != nil {
				return result, err
			}
			ctr.state = End
		default:
			if ctr.batch != nil {
				proc.PutBatch(ctr.batch)
				ctr.batch = nil
			}
			result.Batch = nil
			result.Status = vm.ExecStop
			return result, nil
		}
	}
}

func (ctr *container) collectBuildBatches(ap *Argument, proc *process.Process, anal process.Analyze, isFirst bool) error {
	var err error
	var currentBatch *batch.Batch
	for {
		if ap.ctr.isMerge {
			currentBatch, _, err = ctr.ReceiveFromAllRegs(anal)
		} else {
			currentBatch, _, err = ctr.ReceiveFromSingleReg(0, anal)
		}
		if err != nil {
			return err
		}
		if currentBatch == nil {
			break
		}
		if currentBatch.IsEmpty() {
			proc.PutBatch(currentBatch)
			continue
		}
		anal.Input(currentBatch, isFirst)
		anal.Alloc(int64(currentBatch.Size()))
		ctr.batch, _, err = proc.AppendBatchFromOffset(ctr.batch, currentBatch, 0)
		if err != nil {
			return err
		}
		proc.PutBatch(currentBatch)
	}
	return nil
}

func (ctr *container) build(ap *Argument, proc *process.Process, anal process.Analyze, isFirst bool) error {
	err := ctr.collectBuildBatches(ap, proc, anal, isFirst)
	if err != nil {
		return err
	}
	return nil
}

func (ctr *container) handleRuntimeFilter(ap *Argument, proc *process.Process) error {
	if len(ap.RuntimeFilterSenders) == 0 {
		panic("there must be runtime filter in index build!")
		return nil
	}

	var runtimeFilter *pipeline.RuntimeFilter

	if ap.RuntimeFilterSenders[0].Spec.Expr == nil {
		runtimeFilter = &pipeline.RuntimeFilter{
			Typ: pipeline.RuntimeFilter_PASS,
		}
	} else if ctr.batch == nil || ctr.batch.RowCount() == 0 {
		runtimeFilter = &pipeline.RuntimeFilter{
			Typ: pipeline.RuntimeFilter_DROP,
		}
	}

	if runtimeFilter != nil {
		select {
		case <-proc.Ctx.Done():
			ctr.state = End
		case ap.RuntimeFilterSenders[0].Chan <- runtimeFilter:
			ctr.state = End
		}
		return nil
	}

	inFilterCardLimit := ap.RuntimeFilterSenders[0].Spec.UpperLimit

	if ctr.batch.RowCount() > int(inFilterCardLimit) {
		runtimeFilter = &pipeline.RuntimeFilter{
			Typ: pipeline.RuntimeFilter_PASS,
		}
	} else {
		if len(ctr.batch.Vecs) != 1 {
			panic("there must be only 1 vector in index build batch")
		}
		vec := ctr.batch.Vecs[0]
		vec.InplaceSort()
		data, err := vec.MarshalBinary()
		if err != nil {
			return err
		}

		runtimeFilter = &pipeline.RuntimeFilter{
			Typ:  pipeline.RuntimeFilter_IN,
			Card: int32(vec.Length()),
			Data: data,
		}
	}
	select {
	case <-proc.Ctx.Done():
		ctr.state = End
	case ap.RuntimeFilterSenders[0].Chan <- runtimeFilter:
		ctr.state = End
	}
	return nil
}
