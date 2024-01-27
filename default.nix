{ lib
, buildGoModule
}:
buildGoModule {
    name = "hyprboard";
    src = ./.;
    vendorHash = "sha256-SW1ZeVBJKMWMK3UMq5N+VKhILnusTVdYuIQW26/iRYU=";
}