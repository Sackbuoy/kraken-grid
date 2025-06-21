{
  description = "Go project with Docker image";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    go-flake.url = "github:Sackbuoy/flakes?dir=go/go";
    golangci-lint-flake.url = "github:Sackbuoy/flakes?dir=go/golangci-lint";
  };

  outputs = {
    self,
    nixpkgs,
    go-flake,
    golangci-lint-flake,
    ...
  }: let
    system = "x86_64-linux";
    pkgs = import nixpkgs {inherit system;};
    pname = "kraken-grid";
    author = "Sackbuoy";

    golangciPackage = "latest";
    goPackage = go-flake.lib.getVersion ./go.mod;

    goBuild = pkgs.buildGoModule {
      inherit pname;
      version = "0.1.0";
      src = ./.;
      vendorHash = null; # Will be updated on first build
      buildInputs = [ pkgs.cacert ];
    };

    dockerImage = pkgs.dockerTools.buildImage {
      name = "ghcr.io/${author}/${pname}";
      tag = "latest";
      created = "now";

      copyToRoot = pkgs.buildEnv {
        name = "image-root";
        paths = [
          self.packages.${system}.goBuild
          pkgs.coreutils
          pkgs.shadow
          pkgs.bashInteractive
          pkgs.cacert
        ];
        pathsToLink = ["/bin" "/etc" "/home" "/var"];
      };

      config = {
        Cmd = ["/bin/${pname}"];
        WorkingDir = "/app";
        Volumes = {
          "/home/nonroot/.kube" = {};
        };
        User = "nonroot:nonroot";
      };

      runAsRoot = ''
        #!${pkgs.runtimeShell}
        ${pkgs.dockerTools.shadowSetup}
        groupadd -r nonroot
        useradd -m -r -g nonroot nonroot

        mkdir -p /app /tmp
        chmod 1777 /tmp

        mkdir /home/nonroot/.kube
        touch /home/nonroot/.kube/config
        chmod 744 /home/nonroot/.kube
        chmod 644 /home/nonroot/.kube/config
        chown -R nonroot:nonroot /home/nonroot /app
      '';
    };
  in {
    packages.${system} = {
      goBuild = goBuild;
      docker = dockerImage;
      default = dockerImage;
    };

    devShells.${system}.default = pkgs.mkShell {
      buildInputs = with pkgs; [
        golangci-lint-flake.packages.${system}.${golangciPackage}
        go-flake.packages.${system}.${goPackage}
        gopls
        gotools
        go-outline
        delve
        docker
      ];

      CGO_CFLAGS = "-O2";

      env = {
        GO111MODULE = "on";
      };

    };
  };
}
