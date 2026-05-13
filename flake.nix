{
  description = "bitbottle — gh-style Bitbucket CLI for Server/DC + Cloud";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";

  outputs = { self, nixpkgs }:
    let
      version = "1.60.0";
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;

      assetFor = system:
        let
          map = {
            "x86_64-linux"   = { os = "linux";  arch = "amd64"; sha256 = "5e06f2adb34f50c3dafa9e645b1c71c3942625d7204e0d6b22db7b2ff12f77c1"; };
            "aarch64-linux"  = { os = "linux";  arch = "arm64"; sha256 = "6b6c2ce9e112bde477ccce210e335e26d59c5f4b94927ebbda665a3caa4d9714"; };
            "x86_64-darwin"  = { os = "darwin"; arch = "amd64"; sha256 = "4cba7868eeade521f077407d3d4e91ee3babf5e3aff1f5227538da1aa1b5dd1f"; };
            "aarch64-darwin" = { os = "darwin"; arch = "arm64"; sha256 = "92166f74dc0b183d1dee4038229e7b3b7dc22b9c3c42a2f24c941fe6f36fe599"; };
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
