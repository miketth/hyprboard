{ lib
, buildGoModule
}:
buildGoModule {
    name = "hyprboard";
    src = ./.;
    vendorHash = null;
}