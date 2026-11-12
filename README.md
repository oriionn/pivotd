# pivotd
A simple daemon to autorotate your screen with DDC/CI. It's made for external portable screen with a autorotate function.

## Supported desktops
| Desktop    | Tool used        | Notes                                                       |
|------------|------------------|-------------------------------------------------------------|
| Hyprland   | `hyprctl`        | Needs a Hyprland version with Lua config (`hyprctl eval`)   |
| KDE Plasma | `kscreen-doctor` |                                                             |

The desktop is detected at runtime from the session environment:
- **KDE Plasma**: `XDG_CURRENT_DESKTOP` contains `KDE`, or `KDE_FULL_SESSION=true`
- **Hyprland**: `XDG_CURRENT_DESKTOP` contains `Hyprland`, or `HYPRLAND_INSTANCE_SIGNATURE` is set

Other desktops (GNOME, Sway, X11 sessions…) are not supported yet: pivotd keeps running but logs an `unsupported desktop` error instead of rotating.

## Development
### Requirements
- Go
- make
- ddcutil / libddcutil

### Compile
```sh
make
```

## Installation
### Runtime requirements
- A monitor that reports its orientation over DDC/CI
- [ddcutil](https://www.ddcutil.com/) (`libddcutil`)
- One of the supported desktops 
- Access to the I2C buses (`/dev/i2c-*`), see [I2C access](#i2c-access)

### NixOS
Add the flake to your inputs and enable the module:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    pivotd = {
      url = "git+https://git.oriondev.fr/orion/pivotd";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { nixpkgs, pivotd, ... }: {
    nixosConfigurations.<hostname> = nixpkgs.lib.nixosSystem {
      modules = [
        pivotd.nixosModules.default
        {
          services.pivotd.enable = true;
          users.users.<username>.extraGroups = [ "i2c" ];
        }
      ];
    };
  };
}
```

The module enables `hardware.i2c`, installs the package and starts `pivotd` as a
systemd user service with your graphical session. Rebuild, then log out and back in
so the `i2c` group membership is applied.

### I2C access
pivotd talks to the monitor through `/dev/i2c-*`. On non-NixOS systems:

```sh
# Load the i2c-dev module now and at every boot
sudo modprobe i2c-dev
echo i2c-dev | sudo tee /etc/modules-load.d/i2c-dev.conf

# Give your user access to the I2C devices
sudo groupadd --system i2c   # if the group does not exist yet
sudo usermod -aG i2c $USER
```

If `/dev/i2c-*` is not owned by the `i2c` group on your distribution, add a udev rule
(ddcutil ships one as `60-ddcutil-i2c.rules`), then log out and back in.

You can check that your monitor is reachable with:
```sh
ddcutil detect
```
