# Koolo
Bot for Diablo II Resurrected. Koolo project was built for informational and educational purposes only, it's not designed
for online usage. Feel free to contribute opening pull requests with new features or bugfixes.

## How it works
This bot is based on memory reading, in order to do that it's using a modified version of [MapAssist](https://github.com/OneXDeveloper/MapAssist)
that provides all the required information to the bot. The bot simulates clicks/key pushes to move.

## Requirements
- Diablo II: Resurrected
- [Diablo II: LOD 1.13c (Required by MapAssist)](https://drive.google.com/file/d/1smkzc8kHnL86Ac1b0JuCN_O9RO9MJ-oQ/view)
- [.NET Framework 4.7.2 Runtime](https://dotnet.microsoft.com/en-us/download/dotnet-framework/net472)
- [Visual C++ for VS 2015 (x64 and x86)](https://www.microsoft.com/en-us/download/details.aspx?id=48145)

## Getting started
- If you haven't done yet, [install Diablo II: LOD 1.13c](https://github.com/OneXDeveloper/MapAssist/wiki/Installation#step-1-d2-lod-setup)
- Open `MapAssist/KooloMA.exe` and set your Diablo II: LOD directory, close it.
- Configure custom bot settings under `config/config.yaml` and `config/pickit.yaml` files.
- Open Diablo II: Resurrected and wait until character selection screen.
- Run `koolo.exe`.