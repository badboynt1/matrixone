// Copyright 2021-2023 Matrix Origin
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

package table_scan

import (
	"bytes"
	"time"

	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/perfcounter"
	"github.com/matrixorigin/matrixone/pkg/sql/plan"
	"github.com/matrixorigin/matrixone/pkg/txn/client"
	"github.com/matrixorigin/matrixone/pkg/txn/trace"
	v2 "github.com/matrixorigin/matrixone/pkg/util/metric/v2"
	"github.com/matrixorigin/matrixone/pkg/vm"
	"github.com/matrixorigin/matrixone/pkg/vm/message"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

const opName = "table_scan"

func (tableScan *TableScan) String(buf *bytes.Buffer) {
	buf.WriteString(opName)
	buf.WriteString(": table_scan ")
}

func (tableScan *TableScan) OpType() vm.OpType {
	return vm.TableScan
}

func (tableScan *TableScan) Prepare(proc *process.Process) (err error) {
	if tableScan.OpAnalyzer == nil {
		tableScan.OpAnalyzer = process.NewAnalyzer(tableScan.GetIdx(), tableScan.IsFirst, tableScan.IsLast, "table_scan")
	} else {
		tableScan.OpAnalyzer.Reset()
	}

	if tableScan.TopValueMsgTag > 0 {
		tableScan.ctr.msgReceiver = message.NewMessageReceiver([]int32{tableScan.TopValueMsgTag}, tableScan.GetAddress(), proc.GetMessageBoard())
	}
	err = tableScan.PrepareProjection(proc)
	if tableScan.ctr.buf == nil {
		tableScan.ctr.buf = batch.NewWithSize(len(tableScan.Types))
		tableScan.ctr.buf.Attrs = append(tableScan.ctr.buf.Attrs, tableScan.Attrs...)
		for i := range tableScan.Types {
			tableScan.ctr.buf.Vecs[i] = vector.NewVec(plan.MakeTypeByPlan2Type(tableScan.Types[i]))
		}
	}
	return
}

func (tableScan *TableScan) Call(proc *process.Process) (vm.CallResult, error) {
	var e error
	start := time.Now()
	txnOp := proc.GetTxnOperator()
	seq := uint64(0)
	if txnOp != nil {
		seq = txnOp.NextSequence()
	}

	trace.GetService(proc.GetService()).AddTxnDurationAction(
		txnOp,
		client.TableScanEvent,
		seq,
		tableScan.TableID,
		0,
		nil)

	analyzer := tableScan.OpAnalyzer
	defer func() {
		cost := time.Since(start)

		trace.GetService(proc.GetService()).AddTxnDurationAction(
			txnOp,
			client.TableScanEvent,
			seq,
			tableScan.TableID,
			cost,
			e)
		v2.TxnStatementScanDurationHistogram.Observe(cost.Seconds())
	}()

	if err, isCancel := vm.CancelCheck(proc); isCancel {
		e = err
		return vm.CancelResult, err
	}

	for {
		// receive topvalue message
		if tableScan.ctr.msgReceiver != nil {
			msgs, _, _ := tableScan.ctr.msgReceiver.ReceiveMessage(false, proc.Ctx)
			for i := range msgs {
				msg, ok := msgs[i].(message.TopValueMessage)
				if !ok {
					panic("only support top value message in table scan!")
				}
				tableScan.Reader.SetFilterZM(msg.TopValueZM)
			}
		}

		// read data from storage engine
		tableScan.ctr.buf.CleanOnlyData()

		crs := analyzer.GetOpCounterSet()
		newCtx := perfcounter.AttachS3RequestKey(proc.Ctx, crs)
		isEnd, err := tableScan.Reader.Read(newCtx, tableScan.Attrs, nil, proc.Mp(), tableScan.ctr.buf)
		if err != nil {
			e = err
			return vm.CancelResult, err
		}
		analyzer.AddS3RequestCount(crs)
		analyzer.AddFileServiceCacheInfo(crs)
		analyzer.AddDiskIO(crs)

		if isEnd {
			e = err
			return vm.CancelResult, err
		}

		if tableScan.ctr.buf.IsEmpty() {
			continue
		}

		trace.GetService(proc.GetService()).TxnRead(
			proc.GetTxnOperator(),
			proc.GetTxnOperator().Txn().SnapshotTS,
			tableScan.TableID,
			tableScan.Attrs,
			tableScan.ctr.buf)

		analyzer.InputBlock()
		analyzer.ScanBytes(tableScan.ctr.buf)
		batSize := tableScan.ctr.buf.Size()
		tableScan.ctr.maxAllocSize = max(tableScan.ctr.maxAllocSize, batSize)
		break
	}

	analyzer.Input(tableScan.ctr.buf)
	retBatch := tableScan.ctr.buf

	return vm.CallResult{Batch: retBatch, Status: vm.ExecNext}, nil
}
