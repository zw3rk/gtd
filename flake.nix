{
  description = "GTD - A SQLite-driven CLI task management tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs = inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      # Define the systems we want to support
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];

      perSystem = { config, self', inputs', pkgs, system, ... }: {
        # Define the main GTD package
        packages.gtd = pkgs.buildGoModule {
          pname = "gtd";
          version = "0.1.0";
          src = ./.;
          vendorHash = null;
          
          # Tests need to run with dynamic linking, not static
          doCheck = false;
          
          # Ensure CGO is enabled
          env.CGO_ENABLED = "1";
          
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

        # Define all packages
        packages = {
          default = config.packages.gtd;
          gtd-release = pkgs.runCommand "gtd-release-${system}" {
            nativeBuildInputs = [ pkgs.zip ];
          } ''
            mkdir -p $out/nix-support
            
            # Copy the binary
            mkdir -p $out/bin
            cp ${config.packages.gtd}/bin/gtd $out/bin/
            
            # Create release info
            echo "file binary-dist $out/gtd-${system}.zip" >> $out/nix-support/hydra-build-products
            echo "doc readme $out/README" >> $out/nix-support/hydra-build-products
            
            # Create a README
            cat > $out/README << EOF
            GTD Task Management Tool
            ========================
            
            System: ${system}
            Binary: gtd
            
            Usage:
              ./gtd --help
            
            This is a SQLite-driven CLI task management tool following GTD methodology.
            EOF
            
            # Create zip archive
            cd $out/bin
            zip -r $out/gtd-${system}.zip gtd
          '';
        };

        # Development shell
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
            sqlite.dev
            sqlite-interactive
            
            # Development utilities
            git
            ripgrep
            jq
          ] ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
            # Linux needs gcc for CGO
            gcc
          ] ++ pkgs.lib.optionals pkgs.stdenv.isDarwin [
            # Darwin needs clang
            clang
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
            
            # Ensure CGO can find SQLite
            export CGO_ENABLED=1
            
            # Set proper CGO flags for Linux to avoid segfaults
            ${pkgs.lib.optionalString pkgs.stdenv.isLinux ''
              export CGO_CFLAGS="-I${pkgs.sqlite.dev}/include"
              export CGO_LDFLAGS="-L${pkgs.sqlite.out}/lib -lsqlite3"
            ''}
          '';
        };
      };

      flake = {
        # Hydra CI job configuration
        hydraJobs =
          let
            systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
          in
          {
            # Build the main package on all systems
            build = inputs.nixpkgs.lib.genAttrs systems (system:
              inputs.self.packages.${system}.gtd
            );
            
            # Build release packages for hydra downloads
            release = inputs.nixpkgs.lib.genAttrs systems (system:
              inputs.self.packages.${system}.gtd-release
            );
          };
      };
    };
}
