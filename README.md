# Koolo
Bot for Diablo II Resurrected. Koolo project was built for informational and educational purposes only, it's not designed
for online usage. Feel free to contribute opening pull requests with new features or bugfixes.

## How it works
This bot is based on memory reading, also uses https://github.com/joffreybesos/rustdecrypt and https://github.com/blacha/diablo2/tree/master/packages/map

## Requirements
- Diablo II: Resurrected
- [Diablo II: LOD 1.13c (Required by MapServer)](https://drive.google.com/file/d/1smkzc8kHnL86Ac1b0JuCN_O9RO9MJ-oQ/view)
- [Docker for Windows](https://docs.docker.com/desktop/install/windows-install/)

## Getting started
- If you haven't done yet, [install Diablo II: LOD 1.13c](https://github.com/OneXDeveloper/MapAssist/wiki/Installation#step-1-d2-lod-setup)
- Edit `config/config.yaml` and ensure `D2Path` is pointing to your Diablo II: LOD 1.13c installation directory.
The directory will be mounted as Docker volume to the blacha mapserver.
- Configure custom bot settings under `config/config.yaml` and `config/pickit.yaml` files.
- Open Diablo II: Resurrected and wait until character selection screen.
- Run `koolo.exe`.