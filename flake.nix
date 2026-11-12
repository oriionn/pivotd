{
    description = "Automatic monitor rotation via DDC/CI";
    inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    outputs =
        { self, nixpkgs }: let
            systems = [
                "x86_64-linux"
                "aarch64-linux"
            ];
            forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
        in
        {
            packages = forAllSystems (pkgs: {
                default = self.packages.${pkgs.stdenv.hostPlatform.system}.pivotd;

                pivotd = pkgs.buildGoModule {
                    pname = "pivotd";
                    version = "1.0.0";
                    src = self;

                    vendorHash = null;
                    subPackages = [ "src" ];

                    nativeBuildInputs = [ pkgs.pkg-config ];
                    buildInputs = [ pkgs.ddcutil ];

                    postInstall = ''
                        mv $out/bin/src $out/bin/pivotd

                        install -Dm644 assets/pivotd.service -t $out/lib/systemd/user
                        substituteInPlace $out/lib/systemd/user/pivotd.service \
                            --replace-fail /usr/bin/pivotd $out/bin/pivotd
                    '';

                    meta = {
                        description = "Automatic monitor rotation via DDC/CI (Hyprland & KDE)";
                        mainProgram = "pivotd";
                        platforms = pkgs.lib.platforms.linux;
                    };
                };
            });

            nixosModules.default = {
                config,
                lib,
                pkgs,
                ...
            }: let cfg = config.services.pivotd; in {
                options.services.pivotd = {
                    enable = lib.mkEnableOption "automatic monitor rotation via DDC/CI";
                    package = lib.mkOption {
                        type = lib.types.package;
                        default = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
                        description = "The pivotd package to use.";
                    };
                };

                config = lib.mkIf cfg.enable {
                    hardware.i2c.enable = true;

                    environment.systemPackages = [ cfg.package ];
                    systemd.packages = [ cfg.package ];

                    systemd.user.services.pivotd = {
                        wantedBy = [ "graphical-session.target" ];
                        environment.PATH = lib.mkForce "/run/wrappers/bin:/etc/profiles/per-user/%u/bin:/run/current-system/sw/bin";
                    };
                };
            };
        };
}
