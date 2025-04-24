# vipe-go
A Go implementation of the `vipe` command from the [moreutils](https://joeyh.name/code/moreutils/) package

# Usage
```
command1 | vipe | command2
```
The following flags are supported:
- `--suffix <suffix>`, which specifies the file extension of the temporary file
- `--editor <path>`, which specifies the path to the editor to use, overriding the values of the `EDITOR` and `VISUAL` environment variables
- `--use-exact-path`, which disables all special editor path processing and uses the exact value of the editor path as the editor to use
    - Special editor path processing includes the following:
        - Leading and trailing quotes in the editor path allows specifying an editor path with spaces in it, such as `"C:\Users\User\folder with space\editor.exe"` and `"/home/user/directory with space/editor"`
            - On Unix, if this is the case, the editor path is passed into the `sh` command with the `-c` flag, like this: `sh -c "/home/user/directory with space/editor"`

The following environment variables are recognized from top to bottom, with the topmost and bottommost environment variables having the lowest and highest priorities respectively:
- `EDITOR`, which specifies the path to the editor to use
- `VISUAL`, which also specifies the path to the editor to use

# Installation
First, [install Go](https://go.dev/doc/install) if it's not installed already. Then run the following commands to build:
```
git clone https://github.com/AlanLuu/vipe-go.git
cd vipe-go
go build
```
This will create an executable binary called `vipe` on Linux/macOS and `vipe.exe` on Windows.

# Features
- Compiles into a single executable that can run anywhere, no dependency on Perl
- Supports Windows
    - On Windows, the default editor used by this implementation is `notepad.exe`
- On non-Windows, the default editor used by this implementation is `vi`

# License
vipe is distributed under the terms of the [GPL-2.0 License](https://github.com/AlanLuu/vipe-go/blob/main/LICENSE).
