# This file can be submitted to nixpkgs as:
# pkgs/by-name/cl/clinvk/package.nix
{
  lib,
  buildGoModule,
  fetchFromGitHub,
}:

buildGoModule rec {
  pname = "clinvk";
  version = "0.1.0"; # Update on release

  src = fetchFromGitHub {
    owner = "signalridge";
    repo = "clinvoker";
    rev = "v${version}";
    hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # Update with: nix-prefetch-github signalridge clinvoker --rev v${version}
  };

  vendorHash = "sha256-2dK5503Vd5reKxpVEglSTU4HRZKm3jpnLCnDadZb6t0=";

  subPackages = [ "cmd/clinvk" ];

  # Use -short to skip integration tests that require writable HOME
  checkFlags = [ "-short" ];

  ldflags = [
    "-s"
    "-w"
    "-X github.com/signalridge/clinvoker/internal/app.version=${version}"
  ];

  meta = {
    description = "Unified AI CLI wrapper for Claude Code, Codex CLI, and Gemini CLI";
    homepage = "https://github.com/signalridge/clinvoker";
    license = lib.licenses.mit;
    maintainers = with lib.maintainers; [ ]; # Add your nixpkgs maintainer name
    mainProgram = "clinvk";
  };
}
