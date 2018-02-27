# CHIP-8

![chip8](https://raw.githubusercontent.com/odsod/chip8/gh-pages/chip8.gif)

## What's this

Go implementation of the [CHIP-8 VM][chip8] according to [Cowgod's Technical Reference][cowgod].

Roms downloaded from [Zophar's Domain][zophar].

With ideas from:

* [Michael Fogleman's NES emulator][fogleman]
* [Mastering CHIP-8][mastering]

## Dependencies

~~~
github.com/nsf/termbox-go
github.com/go-gl/gl/v2.1/gl
github.com/go-gl/glfw/v3.1/glfw
~~~

## Example usage

~~~sh
# Terminal UI
chip8 -rom roms/TETRIS

# OpenGL UI
chip8-gl -rom roms/TETRIS
~~~

[chip8]: https://en.wikipedia.org/wiki/CHIP-8
[cowgod]: http://devernay.free.fr/hacks/chip8/C8TECH10.HTM
[zophar]: https://www.zophar.net/pdroms/chip8/chip-8-games-pack.html
[mastering]: http://mattmik.com/files/chip8/mastering/chip8.html
[fogleman]: https://github.com/fogleman/nes
