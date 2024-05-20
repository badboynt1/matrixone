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

package process

import (
	"context"
	"sync"

	"github.com/matrixorigin/matrixone/pkg/logutil"

	"github.com/google/uuid"
)

const ALLCN = "ALLCN"
const CURRENTCN = "CURRENTCN"

type MsgType int32

const (
	MsgTopValue      MsgType = 0
	MsgPipelineStart MsgType = 1
	MsgPipelineStop  MsgType = 2
	MsgRuntimeFilter MsgType = 3
	MsgHashMap       MsgType = 4
	MaxMessage       MsgType = 1024
)

func (m MsgType) MessageName() string {
	switch m {
	case MsgTopValue:
		return "MsgTopValue"
	case MsgRuntimeFilter:
		return "MsgRuntimeFilter"
	}
	return "unknown message type"
}

type MessageAddress struct {
	CnAddr     string
	OperatorID int32
	ParallelID int32
}

type Message interface {
	Serialize() []byte
	Deserialize([]byte) Message
	NeedBlock() bool
	GetMsgTag() int32
	GetReceiverAddr() MessageAddress
}

type MessageCenter struct {
	StmtIDToBoard map[uuid.UUID]*MessageBoard
	RwMutex       *sync.Mutex
}

func (m *MessageCenter) deleteStmtID(stmtID uuid.UUID) bool {
	m.RwMutex.Lock()
	defer m.RwMutex.Unlock()
	_, ok := m.StmtIDToBoard[stmtID]
	delete(m.StmtIDToBoard, stmtID)
	return ok
}

type MessageBoard struct {
	multiCN       bool
	stmtId        uuid.UUID
	MessageCenter *MessageCenter
	Messages      []*Message
	Waiters       []chan bool
	RwMutex       *sync.RWMutex
}

func NewMessageBoard() *MessageBoard {
	m := &MessageBoard{
		Messages: make([]*Message, 0, 1024),
		Waiters:  make([]chan bool, 0, 16),
		RwMutex:  &sync.RWMutex{},
	}
	return m
}

func (m *MessageBoard) SetMultiCN(center *MessageCenter, stmtId uuid.UUID) *MessageBoard {
	logutil.Infof("set multiCN  messageBoard for stmtid %v", stmtId)
	center.RwMutex.Lock()
	defer center.RwMutex.Unlock()
	mb, ok := center.StmtIDToBoard[stmtId]
	if ok {
		logutil.Infof("get messageBoard for stmtid %v, address %v", stmtId, &mb.RwMutex)
		return mb
	}
	m.multiCN = true
	m.stmtId = stmtId
	m.MessageCenter = center
	center.StmtIDToBoard[stmtId] = m
	logutil.Infof("register messageBoard for stmtid %v, address %v", stmtId, &m.RwMutex)
	return m
}

func (m *MessageBoard) Reset() *MessageBoard {
	if m.multiCN {
		return NewMessageBoard()
	}
	m.RwMutex.Lock()
	defer m.RwMutex.Unlock()
	m.Messages = m.Messages[:0]
	m.Waiters = m.Waiters[:0]
	m.multiCN = false
	return m
}

type MessageReceiver struct {
	offset   int32
	tags     []int32
	received []int32
	addr     *MessageAddress
	mb       *MessageBoard
	waiter   chan bool
}

func (proc *Process) SendMessage(m Message) {
	mb := proc.MessageBoard
	logutil.Infof("send message %v to messageBoard address %v", m.GetMsgTag(), &mb.RwMutex)
	if m.GetReceiverAddr().CnAddr == CURRENTCN { // message for current CN
		mb.RwMutex.Lock()
		mb.Messages = append(mb.Messages, &m)
		if m.NeedBlock() {
			// broadcast for block message
			for _, ch := range mb.Waiters {
				if ch != nil && len(ch) == 0 {
					ch <- true
				}
			}
		}
		mb.RwMutex.Unlock()
	} else {
		//todo: send message to other CN, need to lookup cnlist
		panic("unsupported message yet!")
	}
}

func (proc *Process) NewMessageReceiver(tags []int32, addr MessageAddress) *MessageReceiver {
	return &MessageReceiver{
		tags: tags,
		addr: &addr,
		mb:   proc.MessageBoard,
	}
}

func (mr *MessageReceiver) receiveMessageNonBlock() []Message {
	mr.mb.RwMutex.RLock()
	defer mr.mb.RwMutex.RUnlock()
	var result []Message
	lenMessages := int32(len(mr.mb.Messages))
	for ; mr.offset < lenMessages; mr.offset++ {
		if mr.mb.Messages[mr.offset] == nil {
			continue
		}
		message := *mr.mb.Messages[mr.offset]
		if !MatchAddress(message, mr.addr) {
			continue
		}
		for i := range mr.tags {
			if mr.tags[i] == message.GetMsgTag() {
				result = append(result, message)
				mr.received = append(mr.received, mr.offset)
				break
			}
		}
	}
	return result
}

func (mr *MessageReceiver) Free() {
	if len(mr.received) == 0 {
		return
	}
	mr.mb.RwMutex.Lock()
	defer mr.mb.RwMutex.Unlock()
	for i := range mr.received {
		mr.mb.Messages[mr.received[i]] = nil
	}
	mr.received = nil
	mr.waiter = nil
}

func (mr *MessageReceiver) ReceiveMessage(needBlock bool, ctx context.Context) ([]Message, bool) {
	logutil.Infof("trying to receive message %v on messageBoard address %v", mr.tags, &mr.mb.RwMutex)
	var result = mr.receiveMessageNonBlock()
	if !needBlock || len(result) > 0 {
		logutil.Infof("receive message %v on messageBoard address %v", mr.tags, &mr.mb.RwMutex)
		return result, false
	}
	if mr.waiter == nil {
		mr.waiter = make(chan bool, 1)
		mr.mb.RwMutex.Lock()
		mr.mb.Waiters = append(mr.mb.Waiters, mr.waiter)
		mr.mb.RwMutex.Unlock()
	}
	for {
		result = mr.receiveMessageNonBlock()
		if len(result) > 0 {
			break
		}
		logutil.Infof("receive message %v on messageBoard address %v", mr.tags, &mr.mb.RwMutex)
		select {
		case <-mr.waiter:
		case <-ctx.Done():
			return result, true
		}
	}
	logutil.Infof("receive message %v on messageBoard address %v", mr.tags, &mr.mb.RwMutex)
	return result, false
}

func MatchAddress(m Message, raddr *MessageAddress) bool {
	mAddr := m.GetReceiverAddr()
	if mAddr.OperatorID != raddr.OperatorID && mAddr.OperatorID != -1 {
		return false
	}
	if mAddr.ParallelID != raddr.ParallelID && mAddr.ParallelID != -1 {
		return false
	}
	return true
}

func AddrBroadCastOnCurrentCN() MessageAddress {
	return MessageAddress{
		CnAddr:     CURRENTCN,
		OperatorID: -1,
		ParallelID: -1,
	}
}

func AddrBroadCastOnALLCN() MessageAddress {
	return MessageAddress{
		CnAddr:     ALLCN,
		OperatorID: -1,
		ParallelID: -1,
	}
}
