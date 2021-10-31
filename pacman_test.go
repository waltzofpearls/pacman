package main

import (
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPacmanListen(t *testing.T) {
	t.Parallel()

	rootCA, err := ioutil.ReadFile("testdata/Test_Root_CA.crt")
	require.NoError(t, err)
	serverKey, err := ioutil.ReadFile("testdata/unit_test.key")
	require.NoError(t, err)
	serverCert, err := ioutil.ReadFile("testdata/unit_test.crt")
	require.NoError(t, err)

	tests := []struct {
		name        string
		givenConfig *config
		wantError   error
	}{
		{
			name: "failed loading TLS config",
			givenConfig: &config{
				UseMTLS: true,
				RootCA:  "not_a_root_ca",
			},
			wantError: errors.New("cannot append root CA cert"),
		},
		{
			name: "listen with mTLS",
			givenConfig: &config{
				UseMTLS:    true,
				RootCA:     string(rootCA),
				ServerKey:  string(serverKey),
				ServerCert: string(serverCert),
			},
			wantError: nil,
		},
		{
			name: "listen without mTLS",
			givenConfig: &config{
				UseMTLS: false,
			},
			wantError: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := pacman{
				logger: zap.NewNop(),
				config: tc.givenConfig,
			}
			listener, err := p.listen()

			if tc.wantError != nil {
				assert.EqualError(t, err, tc.wantError.Error())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, listener)
			}
		})
	}
}

type netTempError struct{}

func (e *netTempError) Error() string   { return "unit test net temp error" }
func (e *netTempError) Timeout() bool   { return true }
func (e *netTempError) Temporary() bool { return true }

func TestPacmanServe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mock func(chan os.Signal, *NetConnMock, *NetListenerMock)
	}{
		{
			name: "handle listner accept errors",
			mock: func(sig chan os.Signal, conn *NetConnMock, lis *NetListenerMock) {
				gomock.InOrder(
					lis.EXPECT().Accept().Return(nil, new(netTempError)),
					lis.EXPECT().Accept().Return(nil, new(netTempError)),
					lis.EXPECT().Accept().Return(nil, errors.New("expected unit test error")),
					lis.EXPECT().Close().Return(nil),
				)
			},
		},
		{
			name: "handle listner accept errors",
			mock: func(sig chan os.Signal, conn *NetConnMock, lis *NetListenerMock) {
				conn.EXPECT().RemoteAddr().Return(new(net.TCPAddr)).AnyTimes()
				conn.EXPECT().SetReadDeadline(gomock.Any()).AnyTimes()
				conn.EXPECT().Read(gomock.Any()).Return(0, io.EOF).AnyTimes()
				conn.EXPECT().Write([]byte("\nERROR: unknown action\n")).Return(0, nil).AnyTimes()
				conn.EXPECT().Close().Return(nil).AnyTimes()
				lis.EXPECT().Accept().Return(conn, nil).AnyTimes()
				lis.EXPECT().Close().Return(nil)
				sig <- os.Interrupt
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			shutdown := make(chan os.Signal, 1)
			netConnMock := NewNetConnMock(ctrl)
			netListenerMock := NewNetListenerMock(ctrl)
			tc.mock(shutdown, netConnMock, netListenerMock)

			p := pacman{
				logger:   zap.NewNop(),
				config:   &config{},
				shutdown: shutdown,
			}
			p.serve(netListenerMock)
		})
	}
}

func TestPacmanHandle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mock func(*NetConnMock, *HandlerMock)
	}{
		{
			name: "add package",
			mock: func(conn *NetConnMock, hdl *HandlerMock) {
				conn.EXPECT().RemoteAddr().Return(new(net.TCPAddr))
				conn.EXPECT().SetReadDeadline(gomock.Any()).Times(2)
				conn.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (n int, err error) {
					data := []byte("AddPackage CCC AAA BBB")
					n = copy(p, data[:])
					return n, io.EOF
				})
				conn.EXPECT().Close().Return(nil)
				hdl.EXPECT().addPackage(conn, []string{"CCC", "AAA", "BBB"}).Return(nil)
			},
		},
		{
			name: "remove package",
			mock: func(conn *NetConnMock, hdl *HandlerMock) {
				conn.EXPECT().RemoteAddr().Return(new(net.TCPAddr))
				conn.EXPECT().SetReadDeadline(gomock.Any()).Times(2)
				conn.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (n int, err error) {
					data := []byte("RemovePackage CCC")
					n = copy(p, data[:])
					return n, io.EOF
				})
				conn.EXPECT().Close().Return(nil)
				hdl.EXPECT().removePackage(conn, []string{"CCC"}).Return(nil)
			},
		},
		{
			name: "list package",
			mock: func(conn *NetConnMock, hdl *HandlerMock) {
				conn.EXPECT().RemoteAddr().Return(new(net.TCPAddr))
				conn.EXPECT().SetReadDeadline(gomock.Any()).Times(2)
				conn.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (n int, err error) {
					data := []byte("ListPackages")
					n = copy(p, data[:])
					return n, io.EOF
				})
				conn.EXPECT().Close().Return(nil)
				hdl.EXPECT().listPackages(conn).Return(nil)
			},
		},
		{
			name: "unknown action",
			mock: func(conn *NetConnMock, hdl *HandlerMock) {
				conn.EXPECT().RemoteAddr().Return(new(net.TCPAddr))
				conn.EXPECT().SetReadDeadline(gomock.Any()).Times(2)
				conn.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (n int, err error) {
					data := []byte("UnkownAction")
					n = copy(p, data[:])
					return n, io.EOF
				})
				conn.EXPECT().Write([]byte("\nERROR: unknown action\n")).Return(0, nil)
				conn.EXPECT().Close().Return(nil)
			},
		},
		{
			name: "unknown action and error writing to connection",
			mock: func(conn *NetConnMock, hdl *HandlerMock) {
				conn.EXPECT().RemoteAddr().Return(new(net.TCPAddr))
				conn.EXPECT().SetReadDeadline(gomock.Any())
				conn.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (n int, err error) {
					data := []byte("UnkownAction")
					n = copy(p, data[:])
					return n, io.EOF
				})
				conn.EXPECT().Write([]byte("\nERROR: unknown action\n")).Return(0, errors.New("expected unit test error"))
				conn.EXPECT().Close().Return(nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			netConnMock := NewNetConnMock(ctrl)
			handlerMock := NewHandlerMock(ctrl)
			tc.mock(netConnMock, handlerMock)

			p := pacman{
				logger:  zap.NewNop(),
				config:  &config{},
				handler: handlerMock,
			}
			p.handle(netConnMock)
		})
	}
}
