{ pkgs ? import <nixpkgs> {} }:
  pkgs.mkShell {
    # nativeBuildInputs is usually what you want -- tools you need to run
    nativeBuildInputs = with pkgs.buildPackages; [
      go
      dbmate
      sqlite
      flyctl
    ];

    shellHook = ''
      if [ -f .env ]; then
        set -a
        source .env
        set +a
      fi
    '';
}
