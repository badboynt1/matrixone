// Copyright The OpenTelemetry Authors
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

// Portions of this file are additionally subject to the following
// copyright.
//
// Copyright (C) 2022 Matrix Origin.
//
// Modified the behavior and the interface of the step.

package trace

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_initExport(t *testing.T) {
	type args struct {
		enableTracer bool
		config       *tracerProviderConfig
		needRecover  bool
		shutdownCtx  context.Context
	}
	cancledCtx, cancle := context.WithCancel(context.Background())
	cancle()
	ch := make(chan string, 10)
	tests := []struct {
		name  string
		args  args
		empty bool
	}{
		{
			name: "disable",
			args: args{
				enableTracer: false,
				config:       &tracerProviderConfig{enable: false},
				shutdownCtx:  nil,
			},
			empty: true,
		},
		{
			name: "enable_InternalExecutor",
			args: args{
				enableTracer: true,
				config: &tracerProviderConfig{
					enable: true, batchProcessMode: InternalExecutor, sqlExecutor: newDummyExecutorFactory(ch),
					batchProcessor: noopBatchProcessor{},
				},
				needRecover: true,
				shutdownCtx: context.Background(),
			},
			empty: false,
		},
		{
			name: "enable_FileService",
			args: args{
				enableTracer: true,
				config: &tracerProviderConfig{
					enable: true, batchProcessMode: FileService, sqlExecutor: newDummyExecutorFactory(ch),
					batchProcessor: noopBatchProcessor{},
				},
				shutdownCtx: context.Background(),
			},
			empty: false,
		},
		{
			name: "enable_FileService_with_canceled_ctx",
			args: args{
				enableTracer: true,
				config: &tracerProviderConfig{
					enable: true, batchProcessMode: FileService, sqlExecutor: newDummyExecutorFactory(ch),
					batchProcessor: noopBatchProcessor{},
				},
				shutdownCtx: cancledCtx,
			},
			empty: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.needRecover {
				defer func() {
					if err := recover(); err != nil {
						t.Logf("pass with panic.")
						return
					}
					panic("not catch panic")
				}()
			}
			initExporter(context.TODO(), tt.args.config)
			require.Equal(t, "trace.noopBatchProcessor", fmt.Sprintf("%v", reflect.ValueOf(GetGlobalBatchProcessor()).Type()))
			require.Equal(t, Shutdown(tt.args.shutdownCtx), nil)
		})
	}
}

func TestDefaultContext(t *testing.T) {
	var spanId SpanID
	spanId.SetByUUID(GetNodeResource().NodeUuid)
	tests := []struct {
		name string
		want context.Context
	}{
		{
			name: "normal",
			want: ContextWithSpanContext(context.Background(), SpanContextWithIDs(nilTraceID, spanId)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, DefaultContext(), "DefaultContext()")
		})
	}
}

func TestDefaultSpanContext(t *testing.T) {
	var spanId SpanID
	spanId.SetByUUID(GetNodeResource().NodeUuid)
	sc := SpanContextWithIDs(nilTraceID, spanId)
	tests := []struct {
		name string
		want *SpanContext
	}{
		{
			name: "normal",
			want: &sc,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, DefaultSpanContext(), "DefaultSpanContext()")
		})
	}
}

func TestGetNodeResource(t *testing.T) {
	tests := []struct {
		name string
		want *MONodeResource
	}{
		{
			name: "normal",
			want: &MONodeResource{"node_uuid", NodeTypeStandalone},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetNodeResource(), "GetNodeResource()")
		})
	}
}
