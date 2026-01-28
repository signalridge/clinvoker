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
          clinvk = pkgs.buildGoModule {
            pname = "clinvk";
            inherit version;
            src = ./.;

            subPackages = [ "cmd/clinvk" ];

            vendorHash = "sha256-mVSH+lVn6NavaScayx2DEgnKbRV1nC5LansdCwABV/k=";

            # Use -short to skip integration tests that require writable HOME
            checkFlags = [ "-short" ];

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
              mainProgram = "clinvk";
            };
          };

          # Alias for backwards compatibility
          clinvoker = self.packages.${system}.clinvk;

          default = self.packages.${system}.clinvk;
        };

        apps = {
          clinvk = flake-utils.lib.mkApp {
            drv = self.packages.${system}.clinvk;
          };
          # Alias for backwards compatibility
          clinvoker = self.apps.${system}.clinvk;
          default = self.apps.${system}.clinvk;
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
        clinvk = self.packages.${prev.system}.clinvk;
        # Alias for backwards compatibility
        clinvoker = self.packages.${prev.system}.clinvk;
      };
    };
}
