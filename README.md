![Koolo](images/you_died.png)
<h1 align="center">Koolo</h1>

---

Koolo is a small bot for Diablo II: Resurrected. Koolo project was built for informational and educational purposes
only, it's not intended for online usage. Feel free to contribute opening pull requests with new features or bugfixes.

## How it works

Koolo reads game memory and interacts with the game simulating clicks/keystrokes.
It uses the following third party libraries:

- https://github.com/joffreybesos/rustdecrypt
- https://github.com/hectorgimenez/d2go
- https://github.com/hectorgimenez/diablo2 (forked from https://github.com/blacha/diablo2/tree/master/packages/map)

## Requirements

- Diablo II: Resurrected
- Diablo II: LOD 1.13c (required by https://github.com/blacha/diablo2/tree/master/packages/map)

## Quick Start

- If you haven't done yet, install Diablo II: LOD 1.13c
- Edit `config/config.yaml` and ensure `D2LoDPath` is pointing to your Diablo II: LOD 1.13c installation directory.
- Configure custom bot settings under `config/config.yaml` and `config/pickit/*.nip` files for pickit rules.
- Open Diablo II: Resurrected and wait until character selection screen.
- Run `koolo.exe`.

## Development environment

Setting the development environment is pretty straightforward, but MinGW is required to build the project.

### Dependencies

- Download [MingGW](https://sourceforge.net/projects/mingw-w64/files/) ```x86_64-win32-seh``` should be fine, extract it
  and add it to the system PATH on Windows environment variables.
- [Download Go >= 1.20](https://go.dev/dl/)
- [Install git](https://gitforwindows.org/)

```
git clone https://github.com/hectorgimenez/koolo.git
cd koolo
go run cmd/koolo/main.go
```

To produce a .exe build and prepare all the assets, the ```build.bat``` script can be used.
