# Dino

A lightweight terminal text editor written in Go.

## Features

- Tabs for multiple open files
- File tree sidebar
- Syntax highlighting for pretty much any language
- Search
- Copy/cut/paste, undo/redo, auto-indent, auto-pairing of brackets/quotes
- Mouse support
- Ctrl+H shows all keybindings

## Build

```sh
git clone https://github.com/pr1mem/dino
```

```sh
cd dino
```


```sh
go build -o dino .
```

## Install

To install `dino` system-wide so you can just run `dino file` from
anywhere:

```sh
./install.sh
```

This builds the binary and copies it to `/usr/local/bin`.

You can start dino now from everywhere with `dino <filename>`


## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
