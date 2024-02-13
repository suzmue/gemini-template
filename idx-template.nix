{ pkgs, ... }: {
  packages = [ pkgs.go ];
  bootstrap = ''
    cp -rf ${./.}/gemini "$WS_NAME"
    chmod -R +w "$WS_NAME"
    mv "$WS_NAME" "$out"
  '';
}
