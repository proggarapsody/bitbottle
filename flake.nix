{
  description = "bitbottle — gh-style Bitbucket CLI for Server/DC + Cloud";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";

  outputs = { self, nixpkgs }:
    let
      version = "1.56.1";
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;

      assetFor = system:
        let
          map = {
            "x86_64-linux"   = { os = "linux";  arch = "amd64"; sha256 = "9cd0b7985011355adaf0ba70d6464239d5f1ed5887f1ce2b299d089d5d327fd2"; };
            "aarch64-linux"  = { os = "linux";  arch = "arm64"; sha256 = "d339894423e1f22f23bcebf45ae309f48d305fc475ce09dd5793ab78f6d23bbe"; };
            "x86_64-darwin"  = { os = "darwin"; arch = "amd64"; sha256 = "86f4b4e33da95daaf4904432bbdea778ad57fe1e4594c7dd7a80df24630a6848"; };
            "aarch64-darwin" = { os = "darwin"; arch = "arm64"; sha256 = "78b7c8844fe8c631a4d72208b71c1d2183b4759b3d51dcabdedaaf74f97e1d30"; };
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
