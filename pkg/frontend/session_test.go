// Copyright 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package frontend

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/fagongzi/goetty/v2/buf"
	"github.com/golang/mock/gomock"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/config"
	"github.com/matrixorigin/matrixone/pkg/defines"
	mock_frontend "github.com/matrixorigin/matrixone/pkg/frontend/test"
	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	plan2 "github.com/matrixorigin/matrixone/pkg/sql/plan"
	"github.com/matrixorigin/matrixone/pkg/txn/client"
	"github.com/matrixorigin/matrixone/pkg/txn/clock"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestTxnHandler_NewTxn(t *testing.T) {
	convey.Convey("new txn", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.TODO()
		txnOperator := mock_frontend.NewMockTxnOperator(ctrl)
		txnOperator.EXPECT().Txn().Return(txn.TxnMeta{}).AnyTimes()
		txnOperator.EXPECT().Rollback(gomock.Any()).Return(nil).AnyTimes()
		txnOperator.EXPECT().Commit(gomock.Any()).Return(nil).AnyTimes()
		txnClient := mock_frontend.NewMockTxnClient(ctrl)
		cnt := 0
		txnClient.EXPECT().New().DoAndReturn(
			func(ootions ...client.TxnOption) (client.TxnOperator, error) {
				cnt++
				if cnt%2 != 0 {
					return txnOperator, nil
				} else {
					return nil, moerr.NewInternalError(ctx, "startTxn failed")
				}
			}).AnyTimes()
		eng := mock_frontend.NewMockEngine(ctrl)
		eng.EXPECT().New(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Commit(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Rollback(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().New(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Commit(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Rollback(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Hints().Return(engine.Hints{
			CommitOrRollbackTimeout: time.Second,
		}).AnyTimes()

		txn := InitTxnHandler(eng, txnClient)
		txn.ses = &Session{
			requestCtx: ctx,
		}
		err := txn.NewTxn()
		convey.So(err, convey.ShouldBeNil)
		err = txn.NewTxn()
		convey.So(err, convey.ShouldNotBeNil)
		err = txn.NewTxn()
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestTxnHandler_CommitTxn(t *testing.T) {
	convey.Convey("commit txn", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.TODO()
		txnOperator := mock_frontend.NewMockTxnOperator(ctrl)
		txnOperator.EXPECT().Txn().Return(txn.TxnMeta{}).AnyTimes()
		cnt := 0
		txnOperator.EXPECT().Commit(gomock.Any()).DoAndReturn(
			func(ctx context.Context) error {
				cnt++
				if cnt%2 != 0 {
					return nil
				} else {
					return moerr.NewInternalError(ctx, "commit failed")
				}
			}).AnyTimes()

		txnClient := mock_frontend.NewMockTxnClient(ctrl)
		eng := mock_frontend.NewMockEngine(ctrl)
		eng.EXPECT().New(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Commit(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Rollback(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Hints().Return(engine.Hints{
			CommitOrRollbackTimeout: time.Second,
		}).AnyTimes()

		txnClient.EXPECT().New().Return(txnOperator, nil).AnyTimes()

		txn := InitTxnHandler(eng, txnClient)
		txn.ses = &Session{
			requestCtx: ctx,
		}
		err := txn.NewTxn()
		convey.So(err, convey.ShouldBeNil)
		err = txn.CommitTxn()
		convey.So(err, convey.ShouldBeNil)
		err = txn.NewTxn()
		convey.So(err, convey.ShouldBeNil)
		err = txn.CommitTxn()
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestTxnHandler_RollbackTxn(t *testing.T) {
	convey.Convey("rollback txn", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.TODO()
		txnOperator := mock_frontend.NewMockTxnOperator(ctrl)
		txnOperator.EXPECT().Txn().Return(txn.TxnMeta{}).AnyTimes()
		cnt := 0
		txnOperator.EXPECT().Rollback(gomock.Any()).DoAndReturn(
			func(ctc context.Context) error {
				cnt++
				if cnt%2 != 0 {
					return nil
				} else {
					return moerr.NewInternalError(ctx, "rollback failed")
				}
			}).AnyTimes()

		txnClient := mock_frontend.NewMockTxnClient(ctrl)
		eng := mock_frontend.NewMockEngine(ctrl)
		eng.EXPECT().New(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Commit(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Rollback(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Hints().Return(engine.Hints{
			CommitOrRollbackTimeout: time.Second,
		}).AnyTimes()

		txnClient.EXPECT().New().Return(txnOperator, nil).AnyTimes()

		txn := InitTxnHandler(eng, txnClient)
		txn.ses = &Session{
			requestCtx: ctx,
		}
		err := txn.NewTxn()
		convey.So(err, convey.ShouldBeNil)
		err = txn.RollbackTxn()
		convey.So(err, convey.ShouldBeNil)
		err = txn.NewTxn()
		convey.So(err, convey.ShouldBeNil)
		err = txn.RollbackTxn()
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestSession_TxnBegin(t *testing.T) {
	genSession := func(ctrl *gomock.Controller, gSysVars *GlobalSystemVariables) *Session {
		ioses := mock_frontend.NewMockIOSession(ctrl)
		ioses.EXPECT().OutBuf().Return(buf.NewByteBuf(1024)).AnyTimes()
		ioses.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		ioses.EXPECT().RemoteAddress().Return("").AnyTimes()
		ioses.EXPECT().Ref().AnyTimes()
		sv, err := getSystemVariables("test/system_vars_config.toml")
		if err != nil {
			t.Error(err)
		}
		proto := NewMysqlClientProtocol(0, ioses, 1024, sv)
		txnOperator := mock_frontend.NewMockTxnOperator(ctrl)
		txnOperator.EXPECT().Txn().Return(txn.TxnMeta{}).AnyTimes()
		txnOperator.EXPECT().Commit(gomock.Any()).Return(nil).AnyTimes()
		txnClient := mock_frontend.NewMockTxnClient(ctrl)
		txnClient.EXPECT().New().Return(txnOperator, nil).AnyTimes()
		eng := mock_frontend.NewMockEngine(ctrl)
		hints := engine.Hints{CommitOrRollbackTimeout: time.Second * 10}
		eng.EXPECT().Hints().Return(hints).AnyTimes()
		eng.EXPECT().New(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Commit(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		session := NewSession(proto, nil, config.NewParameterUnit(&config.FrontendParameters{}, eng, txnClient, nil), gSysVars, false)
		session.SetRequestContext(context.Background())
		return session
	}
	convey.Convey("new session", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		gSysVars := &GlobalSystemVariables{}
		InitGlobalSystemVariables(gSysVars)

		ses := genSession(ctrl, gSysVars)
		err := ses.TxnBegin()
		convey.So(err, convey.ShouldBeNil)
		err = ses.TxnCommit()
		convey.So(err, convey.ShouldBeNil)
		err = ses.TxnBegin()
		convey.So(err, convey.ShouldBeNil)
		err = ses.SetAutocommit(false)
		convey.So(err, convey.ShouldNotBeNil)
		err = ses.TxnCommit()
		convey.So(err, convey.ShouldBeNil)
		_, _ = ses.GetTxnHandler().GetTxn()
		convey.So(err, convey.ShouldBeNil)

		err = ses.TxnCommit()
		convey.So(err, convey.ShouldBeNil)

		err = ses.SetAutocommit(true)
		convey.So(err, convey.ShouldBeNil)

		err = ses.SetAutocommit(false)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestVariables(t *testing.T) {
	genSession := func(ctrl *gomock.Controller, gSysVars *GlobalSystemVariables) *Session {
		ioses := mock_frontend.NewMockIOSession(ctrl)
		ioses.EXPECT().OutBuf().Return(buf.NewByteBuf(1024)).AnyTimes()
		ioses.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		ioses.EXPECT().RemoteAddress().Return("").AnyTimes()
		ioses.EXPECT().Ref().AnyTimes()
		sv, err := getSystemVariables("test/system_vars_config.toml")
		if err != nil {
			t.Error(err)
		}
		proto := NewMysqlClientProtocol(0, ioses, 1024, sv)
		txnClient := mock_frontend.NewMockTxnClient(ctrl)
		txnClient.EXPECT().New().AnyTimes()
		session := NewSession(proto, nil, config.NewParameterUnit(&config.FrontendParameters{}, nil, txnClient, nil), gSysVars, true)
		session.SetRequestContext(context.Background())
		return session
	}

	checkWant := func(ses, existSes, newSesAfterSession *Session,
		v string,
		sameSesWant1, existSesWant2, newSesAfterSesWant3,
		saneSesGlobalWant4, existSesGlobalWant5, newSesAfterSesGlobalWant6 interface{}) {

		//same session
		v1_val, err := ses.GetSessionVar(v)
		convey.So(err, convey.ShouldBeNil)
		convey.So(sameSesWant1, convey.ShouldEqual, v1_val)
		v1_ctx_val, err := ses.GetTxnCompileCtx().ResolveVariable(v, true, false)
		convey.So(err, convey.ShouldBeNil)
		convey.So(v1_ctx_val, convey.ShouldEqual, v1_val)

		//exist session
		v2_val, err := existSes.GetSessionVar(v)
		convey.So(err, convey.ShouldBeNil)
		convey.So(existSesWant2, convey.ShouldEqual, v2_val)
		v2_ctx_val, err := existSes.GetTxnCompileCtx().ResolveVariable(v, true, false)
		convey.So(err, convey.ShouldBeNil)
		convey.So(v2_ctx_val, convey.ShouldEqual, v2_val)

		//new session after session
		v3_val, err := newSesAfterSession.GetSessionVar(v)
		convey.So(err, convey.ShouldBeNil)
		convey.So(newSesAfterSesWant3, convey.ShouldEqual, v3_val)
		v3_ctx_val, err := newSesAfterSession.GetTxnCompileCtx().ResolveVariable(v, true, false)
		convey.So(err, convey.ShouldBeNil)
		convey.So(v3_ctx_val, convey.ShouldEqual, v3_val)

		//same session global
		v4_val, err := ses.GetGlobalVar(v)
		convey.So(err, convey.ShouldBeNil)
		convey.So(saneSesGlobalWant4, convey.ShouldEqual, v4_val)
		v4_ctx_val, err := ses.GetTxnCompileCtx().ResolveVariable(v, true, true)
		convey.So(err, convey.ShouldBeNil)
		convey.So(v4_ctx_val, convey.ShouldEqual, v4_val)

		//exist session global
		v5_val, err := existSes.GetGlobalVar(v)
		convey.So(err, convey.ShouldBeNil)
		convey.So(existSesGlobalWant5, convey.ShouldEqual, v5_val)
		v5_ctx_val, err := existSes.GetTxnCompileCtx().ResolveVariable(v, true, true)
		convey.So(err, convey.ShouldBeNil)
		convey.So(v5_ctx_val, convey.ShouldEqual, v5_val)

		//new session after session global
		v6_val, err := newSesAfterSession.GetGlobalVar(v)
		convey.So(err, convey.ShouldBeNil)
		convey.So(newSesAfterSesGlobalWant6, convey.ShouldEqual, v6_val)
		v6_ctx_val, err := newSesAfterSession.GetTxnCompileCtx().ResolveVariable(v, true, true)
		convey.So(err, convey.ShouldBeNil)
		convey.So(v6_ctx_val, convey.ShouldEqual, v6_val)
	}

	checkWant2 := func(ses, existSes, newSesAfterSession *Session,
		v string,
		sameSesWant1, existSesWant2, newSesAfterSesWant3 interface{}) {

		//same session
		v1_val, err := ses.GetSessionVar(v)
		convey.So(err, convey.ShouldBeNil)
		convey.So(sameSesWant1, convey.ShouldEqual, v1_val)
		v1_ctx_val, err := ses.GetTxnCompileCtx().ResolveVariable(v, true, false)
		convey.So(err, convey.ShouldBeNil)
		convey.So(v1_ctx_val, convey.ShouldEqual, v1_val)

		//exist session
		v2_val, err := existSes.GetSessionVar(v)
		convey.So(err, convey.ShouldBeNil)
		convey.So(existSesWant2, convey.ShouldEqual, v2_val)
		v2_ctx_val, err := existSes.GetTxnCompileCtx().ResolveVariable(v, true, false)
		convey.So(err, convey.ShouldBeNil)
		convey.So(v2_ctx_val, convey.ShouldEqual, v2_val)

		//new session after session
		v3_val, err := newSesAfterSession.GetSessionVar(v)
		convey.So(err, convey.ShouldBeNil)
		convey.So(newSesAfterSesWant3, convey.ShouldEqual, v3_val)
		v3_ctx_val, err := newSesAfterSession.GetTxnCompileCtx().ResolveVariable(v, true, false)
		convey.So(err, convey.ShouldBeNil)
		convey.So(v3_ctx_val, convey.ShouldEqual, v3_val)

		//same session global
		_, err = ses.GetGlobalVar(v)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeError, moerr.NewInternalError(context.TODO(), errorSystemVariableSessionEmpty()))
		_, err = ses.GetTxnCompileCtx().ResolveVariable(v, true, true)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeError, moerr.NewInternalError(context.TODO(), errorSystemVariableSessionEmpty()))

		//exist session global
		_, err = existSes.GetGlobalVar(v)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeError, moerr.NewInternalError(context.TODO(), errorSystemVariableSessionEmpty()))
		_, err = existSes.GetTxnCompileCtx().ResolveVariable(v, true, true)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeError, moerr.NewInternalError(context.TODO(), errorSystemVariableSessionEmpty()))

		//new session after session global
		_, err = newSesAfterSession.GetGlobalVar(v)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeError, moerr.NewInternalError(context.TODO(), errorSystemVariableSessionEmpty()))
		_, err = newSesAfterSession.GetTxnCompileCtx().ResolveVariable(v, true, true)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeError, moerr.NewInternalError(context.TODO(), errorSystemVariableSessionEmpty()))
	}

	convey.Convey("scope global", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		gSysVars := &GlobalSystemVariables{}
		InitGlobalSystemVariables(gSysVars)

		ses := genSession(ctrl, gSysVars)
		existSes := genSession(ctrl, gSysVars)

		v1 := "testglobalvar_dyn"
		_, v1_default, _ := gSysVars.GetGlobalSysVar(v1)
		v1_want := 10
		err := ses.SetSessionVar(v1, v1_want)
		convey.So(err, convey.ShouldNotBeNil)

		// no check after fail set
		newSes2 := genSession(ctrl, gSysVars)
		checkWant(ses, existSes, newSes2, v1, v1_default, v1_default, v1_default, v1_default, v1_default, v1_default)

		err = ses.SetGlobalVar(v1, v1_want)
		convey.So(err, convey.ShouldBeNil)

		newSes3 := genSession(ctrl, gSysVars)
		checkWant(ses, existSes, newSes3, v1, v1_want, v1_want, v1_want, v1_want, v1_want, v1_want)

		v2 := "testglobalvar_nodyn"
		_, v2_default, _ := gSysVars.GetGlobalSysVar(v2)
		v2_want := 10
		err = ses.SetSessionVar(v2, v2_want)
		convey.So(err, convey.ShouldNotBeNil)

		newSes4 := genSession(ctrl, gSysVars)
		checkWant(ses, existSes, newSes4, v2, v2_default, v2_default, v2_default, v2_default, v2_default, v2_default)

		err = ses.SetGlobalVar(v2, v2_want)
		convey.So(err, convey.ShouldNotBeNil)

		newSes5 := genSession(ctrl, gSysVars)
		checkWant(ses, existSes, newSes5, v2, v2_default, v2_default, v2_default, v2_default, v2_default, v2_default)
	})

	convey.Convey("scope session", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		gSysVars := &GlobalSystemVariables{}
		InitGlobalSystemVariables(gSysVars)

		ses := genSession(ctrl, gSysVars)
		existSes := genSession(ctrl, gSysVars)

		v1 := "testsessionvar_dyn"
		_, v1_default, _ := gSysVars.GetGlobalSysVar(v1)
		v1_want := 10
		err := ses.SetSessionVar(v1, v1_want)
		convey.So(err, convey.ShouldBeNil)

		newSes1 := genSession(ctrl, gSysVars)
		checkWant2(ses, existSes, newSes1, v1, v1_want, v1_default, v1_default)

		err = ses.SetGlobalVar(v1, v1_want)
		convey.So(err, convey.ShouldNotBeNil)

		newSes2 := genSession(ctrl, gSysVars)
		checkWant2(ses, existSes, newSes2, v1, v1_want, v1_default, v1_default)

		v2 := "testsessionvar_nodyn"
		_, v2_default, _ := gSysVars.GetGlobalSysVar(v2)
		v2_want := 10
		err = ses.SetSessionVar(v2, v2_want)
		convey.So(err, convey.ShouldNotBeNil)

		newSes3 := genSession(ctrl, gSysVars)
		checkWant2(ses, existSes, newSes3, v2, v2_default, v2_default, v2_default)

		err = ses.SetGlobalVar(v2, v2_want)
		convey.So(err, convey.ShouldNotBeNil)
		newSes4 := genSession(ctrl, gSysVars)
		checkWant2(ses, existSes, newSes4, v2, v2_default, v2_default, v2_default)

	})

	convey.Convey("scope both - set session", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		gSysVars := &GlobalSystemVariables{}
		InitGlobalSystemVariables(gSysVars)

		ses := genSession(ctrl, gSysVars)
		existSes := genSession(ctrl, gSysVars)

		v1 := "testbothvar_dyn"
		_, v1_default, _ := gSysVars.GetGlobalSysVar(v1)
		v1_want := 10
		err := ses.SetSessionVar(v1, v1_want)
		convey.So(err, convey.ShouldBeNil)

		newSes2 := genSession(ctrl, gSysVars)
		checkWant(ses, existSes, newSes2, v1, v1_want, v1_default, v1_default, v1_default, v1_default, v1_default)

		v2 := "testbotchvar_nodyn"
		err = ses.SetSessionVar(v2, 10)
		convey.So(err, convey.ShouldNotBeNil)

		err = ses.SetGlobalVar(v2, 10)
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("scope both", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		gSysVars := &GlobalSystemVariables{}
		InitGlobalSystemVariables(gSysVars)

		ses := genSession(ctrl, gSysVars)
		existSes := genSession(ctrl, gSysVars)

		v1 := "testbothvar_dyn"
		_, v1_default, _ := gSysVars.GetGlobalSysVar(v1)
		v1_want := 10

		err := ses.SetGlobalVar(v1, v1_want)
		convey.So(err, convey.ShouldBeNil)

		newSes2 := genSession(ctrl, gSysVars)
		checkWant(ses, existSes, newSes2, v1, v1_default, v1_default, v1_want, v1_want, v1_want, v1_want)
	})

	convey.Convey("user variables", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		gSysVars := &GlobalSystemVariables{}
		InitGlobalSystemVariables(gSysVars)

		ses := genSession(ctrl, gSysVars)

		vars := ses.CopyAllSessionVars()
		convey.So(len(vars), convey.ShouldNotBeZeroValue)

		err := ses.SetUserDefinedVar("abc", 1)
		convey.So(err, convey.ShouldBeNil)

		_, _, err = ses.GetUserDefinedVar("abc")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestSession_TxnCompilerContext(t *testing.T) {
	genSession := func(ctrl *gomock.Controller, pu *config.ParameterUnit, gSysVars *GlobalSystemVariables) *Session {
		ioses := mock_frontend.NewMockIOSession(ctrl)
		ioses.EXPECT().OutBuf().Return(buf.NewByteBuf(1024)).AnyTimes()
		ioses.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		ioses.EXPECT().RemoteAddress().Return("").AnyTimes()
		ioses.EXPECT().Ref().AnyTimes()
		sv, err := getSystemVariables("test/system_vars_config.toml")
		if err != nil {
			t.Error(err)
		}
		proto := NewMysqlClientProtocol(0, ioses, 1024, sv)
		session := NewSession(proto, nil, pu, gSysVars, false)
		session.SetRequestContext(context.Background())
		return session
	}

	convey.Convey("test", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.TODO()
		txnOperator := mock_frontend.NewMockTxnOperator(ctrl)
		txnOperator.EXPECT().Commit(ctx).Return(nil).AnyTimes()
		txnOperator.EXPECT().Rollback(ctx).Return(nil).AnyTimes()
		txnClient := mock_frontend.NewMockTxnClient(ctrl)
		txnClient.EXPECT().New().Return(txnOperator, nil).AnyTimes()
		eng := mock_frontend.NewMockEngine(ctrl)
		eng.EXPECT().New(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Commit(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Rollback(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		eng.EXPECT().Hints().Return(engine.Hints{
			CommitOrRollbackTimeout: time.Second,
		}).AnyTimes()

		db := mock_frontend.NewMockDatabase(ctrl)
		db.EXPECT().Relations(gomock.Any()).Return(nil, nil).AnyTimes()

		table := mock_frontend.NewMockRelation(ctrl)
		table.EXPECT().Ranges(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		table.EXPECT().TableDefs(gomock.Any()).Return(nil, nil).AnyTimes()
		table.EXPECT().GetPrimaryKeys(gomock.Any()).Return(nil, nil).AnyTimes()
		table.EXPECT().GetHideKeys(gomock.Any()).Return(nil, nil).AnyTimes()
		table.EXPECT().Stats(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		table.EXPECT().TableColumns(gomock.Any()).Return(nil, nil).AnyTimes()
		table.EXPECT().GetTableID(gomock.Any()).Return(uint64(10)).AnyTimes()
		db.EXPECT().Relation(gomock.Any(), gomock.Any()).Return(table, nil).AnyTimes()
		eng.EXPECT().Database(gomock.Any(), gomock.Any(), gomock.Any()).Return(db, nil).AnyTimes()

		pu := config.NewParameterUnit(&config.FrontendParameters{}, eng, txnClient, nil)

		gSysVars := &GlobalSystemVariables{}
		InitGlobalSystemVariables(gSysVars)

		ses := genSession(ctrl, pu, gSysVars)

		tcc := ses.GetTxnCompileCtx()
		defDBName := tcc.DefaultDatabase()
		convey.So(defDBName, convey.ShouldEqual, "")
		convey.So(tcc.DatabaseExists("abc"), convey.ShouldBeTrue)

		_, _, err := tcc.getRelation("abc", "t1")
		convey.So(err, convey.ShouldBeNil)

		object, tableRef := tcc.Resolve("abc", "t1")
		convey.So(object, convey.ShouldNotBeNil)
		convey.So(tableRef, convey.ShouldNotBeNil)

		pkd := tcc.GetPrimaryKeyDef("abc", "t1")
		convey.So(len(pkd), convey.ShouldBeZeroValue)

		hkd := tcc.GetHideKeyDef("abc", "t1")
		convey.So(hkd, convey.ShouldBeNil)

		stats := tcc.Stats(&plan2.ObjectRef{SchemaName: "abc", ObjName: "t1"}, &plan2.Expr{})
		convey.So(stats, convey.ShouldBeNil)
	})
}

func TestSession_GetTempTableStorage(t *testing.T) {
	genSession := func(ctrl *gomock.Controller, pu *config.ParameterUnit, gSysVars *GlobalSystemVariables) *Session {
		ioses := mock_frontend.NewMockIOSession(ctrl)
		ioses.EXPECT().OutBuf().Return(buf.NewByteBuf(1024)).AnyTimes()
		ioses.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		ioses.EXPECT().RemoteAddress().Return("").AnyTimes()
		ioses.EXPECT().Ref().AnyTimes()
		sv, err := getSystemVariables("test/system_vars_config.toml")
		if err != nil {
			t.Error(err)
		}
		proto := NewMysqlClientProtocol(0, ioses, 1024, sv)
		session := NewSession(proto, nil, pu, gSysVars, false)
		session.SetRequestContext(context.Background())
		return session
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	txnClient := mock_frontend.NewMockTxnClient(ctrl)
	eng := mock_frontend.NewMockEngine(ctrl)
	pu := config.NewParameterUnit(&config.FrontendParameters{}, eng, txnClient, nil)
	gSysVars := &GlobalSystemVariables{}

	ses := genSession(ctrl, pu, gSysVars)
	assert.Panics(t, func() {
		_ = ses.GetTempTableStorage()
	})
}

func TestIfInitedTempEngine(t *testing.T) {
	genSession := func(ctrl *gomock.Controller, pu *config.ParameterUnit, gSysVars *GlobalSystemVariables) *Session {
		ioses := mock_frontend.NewMockIOSession(ctrl)
		ioses.EXPECT().OutBuf().Return(buf.NewByteBuf(1024)).AnyTimes()
		ioses.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		ioses.EXPECT().RemoteAddress().Return("").AnyTimes()
		ioses.EXPECT().Ref().AnyTimes()
		sv, err := getSystemVariables("test/system_vars_config.toml")
		if err != nil {
			t.Error(err)
		}
		proto := NewMysqlClientProtocol(0, ioses, 1024, sv)
		session := NewSession(proto, nil, pu, gSysVars, false)
		session.SetRequestContext(context.Background())
		return session
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	txnClient := mock_frontend.NewMockTxnClient(ctrl)
	eng := mock_frontend.NewMockEngine(ctrl)
	pu := config.NewParameterUnit(&config.FrontendParameters{}, eng, txnClient, nil)
	gSysVars := &GlobalSystemVariables{}

	ses := genSession(ctrl, pu, gSysVars)
	assert.False(t, ses.IfInitedTempEngine())
}

func TestSetTempTableStorage(t *testing.T) {
	genSession := func(ctrl *gomock.Controller, pu *config.ParameterUnit, gSysVars *GlobalSystemVariables) *Session {
		ioses := mock_frontend.NewMockIOSession(ctrl)
		ioses.EXPECT().OutBuf().Return(buf.NewByteBuf(1024)).AnyTimes()
		ioses.EXPECT().Write(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		ioses.EXPECT().RemoteAddress().Return("").AnyTimes()
		ioses.EXPECT().Ref().AnyTimes()
		sv, err := getSystemVariables("test/system_vars_config.toml")
		if err != nil {
			t.Error(err)
		}
		proto := NewMysqlClientProtocol(0, ioses, 1024, sv)
		session := NewSession(proto, nil, pu, gSysVars, false)
		session.SetRequestContext(context.Background())
		return session
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	txnClient := mock_frontend.NewMockTxnClient(ctrl)
	eng := mock_frontend.NewMockEngine(ctrl)
	pu := config.NewParameterUnit(&config.FrontendParameters{}, eng, txnClient, nil)
	gSysVars := &GlobalSystemVariables{}

	ses := genSession(ctrl, pu, gSysVars)

	ck := clock.NewHLCClock(func() int64 {
		return time.Now().Unix()
	}, math.MaxInt)
	dnStore, _ := ses.SetTempTableStorage(ck)

	assert.Equal(t, defines.TEMPORARY_TABLE_DN_ADDR, dnStore.TxnServiceAddress)
}
