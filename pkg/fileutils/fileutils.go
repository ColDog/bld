package fileutils

import "os"

// File modes.
const (
	Setuid os.FileMode = 1 << (12 - 1 - iota)
	Setgid
	Sticky
	UserRead
	UserWrite
	UserExecute
	GroupRead
	GroupWrite
	GroupExecute
	OtherRead
	OtherWrite
	OtherExecute
)

// Safe defaults for file permissions.
const (
	Regular = Setuid | Setgid | Sticky | UserRead | UserWrite | GroupRead |
		GroupWrite | OtherRead | OtherWrite
	Executable = Setuid | Setgid | Sticky | UserRead | UserWrite | UserExecute |
		GroupRead | GroupWrite | GroupExecute | OtherExecute | OtherRead |
		OtherWrite
	Directory = Executable
)
