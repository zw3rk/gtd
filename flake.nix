{
  description = "Claude GTD - A SQLite-driven CLI task management tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        claude-gtd = pkgs.buildGoModule {
          pname = "claude-gtd";
          version = "0.1.0";
          src = ./.;
          vendorHash = null; # Will be updated when we add dependencies
          
          # Enable CGO for SQLite support
          CGO_ENABLED = 1;
          
          # Static linking flags
          ldflags = [
            "-s"
            "-w"
            "-linkmode external"
            "-extldflags '-static'"
          ];
          
          # Ensure we have static libraries for SQLite
          buildInputs = with pkgs; [
            sqlite.dev
          ] ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
            musl
          ];
          
          nativeBuildInputs = with pkgs; [
            pkg-config
          ] ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
            musl
          ];
        };
      in
      {
        packages = {
          default = claude-gtd;
          claude-gtd = claude-gtd;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go development
            go
            gopls
            go-tools
            golangci-lint
            delve
            
            # Build tools
            gnumake
            pkg-config
            
            # SQLite development
            sqlite
            sqlite-interactive
            
            # Development utilities
            git
            ripgrep
            jq
            
            # For static builds (Linux only)
          ] ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
            musl
            musl.dev
          ];
          
          shellHook = ''
            echo "Claude GTD Development Environment"
            echo "Go version: $(go version)"
            echo "SQLite version: $(sqlite3 --version | cut -d' ' -f1)"
            echo ""
            echo "Available commands:"
            echo "  make build    - Build the binary"
            echo "  make test     - Run tests"
            echo "  make lint     - Run linter"
            echo ""
          '';
        };
      });
}