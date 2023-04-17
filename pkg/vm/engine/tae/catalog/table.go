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

package catalog

import (
	"bytes"
	"fmt"
	"sync/atomic"

	pkgcatalog "github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/iface/data"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/iface/txnif"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/txn/txnbase"
)

type TableDataFactory = func(meta *TableEntry) data.Table

func tableVisibilityFn[T *TableEntry](n *common.GenericDLNode[*TableEntry], txn txnif.TxnReader) (visible, dropped bool) {
	table := n.GetPayload()
	visible, dropped = table.GetVisibility(txn)
	return
}

type TableEntry struct {
	*BaseEntryImpl[*TableMVCCNode]
	*TableNode
	ID      uint64
	db      *DBEntry
	entries map[types.Uuid]*common.GenericDLNode[*SegmentEntry]
	//link.head and link.tail is nil when create tableEntry object.
	link      *common.GenericSortedDList[*SegmentEntry]
	tableData data.Table
	rows      atomic.Uint64
	// fullname is format as 'tenantID-tableName', the tenantID prefix is only used 'mo_catalog' database
	fullName string
}

func genTblFullName(tenantID uint32, name string) string {
	if name == pkgcatalog.MO_DATABASE || name == pkgcatalog.MO_TABLES || name == pkgcatalog.MO_COLUMNS {
		tenantID = 0
	}
	return fmt.Sprintf("%d-%s", tenantID, name)
}

func NewTableEntry(db *DBEntry, schema *Schema, txnCtx txnif.AsyncTxn, dataFactory TableDataFactory) *TableEntry {
	id := db.catalog.NextTable()
	return NewTableEntryWithTableId(db, schema, txnCtx, dataFactory, id)
}

func NewTableEntryWithTableId(db *DBEntry, schema *Schema, txnCtx txnif.AsyncTxn, dataFactory TableDataFactory, tableId uint64) *TableEntry {
	if txnCtx != nil {
		// Only in unit test, txnCtx can be nil
		schema.AcInfo.TenantID = txnCtx.GetTenantID()
		schema.AcInfo.UserID, schema.AcInfo.RoleID = txnCtx.GetUserAndRoleID()
	}
	schema.AcInfo.CreateAt = types.CurrentTimestamp()
	e := &TableEntry{
		ID: tableId,
		BaseEntryImpl: NewBaseEntry(
			func() *TableMVCCNode { return &TableMVCCNode{} }),
		db: db,
		TableNode: &TableNode{
			schema: schema,
		},
		link:    common.NewGenericSortedDList(compareSegmentFn),
		entries: make(map[types.Uuid]*common.GenericDLNode[*SegmentEntry]),
	}
	if dataFactory != nil {
		e.tableData = dataFactory(e)
	}
	e.CreateWithTxnAndSchema(txnCtx, schema)
	return e
}

func NewSystemTableEntry(db *DBEntry, id uint64, schema *Schema) *TableEntry {
	e := &TableEntry{
		ID: id,
		BaseEntryImpl: NewBaseEntry(
			func() *TableMVCCNode { return &TableMVCCNode{} }),
		db: db,
		TableNode: &TableNode{
			schema: schema,
		},
		link:    common.NewGenericSortedDList(compareSegmentFn),
		entries: make(map[types.Uuid]*common.GenericDLNode[*SegmentEntry]),
	}
	e.CreateWithTS(types.SystemDBTS, &TableMVCCNode{})
	var sid types.Uuid
	if schema.Name == SystemTableSchema.Name {
		sid = SystemSegment_Table_ID
	} else if schema.Name == SystemDBSchema.Name {
		sid = SystemSegment_DB_ID
	} else if schema.Name == SystemColumnSchema.Name {
		sid = SystemSegment_Columns_ID
	} else {
		panic("not supported")
	}
	segment := NewSysSegmentEntry(e, sid)
	e.AddEntryLocked(segment)
	return e
}

func NewReplayTableEntry() *TableEntry {
	e := &TableEntry{
		BaseEntryImpl: NewReplayBaseEntry(
			func() *TableMVCCNode { return &TableMVCCNode{} }),
		link:    common.NewGenericSortedDList(compareSegmentFn),
		entries: make(map[types.Uuid]*common.GenericDLNode[*SegmentEntry]),
	}
	return e
}

func MockStaloneTableEntry(id uint64, schema *Schema) *TableEntry {
	return &TableEntry{
		ID: id,
		BaseEntryImpl: NewBaseEntry(
			func() *TableMVCCNode { return &TableMVCCNode{} }),
		TableNode: &TableNode{
			schema: schema,
		},
		link:    common.NewGenericSortedDList(compareSegmentFn),
		entries: make(map[types.Uuid]*common.GenericDLNode[*SegmentEntry]),
	}
}
func (entry *TableEntry) GetID() uint64 { return entry.ID }
func (entry *TableEntry) IsVirtual() bool {
	if !entry.db.IsSystemDB() {
		return false
	}
	return entry.schema.Name == pkgcatalog.MO_DATABASE ||
		entry.schema.Name == pkgcatalog.MO_TABLES ||
		entry.schema.Name == pkgcatalog.MO_COLUMNS
}

func (entry *TableEntry) GetRows() uint64 {
	return entry.rows.Load()
}

func (entry *TableEntry) AddRows(delta uint64) uint64 {
	return entry.rows.Add(delta)
}

func (entry *TableEntry) RemoveRows(delta uint64) uint64 {
	return entry.rows.Add(^(delta - 1))
}

func (entry *TableEntry) GetSegmentByID(id types.Uuid) (seg *SegmentEntry, err error) {
	entry.RLock()
	defer entry.RUnlock()
	node := entry.entries[id]
	if node == nil {
		return nil, moerr.GetOkExpectedEOB()
	}
	return node.GetPayload(), nil
}

func (entry *TableEntry) MakeSegmentIt(reverse bool) *common.GenericSortedDListIt[*SegmentEntry] {
	entry.RLock()
	defer entry.RUnlock()
	return common.NewGenericSortedDListIt(entry.RWMutex, entry.link, reverse)
}

func (entry *TableEntry) CreateSegment(
	txn txnif.AsyncTxn,
	state EntryState,
	dataFactory SegmentDataFactory,
	opts *objectio.CreateSegOpt) (created *SegmentEntry, err error) {
	entry.Lock()
	defer entry.Unlock()
	var id types.Uuid
	if opts != nil && opts.Id != nil {
		id = *opts.Id
	} else {
		id = objectio.NewSegmentid()
	}
	created = NewSegmentEntry(entry, id, txn, state, dataFactory)
	entry.AddEntryLocked(created)
	return
}

func (entry *TableEntry) MakeCommand(id uint32) (cmd txnif.TxnCmd, err error) {
	cmdType := CmdUpdateTable
	entry.RLock()
	defer entry.RUnlock()
	return newTableCmd(id, cmdType, entry), nil
}

func (entry *TableEntry) Set1PC() {
	entry.GetLatestNodeLocked().Set1PC()
}
func (entry *TableEntry) Is1PC() bool {
	return entry.GetLatestNodeLocked().Is1PC()
}
func (entry *TableEntry) AddEntryLocked(segment *SegmentEntry) {
	n := entry.link.Insert(segment)
	entry.entries[segment.ID] = n
}

func (entry *TableEntry) deleteEntryLocked(segment *SegmentEntry) error {
	if n, ok := entry.entries[segment.ID]; !ok {
		return moerr.GetOkExpectedEOB()
	} else {
		entry.link.Delete(n)
		delete(entry.entries, segment.ID)
	}
	return nil
}

func (entry *TableEntry) GetSchema() *Schema {
	return entry.schema
}

func (entry *TableEntry) GetColDefs() []*ColDef {
	colDefs := entry.schema.ColDefs
	colDefs = append(colDefs, entry.schema.PhyAddrKey)
	return colDefs
}

func (entry *TableEntry) GetFullName() string {
	if len(entry.fullName) == 0 {
		entry.fullName = genTblFullName(entry.schema.AcInfo.TenantID, entry.schema.Name)
	}
	return entry.fullName
}

func (entry *TableEntry) GetDB() *DBEntry {
	return entry.db
}

func (entry *TableEntry) PPString(level common.PPLevel, depth int, prefix string) string {
	var w bytes.Buffer
	_, _ = w.WriteString(fmt.Sprintf("%s%s%s", common.RepeatStr("\t", depth), prefix, entry.StringWithLevel(level)))
	if level == common.PPL0 {
		return w.String()
	}
	it := entry.MakeSegmentIt(true)
	for it.Valid() {
		segment := it.Get().GetPayload()
		_ = w.WriteByte('\n')
		_, _ = w.WriteString(segment.PPString(level, depth+1, prefix))
		it.Next()
	}
	return w.String()
}

func (entry *TableEntry) String() string {
	entry.RLock()
	defer entry.RUnlock()
	return entry.StringLocked()
}

func (entry *TableEntry) StringWithLevel(level common.PPLevel) string {
	entry.RLock()
	defer entry.RUnlock()
	return entry.StringLockedWithLevel(level)
}
func (entry *TableEntry) StringLockedWithLevel(level common.PPLevel) string {
	if level <= common.PPL1 {
		return fmt.Sprintf("TBL[%d][name=%s][C@%s,D@%s]",
			entry.ID, entry.schema.Name, entry.GetCreatedAt().ToString(), entry.GetDeleteAt().ToString())
	}
	return fmt.Sprintf("TBL%s[name=%s]", entry.BaseEntryImpl.StringLocked(), entry.schema.Name)
}

func (entry *TableEntry) StringLocked() string {
	return entry.StringLockedWithLevel(common.PPL1)
}

func (entry *TableEntry) GetCatalog() *Catalog { return entry.db.catalog }

func (entry *TableEntry) GetTableData() data.Table { return entry.tableData }

func (entry *TableEntry) LastAppendableSegmemt() (seg *SegmentEntry) {
	it := entry.MakeSegmentIt(false)
	for it.Valid() {
		itSeg := it.Get().GetPayload()
		dropped := itSeg.HasDropCommitted()
		if itSeg.IsAppendable() && !dropped {
			seg = itSeg
			break
		}
		it.Next()
	}
	return seg
}

func (entry *TableEntry) LastNonAppendableSegmemt() (seg *SegmentEntry) {
	it := entry.MakeSegmentIt(false)
	for it.Valid() {
		itSeg := it.Get().GetPayload()
		dropped := itSeg.HasDropCommitted()
		if !itSeg.IsAppendable() && !dropped {
			seg = itSeg
			break
		}
		it.Next()
	}
	return seg
}

func (entry *TableEntry) AsCommonID() *common.ID {
	return &common.ID{
		TableID: entry.ID,
	}
}

func (entry *TableEntry) RecurLoop(processor Processor) (err error) {
	segIt := entry.MakeSegmentIt(true)
	for segIt.Valid() {
		segment := segIt.Get().GetPayload()
		if err = processor.OnSegment(segment); err != nil {
			if moerr.IsMoErrCode(err, moerr.OkStopCurrRecur) {
				err = nil
				segIt.Next()
				continue
			}
			break
		}
		blkIt := segment.MakeBlockIt(true)
		for blkIt.Valid() {
			block := blkIt.Get().GetPayload()
			if err = processor.OnBlock(block); err != nil {
				if moerr.IsMoErrCode(err, moerr.OkStopCurrRecur) {
					err = nil
					blkIt.Next()
					continue
				}
				break
			}
			blkIt.Next()
		}
		if err = processor.OnPostSegment(segment); err != nil {
			break
		}
		segIt.Next()
	}
	if moerr.IsMoErrCode(err, moerr.OkStopCurrRecur) {
		err = nil
	}
	return err
}

func (entry *TableEntry) DropSegmentEntry(id types.Uuid, txn txnif.AsyncTxn) (deleted *SegmentEntry, err error) {
	seg, err := entry.GetSegmentByID(id)
	if err != nil {
		return
	}
	seg.Lock()
	defer seg.Unlock()
	needWait, waitTxn := seg.NeedWaitCommitting(txn.GetStartTS())
	if needWait {
		seg.Unlock()
		waitTxn.GetTxnState(true)
		seg.Lock()
	}
	var isNewNode bool
	isNewNode, err = seg.DropEntryLocked(txn)
	if err == nil && isNewNode {
		deleted = seg
	}
	return
}

func (entry *TableEntry) RemoveEntry(segment *SegmentEntry) (err error) {
	logutil.Info("[Catalog]", common.OperationField("remove"),
		common.OperandField(segment.String()))
	// segment.Close()
	entry.Lock()
	defer entry.Unlock()
	return entry.deleteEntryLocked(segment)
}

func (entry *TableEntry) PrepareRollback() (err error) {
	var isEmpty bool
	isEmpty, err = entry.BaseEntryImpl.PrepareRollback()
	if err != nil {
		return
	}
	if isEmpty {
		err = entry.GetDB().RemoveEntry(entry)
		if err != nil {
			return
		}
	}
	return
}

// IsActive is coarse API: no consistency check
func (entry *TableEntry) IsActive() bool {
	db := entry.GetDB()
	if !db.IsActive() {
		return false
	}
	return !entry.HasDropCommitted()
}

// GetTerminationTS is coarse API: no consistency check
func (entry *TableEntry) GetTerminationTS() (ts types.TS, terminated bool) {
	dbEntry := entry.GetDB()

	dbEntry.RLock()
	terminated, ts = dbEntry.TryGetTerminatedTS(true)
	dbEntry.RUnlock()

	return
}

func (entry *TableEntry) UpdateConstraint(txn txnif.TxnReader, cstr []byte) (isNewNode bool, err error) {
	entry.Lock()
	defer entry.Unlock()
	needWait, txnToWait := entry.NeedWaitCommitting(txn.GetStartTS())
	if needWait {
		entry.Unlock()
		txnToWait.GetTxnState(true)
		entry.Lock()
	}
	err = entry.CheckConflict(txn)
	if err != nil {
		return
	}
	var node *MVCCNode[*TableMVCCNode]
	isNewNode, node = entry.getOrSetUpdateNode(txn)
	node.BaseNode.Update(
		&TableMVCCNode{
			SchemaConstraints: string(cstr),
		})
	return
}

func (entry *TableEntry) CreateWithTxnAndSchema(txn txnif.AsyncTxn, schema *Schema) {
	node := &MVCCNode[*TableMVCCNode]{
		EntryMVCCNode: &EntryMVCCNode{
			CreatedAt: txnif.UncommitTS,
		},
		TxnMVCCNode: txnbase.NewTxnMVCCNodeWithTxn(txn),
		BaseNode: &TableMVCCNode{
			SchemaConstraints: string(schema.Constraint),
		},
	}
	entry.Insert(node)
}
