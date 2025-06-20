{
  description = "GTD - A SQLite-driven CLI task management tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      # Define the systems we want to support
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
    in
    flake-utils.lib.eachSystem supportedSystems (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        gtd = pkgs.buildGoModule {
          pname = "gtd";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-aJY9i1dmcoMvuQXyCwxH7k0LfjnKi+AtD0IpZzj0Rb8=";
          
          # Enable CGO for SQLite support
          env.CGO_ENABLED = "1";
          
          # Skip tests temporarily due to file permission issues in nix sandbox
          doCheck = false;
          
          # Static linking flags (Linux only for true static builds)
          ldflags = [
            "-s"
            "-w"
          ] ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
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
          
          # Fixup phase for macOS to use system libraries instead of Nix store paths
          postFixup = pkgs.lib.optionalString pkgs.stdenv.isDarwin ''
            echo "Fixing up macOS dependencies..."
            
            # Fix libresolv to use system library
            install_name_tool -change \
              ${pkgs.darwin.libresolv}/lib/libresolv.9.dylib \
              /usr/lib/libresolv.9.dylib \
              $out/bin/gtd
            
            echo "macOS dependency fixup complete"
            otool -L $out/bin/gtd
          '';
        };
      in
      {
        packages = {
          default = gtd;
          gtd = gtd;
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
            echo "GTD Development Environment"
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
      }) // {
        # Hydra CI job configuration
        hydraJobs = {
          # Build on key CI platforms
          build = nixpkgs.lib.genAttrs [ "x86_64-linux" "aarch64-linux" ] (system:
            let pkgs = nixpkgs.legacyPackages.${system}; in
            pkgs.buildGoModule {
              pname = "gtd";
              version = "0.1.0";
              src = ./.;
              vendorHash = "sha256-aJY9i1dmcoMvuQXyCwxH7k0LfjnKi+AtD0IpZzj0Rb8=";
              
              # Enable CGO for SQLite support
              env.CGO_ENABLED = "1";
              
              # Skip tests temporarily due to file permission issues in nix sandbox
              doCheck = false;
              
              # Static linking flags (Linux only for true static builds)
              ldflags = [
                "-s"
                "-w"
                "-linkmode external"
                "-extldflags '-static'"
              ];
              
              # Ensure we have static libraries for SQLite
              buildInputs = with pkgs; [
                sqlite.dev
                musl
              ];
              
              nativeBuildInputs = with pkgs; [
                pkg-config
                musl
              ];
            }
          );
        };
      };
}