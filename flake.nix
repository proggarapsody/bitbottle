{
  description = "bitbottle — gh-style Bitbucket CLI for Server/DC + Cloud";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";

  outputs = { self, nixpkgs }:
    let
      version = "1.49.0";
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;

      assetFor = system:
        let
          map = {
            "x86_64-linux"   = { os = "linux";  arch = "amd64"; sha256 = "448c59d8210038e1d41249edd598103dc39908b4f8a430291a22def46e6e3652"; };
            "aarch64-linux"  = { os = "linux";  arch = "arm64"; sha256 = "9fc8a3ae343fd60fd6d5e10c3bd49065cd1633afb3e2aeda2e647bd996e8e3e0"; };
            "x86_64-darwin"  = { os = "darwin"; arch = "amd64"; sha256 = "8b368bc7cb7e16f54b257a047351621d98a22a280ab1e6dfd198f1ff41cfa279"; };
            "aarch64-darwin" = { os = "darwin"; arch = "arm64"; sha256 = "cc64f939657bb6574b2d9f6d1aa0abfaa6f1dc0b9bc8bd1f0319df5bd0fb7a57"; };
          };
        in map.${system};
    in {
      packages = forAllSystems (system:
        let
          pkgs = import nixpkgs { inherit system; };
          asset = assetFor system;
          tarball = pkgs.fetchurl {
            url = "https://github.com/proggarapsody/bitbottle/releases/download/v${version}/bitbottle_${asset.os}_${asset.arch}.tar.gz";
            sha256 = asset.sha256;
          };
        in {
          default = pkgs.stdenv.mkDerivation {
            pname = "bitbottle";
            inherit version;
            src = tarball;
            sourceRoot = ".";
            dontConfigure = true;
            dontBuild = true;
            dontStrip = true;
            installPhase = ''
              mkdir -p $out/bin
              install -m755 bitbottle $out/bin/bitbottle
            '';
            meta = with pkgs.lib; {
              description = "bitbottle — gh-style Bitbucket CLI for Server/DC + Cloud";
              homepage = "https://github.com/proggarapsody/bitbottle";
              license = licenses.mit;
              platforms = systems;
              mainProgram = "bitbottle";
            };
          };
        });

      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/bitbottle";
        };
      });
    };
}
