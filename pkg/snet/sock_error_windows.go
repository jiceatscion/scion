// Copyright 2024 OVGU Magdeburg
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package snet

import (
	"errors"

	"golang.org/x/sys/windows"
)

const (
	ErrAddrInUse = windows.WSAEADDRINUSE
)

// errorIsAddrUnavailable checks whether the error returned from a syscall to
// bind indicates that the requested address is not available.
func errorIsAddrUnavailable(err error) bool {
	// WSAEADDRINUSE is returned if another socket is bound to the same address.
	// WSAEACCES is returned if another process is bound to the same address
	// with exclusive access.
	return errors.Is(err, windows.WSAEADDRINUSE) || errors.Is(err, windows.WSAEACCES)
}
