package main

import (
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestActionAddPackage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mock      func(*RegistryMock, *NetConnMock)
		givenArgs []string
	}{
		{
			name: "no package name",
			mock: func(reg *RegistryMock, conn *NetConnMock) {
				conn.EXPECT().Write([]byte("\nERROR: no package name\n")).Return(0, nil)
			},
			givenArgs: []string{},
		},
		{
			name: "failed adding package",
			mock: func(reg *RegistryMock, conn *NetConnMock) {
				reg.EXPECT().add("BBB", []string{"AAA"}).Return(errors.New("expected unit test error"))
				conn.EXPECT().Write([]byte("\nERROR: failed adding package: expected unit test error\n")).Return(0, nil)
			},
			givenArgs: []string{"BBB", "AAA"},
		},
		{
			name: "happy path",
			mock: func(reg *RegistryMock, conn *NetConnMock) {
				reg.EXPECT().add("BBB", []string{"AAA"}).Return(nil)
				conn.EXPECT().Write([]byte("\nPackage added\n")).Return(0, nil)
			},
			givenArgs: []string{"BBB", "AAA"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			registryMock := NewRegistryMock(ctrl)
			netConnMock := NewNetConnMock(ctrl)
			tc.mock(registryMock, netConnMock)

			logger := zap.NewNop()
			action := newAction(logger, registryMock)

			err := action.addPackage(netConnMock, tc.givenArgs...)
			require.NoError(t, err)
		})
	}
}

func TestActionRemovePackage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mock      func(*RegistryMock, *NetConnMock)
		givenArgs []string
	}{
		{
			name: "no package name",
			mock: func(reg *RegistryMock, conn *NetConnMock) {
				conn.EXPECT().Write([]byte("\nERROR: no package name\n")).Return(0, nil)
			},
			givenArgs: []string{},
		},
		{
			name: "failed removing package",
			mock: func(reg *RegistryMock, conn *NetConnMock) {
				reg.EXPECT().remove("AAA").Return(errors.New("expected unit test error"))
				conn.EXPECT().Write([]byte("\nERROR: failed removing package: expected unit test error\n")).Return(0, nil)
			},
			givenArgs: []string{"AAA"},
		},
		{
			name: "happy path",
			mock: func(reg *RegistryMock, conn *NetConnMock) {
				reg.EXPECT().remove("AAA").Return(nil)
				conn.EXPECT().Write([]byte("\nPackage removed\n")).Return(0, nil)
			},
			givenArgs: []string{"AAA"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			registryMock := NewRegistryMock(ctrl)
			netConnMock := NewNetConnMock(ctrl)
			tc.mock(registryMock, netConnMock)

			logger := zap.NewNop()
			action := newAction(logger, registryMock)

			err := action.removePackage(netConnMock, tc.givenArgs...)
			require.NoError(t, err)
		})
	}
}

func TestActionListPackages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mock func(*RegistryMock, *NetConnMock)
	}{
		{
			name: "happy path",
			mock: func(reg *RegistryMock, conn *NetConnMock) {
				reg.EXPECT().list().Return("test test test")
				conn.EXPECT().Write([]byte("\ntest test test\n")).Return(0, nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			registryMock := NewRegistryMock(ctrl)
			netConnMock := NewNetConnMock(ctrl)
			tc.mock(registryMock, netConnMock)

			logger := zap.NewNop()
			action := newAction(logger, registryMock)

			err := action.listPackages(netConnMock)
			require.NoError(t, err)
		})
	}
}
