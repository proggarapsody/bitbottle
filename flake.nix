{
  description = "bitbottle — gh-style Bitbucket CLI for Server/DC + Cloud";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";

  outputs = { self, nixpkgs }:
    let
      version = "1.61.0";
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;

      assetFor = system:
        let
          map = {
            "x86_64-linux"   = { os = "linux";  arch = "amd64"; sha256 = "fc0766c50e546f46c74f2d48af654de100cc01e78202c63b7ba57d11fdb163a6"; };
            "aarch64-linux"  = { os = "linux";  arch = "arm64"; sha256 = "385d8c8c12f0b5b0021f1805bb800aeb54cd33be0021225e1bd8107f19b61eb2"; };
            "x86_64-darwin"  = { os = "darwin"; arch = "amd64"; sha256 = "2c0374efab841003dacd72e7f129d7913b34eb84200a477d6158d1102200ffb7"; };
            "aarch64-darwin" = { os = "darwin"; arch = "arm64"; sha256 = "feeb91e3109ebb032bebc487ab8c04c8422dfa8f7441651c4e80d580e2f5186a"; };
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
