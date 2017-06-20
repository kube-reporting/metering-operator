// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

//
// System calls for 386, Windows are implemented in runtime/syscall_windows.goc
//

TEXT ·getprocaddress(SB), 7, $0-8
	JMP	syscall·getprocaddress(SB)

TEXT ·loadlibrary(SB), 7, $0-4
	JMP	syscall·loadlibrary(SB)
