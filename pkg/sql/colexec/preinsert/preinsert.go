// Copyright 2022 Matrix Origin
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

package preinsert

import (
	"bytes"

	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	pb "github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec"
	"github.com/matrixorigin/matrixone/pkg/sql/util"
	"github.com/matrixorigin/matrixone/pkg/vm"
)

func (arg *Argument) String(buf *bytes.Buffer) {
	buf.WriteString("pre processing insert")
}

func (arg *Argument) Prepare(_ *proc) error {
	return nil
}

func (arg *Argument) Call(proc *proc) (vm.CallResult, error) {
	result, err := arg.children[0].Call(proc)
	if err != nil {
		return result, err
	}
	if result.Batch == nil || result.Batch.IsEmpty() {
		return result, nil
	}
	bat := result.Batch

	analy := proc.GetAnalyze(arg.info.Idx)
	analy.Start()
	defer analy.Stop()
	defer proc.PutBatch(bat)

	newBat := batch.NewWithSize(len(arg.Attrs))
	newBat.Attrs = make([]string, 0, len(arg.Attrs))
	for idx := range arg.Attrs {
		newBat.Attrs = append(newBat.Attrs, arg.Attrs[idx])
		srcVec := bat.Vecs[idx]
		vec := proc.GetVector(*srcVec.GetType())
		if err := vector.GetUnionAllFunction(*srcVec.GetType(), proc.Mp())(vec, srcVec); err != nil {
			newBat.Clean(proc.Mp())
			return result, err
		}
		newBat.SetVector(int32(idx), vec)
	}
	newBat.AddRowCount(bat.RowCount())

	if arg.HasAutoCol {
		err := genAutoIncrCol(newBat, proc, arg)
		if err != nil {
			newBat.Clean(proc.GetMPool())
			return result, err
		}
	}
	// check new rows not null
	err = colexec.BatchDataNotNullCheck(newBat, arg.TableDef, proc.Ctx)
	if err != nil {
		newBat.Clean(proc.GetMPool())
		return result, err
	}

	// calculate the composite primary key column and append the result vector to batch
	err = genCompositePrimaryKey(newBat, proc, arg.TableDef)
	if err != nil {
		newBat.Clean(proc.GetMPool())
		return result, err
	}
	err = genClusterBy(newBat, proc, arg.TableDef)
	if err != nil {
		newBat.Clean(proc.GetMPool())
		return result, err
	}
	if arg.IsUpdate {
		idx := len(bat.Vecs) - 1
		newBat.Attrs = append(newBat.Attrs, catalog.Row_ID)
		rowIdVec := proc.GetVector(*bat.GetVector(int32(idx)).GetType())
		err := rowIdVec.UnionBatch(bat.Vecs[idx], 0, bat.Vecs[idx].Length(), nil, proc.Mp())
		if err != nil {
			return result, err
		}
		newBat.Vecs = append(newBat.Vecs, rowIdVec)
	}

	result.Batch = newBat
	return result, nil
}

func genAutoIncrCol(bat *batch.Batch, proc *proc, arg *Argument) error {
	lastInsertValue, err := proc.IncrService.InsertValues(
		proc.Ctx,
		arg.TableDef.TblId,
		bat)
	if err != nil {
		if moerr.IsMoErrCode(err, moerr.ErrNoSuchTable) {
			return moerr.NewNoSuchTableNoCtx(arg.SchemaName, arg.TableDef.Name)
		}
		return err
	}
	proc.SetLastInsertID(lastInsertValue)
	return nil
}

func genCompositePrimaryKey(bat *batch.Batch, proc *proc, tableDef *pb.TableDef) error {
	// Check whether the composite primary key column is included
	if tableDef.Pkey == nil || tableDef.Pkey.CompPkeyCol == nil {
		return nil
	}

	return util.FillCompositeKeyBatch(bat, catalog.CPrimaryKeyColName, tableDef.Pkey.Names, proc)
}

func genClusterBy(bat *batch.Batch, proc *proc, tableDef *pb.TableDef) error {
	if tableDef.ClusterBy == nil {
		return nil
	}
	clusterBy := tableDef.ClusterBy.Name
	if clusterBy == "" || !util.JudgeIsCompositeClusterByColumn(clusterBy) {
		return nil
	}
	return util.FillCompositeClusterByBatch(bat, clusterBy, proc)
}
