<p align="center">
  <img src="assets/koolo.webp" alt="Koolo" width="150">
</p>
<h3 align="center">Koolo</h3>

---

Koolo is a small bot for Diablo II: Resurrected (Expansion). Koolo project was built for informational and educational purposes
only, it's not intended for online usage. Feel free to contribute opening pull requests with new features or bugfixes.
Koolo reads game memory and interacts with the game injecting clicks/keystrokes to the game window. As good as it can.

Feel free to join our Discord community to report bugs, ask for help or just to chat: [Koolo Discord]( https://discord.gg/zgFMyzAFHE)

## Disclaimer
Can I get banned for using Koolo? The answer is a crystal clear yes, you can get banned although at this point I'm
not aware of any ban for using it. I'm not responsible for any ban or any other consequence that may arise from it.

## Features
- Blizzard Sorceress, Nova Sorceress, FoH, Berserk Barbarian Hork (Travincal), Mosaic are currently supported. Hammerdin, Javazon and Winddruid are WIP
- Supported runs: Countess, Andariel, Ancient Tunnels, Summoner, Mephisto, Council, Eldritch-Shenk, Endugu, Drifter Cavern, Pindleskin, Nihlathak,
  Tristram, Lower Kurast and Superchests, Stony Tomb, The Pit, Arachnid Lair, Baal, Duriel, Tal Rasha Tombs, Diablo, Cows, Treshsocket
- Multi window support (run multiple bots at the same time)
- Bot integration for Discord and Telegram
- "Companion mode" one leader bot will be creating games and the rest of the bots will join the game... (not working currently)
- Pickit based on NIP files
- Auto potion for health and mana (also mercenary)
- Chicken when low health
- Inventory slot locking
- Revive mercenary
- CTA buff and class buffs
- Auto repair
- Skip on immune
- Auto leveling sorceress and paladin (WIP) this feature is not finished.
- Auto gambling
- Auto cubing and crafting (WIP)
- Terror Zones (WIP)
- Classic is not supported

## Requirements
- Diablo II: Resurrected (1280x720 required, windowed mode, ensure accessibility large fonts disabled)
- **Diablo II: LOD 1.13c** (IMPORTANT: It will **NOT** work without it, this step is not optional)

## Quick Start
### Preparing the character
- Koolo will read game keybindings in order to use the skills, doesn't matter what key is used, but the skills for the build must be set.
- For blizzard sorceress, set the **left** skill to Glacial Spike or Ice Blast, and for Hammerdin to Blessed Hammer.
- Foh set keybind for FOH and holybolt on left skill, conviction on right skill
- Berserk barb set berserk as left skill. Also to use FindItem you need higher goldfind on secondary weapons slot. Alibaba + anything will work.
- Buy TP and ID tomes and one stack of keys and keep them in the inventory, additionally set the TP tome to a key binding, this is **required**.
- Horadric Cube can be stashed or kept in inventory, Koolo will use it to cube recipes if enabled.
- Keep the charms in the inventory, Koolo can be configured to lock specific inventory slots.

### Running the tool
- If you haven't done yet, install **Diablo II: LOD 1.13c** (required)
- [Download](https://github.com/hectorgimenez/koolo/releases) the latest Koolo release (recommended for most users), or alternatively you can [build it from source](#development-environment)
- Extract the zip file in a directory of your choice.
- Run `koolo.exe`.
- Follow the setup wizard, it will guide you through the process of setting up the bot, you will need to setup some directories and character configuration.
- If you want to back up/restore your configuration, and for manual setup, you can find the configuration files in the `config` directory.

## Pickit rules
Item pickit is based on [NIP files](https://github.com/blizzhackers/pickits/blob/master/NipGuide.md), you can find them in the `config/{character}/pickit` directory.

All the .nip files contained in the pickit directory will be loaded, so you can have multiple pickit files.

There are some considerations to take into account:
- If item fully matches the pickit rule before being identified, it will be picked up and stashed unidentified.
- If item doesn't match the full rule, will be identified and checked again, if fully matches a rule it will be stashed otherwise sold to vendor.
- If there is an error on the NIP file or Koolo can not understand it, the application will not start.
- Pickit rules can not be changed in runtime (yet), you will need to restart Koolo to apply changes.

## Development environment
**Note:** This is only required if you want to build the project from source. If you want to run the bot, you can just download the [latest release](https://github.com/hectorgimenez/koolo/releases).

Setting the development environment is pretty straightforward, but the following dependencies are **required** to build the project.

### Dependencies
- [Download Go >= 1.23](https://go.dev/dl/)
- [Install git](https://gitforwindows.org/)

### Building from source
Open the terminal and run the following commands in project root directory:
```shell
git clone https://github.com/hectorgimenez/koolo.git
cd koolo
build.bat
```
This will produce the "build" directory with the executable file and all the required assets.

### Updating with latest changes
In order to fetch latest `main` branch changes run the following commands in project root directory:
```shell
git pull
build.bat
```
**Note**: `build` directory **will be deleted**, so if you customized any file in there, make sure to backup it before running `build.bat`.
