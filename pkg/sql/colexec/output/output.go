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

package output

import (
	"bytes"

	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/vm"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

func (arg *Argument) String(buf *bytes.Buffer) {
	buf.WriteString("sql output")
}

func (arg *Argument) Prepare(_ *process.Process) error {
	return nil
}

func (arg *Argument) Call(proc *process.Process) (vm.CallResult, error) {
	ap := arg
	bat := proc.InputBatch()
	result := vm.NewCallResult()
	if bat == nil {
		result.Status = vm.ExecStop
		return result, nil
	}
	if bat.IsEmpty() {
		proc.PutBatch(bat)
		proc.SetInputBatch(batch.EmptyBatch)
		return result, nil
	}
	if err := ap.Func(ap.Data, bat); err != nil {
		bat.Clean(proc.Mp())
		result.Status = vm.ExecStop
		return result, err
	}
	proc.PutBatch(bat)
	return result, nil
}
