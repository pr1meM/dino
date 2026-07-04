<h1 align="center">
   Dino
</h1>
<p align= "center">
   <kbd>
   <img  src="https://raw.githubusercontent.com/pr1mem/dino/main/images/logo.png">
   </kbd><br><br>
   <img src="https://img.shields.io/github/languages/top/pr1mem/dino">
   <img src="https://img.shields.io/github/stars/pr1mem/dino">
   <img src="https://img.shields.io/github/forks/pr1mem/dino">
   <br>
   <img src="https://img.shields.io/github/last-commit/pr1mem/dino">
   <img src="https://img.shields.io/github/license/pr1mem/dino">
   <br>
   <img src="https://img.shields.io/github/issues/pr1mem/dino">
   <img src="https://img.shields.io/github/issues-closed/pr1mem/dino">
   <br>
   <br>
</p>

A lightweight terminal text editor written in Go.

## Features

- Tabs for multiple open files
- File tree sidebar
- Syntax highlighting for pretty much any language
- Search
- Copy/cut/paste, undo/redo, auto-indent, auto-pairing of brackets/quotes
- Mouse support
- Ctrl+H shows all keybindings


## Build from source

```sh
git clone https://github.com/pr1mem/dino
cd dino
go build -o dino .
```

## Install system-wide

To install `dino` so you can run it from anywhere:

```sh
./install.sh
```

This builds the binary and copies it to `/usr/local/bin`. Once installed, you can open any file with:

```sh
dino <filename>
```


## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
