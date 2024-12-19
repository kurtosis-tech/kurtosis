{ lib
, buildGoModule
, writeTextFile
}:

buildGoModule rec {
  pname = "kurtosis";
  version = "0.89.11";

  src = ./.;

  proxyVendor = true;
  vendorHash = "sha256-GaEIitoRiuYxtS7cDKobFyIlraDNQjcvbRvzG3nUKFU=";

  postPatch =
    let
      kurtosisVersion = writeTextFile {
        name = "kurtosis_verion.go";
        text = ''
          package kurtosis_version
          const (
            KurtosisVersion = "${version}"
          )
        '';
      };
    in
    ''
      ln -s ${kurtosisVersion} kurtosis_version/kurtosis_version.go
    '';

  # disable checks temporarily since they connect to the internet
  # namely user_support_constants_test.go
  doCheck = false;

  # keep this for future reference
  preCheck = ''
    # some tests in commands use XDG home related environment variables
    export HOME=/tmp
  '';

  postInstall = ''
    mv $out/bin/cli $out/bin/kurtosis
    mv $out/bin/files_artifacts_expander $out/bin/files-artifacts-expander
    mv $out/bin/api_container $out/bin/api-container

  '';

  meta = with lib; {
    description = "A platform for launching an ephemeral Ethereum backend";
    mainProgram = "kurtosis";
    homepage = "https://github.com/kurtosis-tech/kurtosis";
    license = licenses.asl20;
  };
}
