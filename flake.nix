{
  description = "Unified AI CLI wrapper for multiple backends";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = if (self ? rev) then self.shortRev else "dev";
      in
      {
        packages = {
          clinvoker = pkgs.buildGoModule {
            pname = "clinvoker";
            inherit version;
            src = ./.;

            subPackages = [ "cmd/clinvoker" ];

            vendorHash = "sha256-GT9lkHVG0Q3AO9IzI840Oyt1u5OUgQAaBncuCFosF+Y=";

            ldflags = [
              "-s"
              "-w"
              "-X github.com/signalridge/clinvoker/internal/app.version=${version}"
              "-X github.com/signalridge/clinvoker/internal/app.commit=${version}"
              "-X github.com/signalridge/clinvoker/internal/app.date=1970-01-01"
            ];

            meta = with pkgs.lib; {
              description = "Unified AI CLI wrapper for Claude Code, Codex CLI, and Gemini CLI";
              homepage = "https://github.com/signalridge/clinvoker";
              license = licenses.mit;
              maintainers = [ ];
              mainProgram = "clinvoker";
            };
          };

          default = self.packages.${system}.clinvoker;
        };

        apps = {
          clinvoker = flake-utils.lib.mkApp {
            drv = self.packages.${system}.clinvoker;
          };
          default = self.apps.${system}.clinvoker;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            golangci-lint
            goreleaser
          ];

          shellHook = ''
            echo "clinvoker development shell"
            echo "Go version: $(go version)"
          '';
        };
      }
    )
    // {
      overlays.default = final: prev: {
        clinvoker = self.packages.${prev.system}.clinvoker;
      };
    };
}
