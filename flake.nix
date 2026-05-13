{
  description = "bitbottle — gh-style Bitbucket CLI for Server/DC + Cloud";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";

  outputs = { self, nixpkgs }:
    let
      version = "1.59.0";
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;

      assetFor = system:
        let
          map = {
            "x86_64-linux"   = { os = "linux";  arch = "amd64"; sha256 = "1d31db6faca1ef09637728fe10b0c1f971fc0cde430ec3c63ed938905ab84eb5"; };
            "aarch64-linux"  = { os = "linux";  arch = "arm64"; sha256 = "b95e30ae6d60ce1efbf0ce47fc94f98d8cb145ec4912ca0e3f39f2a80e5c3cc9"; };
            "x86_64-darwin"  = { os = "darwin"; arch = "amd64"; sha256 = "d07663eb889c1fb766a6c5eb56cf9e159b34ee27f51003d2a9e0b28d6e16af3a"; };
            "aarch64-darwin" = { os = "darwin"; arch = "arm64"; sha256 = "19d875ed79e67ebce3173e5c6a5e6f843645d42eca3d4a5e073ba76b2714f649"; };
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
