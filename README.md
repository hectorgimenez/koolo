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
- Run d2.install.reg to install the required registry key.
- Configure custom bot settings under `config/config.yaml` and `config/pickit/*.nip` files for pickit rules.
- Open Diablo II: Resurrected and wait until character selection screen.
- Run `koolo.exe`.

## Features

- Blizzard Sorceress and Hammerdin are currently supported
- Supported runs: Countess, Andariel, Ancient Tunnels, Summoner, Mephisto, Council, Eldritch, Pindleskin, Nihlathak, Tristram, Lower Kurast, Baal (WIP), Diablo (WIP)
- Bot integration for Discord and Telegram
- "Companion mode" one leader bot will be creating games and the rest of the bots will join the game... and sometimes it works
- Pickit based on NIP files
- Auto potion for health and mana (also mercenary)
- Chicken when low health
- Inventory slot locking
- Revive mercenary
- CTA buff and class buffs
- Auto repair
- Skip on immune

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
