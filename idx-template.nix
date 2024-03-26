{ pkgs, useLangChain ? false, ... }: {
  packages = [ pkgs.go ];
  bootstrap = ''
    cp -rf ${./.}/go-gemini "$WS_NAME"
    chmod -R +w "$WS_NAME"
    ${if useLangChain then "cp -rf ${./.}/langchain-overlay/* $WS_NAME" else "" }
    mv "$WS_NAME" "$out"
  '';
}
