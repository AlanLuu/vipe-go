package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func anyFlagProvided(names ...string) bool {
	found := false
	for _, name := range names {
		flag.Visit(func(f *flag.Flag) {
			if f.Name == name {
				found = true
			}
		})
		if found {
			break
		}
	}
	return found
}

func handleError(err error, str string) int {
	const vipe = "vipe: "
	if err != nil {
		fmt.Fprintln(os.Stderr, vipe+err.Error())
	}
	if str != "" {
		fmt.Fprintln(os.Stderr, vipe+str)
	}
	return 1
}

func isAndroid() bool {
	return runtime.GOOS == "android"
}

func isLinux() bool {
	return runtime.GOOS == "linux"
}

func isLinuxOrAndroid() bool {
	return isLinux() || isAndroid()
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func joinSlice(s []string, sep rune) string {
	endIndex := len(s) - 1
	var builder strings.Builder
	for i, e := range s {
		builder.WriteString(e)
		if i < endIndex {
			builder.WriteRune(sep)
		}
	}
	return builder.String()
}

func passedToShDashC(s []string) bool {
	return len(s) >= 2 && s[0] == "sh" && s[1] == "-c"
}

func stripTrailingQuotes(str string) (string, bool) {
	hasQuotePrefix :=
		strings.HasPrefix(str, "\"") ||
			strings.HasPrefix(str, "'")
	hasQuoteSuffix :=
		strings.HasSuffix(str, "\"") ||
			strings.HasSuffix(str, "'")
	if hasQuotePrefix || hasQuoteSuffix {
		if !isWindows() {
			return str, true
		}
		return strings.Trim(str, "\"'"), true
	}
	return "", false
}

func realMain() int {
	var (
		editorFlag = flag.String(
			"editor",
			"",
			"Path to editor to use",
		)
		suffixFlag = flag.String(
			"suffix",
			"",
			"File extension of the temporary file",
		)
		useExactPathFlag = flag.Bool(
			"use-exact-path",
			false,
			"Use exact editor path without any special path processing",
		)
	)
	flag.Parse()

	stdinStat, stdinStatErr := os.Stdin.Stat()
	if stdinStatErr != nil {
		return handleError(stdinStatErr, "")
	}
	stdoutStat, stdoutStatErr := os.Stdout.Stat()
	if stdoutStatErr != nil {
		return handleError(stdoutStatErr, "")
	}
	stdinFromPipe := func() bool {
		return stdinStat.Mode()&os.ModeCharDevice == 0
	}
	stdoutToTerminal := func() bool {
		return stdoutStat.Mode()&os.ModeCharDevice != 0
	}

	suffix := *suffixFlag
	if suffix != "" && !strings.HasPrefix(suffix, ".") {
		suffix = "." + suffix
	}
	tempFile, tempFileErr := os.CreateTemp("", "vipe-*"+suffix)
	if tempFileErr != nil {
		return handleError(tempFileErr, "")
	}
	tempFileName := tempFile.Name()
	defer os.Remove(tempFileName)

	//Don't read from stdin if it's from a terminal
	if stdinFromPipe() {
		_, writeErr := io.Copy(tempFile, os.Stdin)
		if writeErr != nil {
			tempFile.Close()
			return handleError(writeErr, "")
		}
	}
	tempFile.Close()

	var editor []string
	switch {
	case anyFlagProvided("editor"):
		value := *editorFlag
		if *useExactPathFlag || value == "" {
			editor = []string{value}
		} else if newValue, ok := stripTrailingQuotes(value); ok {
			if !isWindows() {
				editor = []string{"sh", "-c", newValue}
			} else {
				editor = []string{newValue}
			}
		} else {
			editor = strings.Fields(value)
			if len(editor) == 0 {
				editor = []string{value}
			}
		}
	default:
		if isWindows() {
			editor = []string{"notepad.exe"}
		} else {
			editor = []string{"vi"}
			const path = "/usr/bin/editor"
			if _, err := os.Stat(path); err == nil {
				editor[0] = path
			} else if isLinuxOrAndroid() {
				const termuxPath = "/data/data/com.termux/files" + path
				if _, err := os.Stat(termuxPath); err == nil {
					editor[0] = termuxPath
				}
			}
		}
		if value, ok := os.LookupEnv("EDITOR"); ok {
			if *useExactPathFlag || value == "" {
				editor = []string{value}
			} else if newValue, ok := stripTrailingQuotes(value); ok {
				if !isWindows() {
					editor = []string{"sh", "-c", newValue}
				} else {
					editor = []string{newValue}
				}
			} else {
				editor = strings.Fields(value)
				if len(editor) == 0 {
					editor = []string{value}
				}
			}
		}
		if value, ok := os.LookupEnv("VISUAL"); ok {
			if *useExactPathFlag || value == "" {
				editor = []string{value}
			} else if newValue, ok := stripTrailingQuotes(value); ok {
				if !isWindows() {
					editor = []string{"sh", "-c", newValue}
				} else {
					editor = []string{newValue}
				}
			} else {
				editor = strings.Fields(value)
				if len(editor) == 0 {
					editor = []string{value}
				}
			}
		}
	}
	editor = append(editor, tempFileName)

	cmd := exec.Command(editor[0], editor[1:]...)
	if stdinFromPipe() {
		//Set stdin of editor cmd to terminal if os.stdin is from pipe
		var tty *os.File
		var ttyErr error
		if isWindows() {
			tty, ttyErr = os.OpenFile("CONIN$", os.O_RDONLY, 0)
		} else {
			tty, ttyErr = os.OpenFile("/dev/tty", os.O_RDONLY, 0)
		}
		if ttyErr != nil {
			return handleError(ttyErr, "")
		}
		cmd.Stdin = tty
	} else {
		cmd.Stdin = os.Stdin
	}
	if !stdoutToTerminal() {
		//Set stdout of editor cmd to terminal if os.stdout is not to terminal
		var tty *os.File
		var ttyErr error
		if isWindows() {
			tty, ttyErr = os.OpenFile("CONOUT$", os.O_RDWR, 0)
		} else {
			tty, ttyErr = os.OpenFile("/dev/tty", os.O_WRONLY, 0)
		}
		if ttyErr != nil {
			return handleError(ttyErr, "")
		}
		cmd.Stdout = tty
	} else {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr
	runErr := cmd.Run()
	if cmd.Stdin != os.Stdin {
		cmd.Stdin.(*os.File).Close()
	}
	if cmd.Stdout != os.Stdout {
		cmd.Stdout.(*os.File).Close()
	}
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			var joinedStr string
			if passedToShDashC(editor) {
				joinedStr = joinSlice(editor[2:], ' ')
			} else {
				joinedStr = joinSlice(editor, ' ')
			}
			if joinedStr == "" {
				joinedStr = "\"\""
			}
			return handleError(
				nil,
				fmt.Sprintf(
					"%v exited with a status of %v, aborting",
					joinedStr,
					exitErr.ExitCode(),
				),
			)
		}
		return handleError(runErr, "")
	}

	tempFile, tempFileErr = os.Open(tempFileName)
	if tempFileErr != nil {
		return handleError(tempFileErr, "")
	}
	defer tempFile.Close()
	_, writeErr := io.Copy(os.Stdout, tempFile)
	if writeErr != nil {
		return handleError(writeErr, "")
	}

	return 0
}

func main() {
	os.Exit(realMain())
}
