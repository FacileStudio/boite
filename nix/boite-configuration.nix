{ config, pkgs, lib, ... }:

let
  tiroirVersion = "0.2.0";
  tiroir = pkgs.stdenv.mkDerivation {
    pname = "tiroir";
    version = tiroirVersion;
    src = pkgs.fetchurl {
      url = "https://github.com/FacileStudio/tiroir/releases/download/v${tiroirVersion}/tiroir_${tiroirVersion}_linux_amd64.tar.gz";
      sha256 = "288e086f8149ff6ad215f13bf1c111af7a6bbc828d12f3148e35c41e7fee5d96";
    };
    sourceRoot = ".";
    installPhase = ''
      mkdir -p $out/bin
      install -m 0755 tiroir $out/bin/tiroir
    '';
  };
in
{
  networking.hostName = "boite";
  networking.useDHCP = lib.mkDefault true;

  users.groups.boite = {};
  users.users.boite = {
    isNormalUser = true;
    home = "/home/boite";
    group = "boite";
    extraGroups = [ "wheel" ];
  };

  security.sudo.wheelNeedsPassword = false;

  services.openssh = {
    enable = true;
    ports = [ 22 ];
    settings = {
      PermitRootLogin = "prohibit-password";
      PasswordAuthentication = false;
    };
  };

  environment.systemPackages = [
    pkgs.curl
    pkgs.git
    pkgs.util-linux
    tiroir
  ];

  environment.interactiveShellInit = ''
    export WORKSPACE=/workspace
    if command -v tiroir >/dev/null 2>&1; then
      eval "$(tiroir export)"
    fi
  '';

  systemd.tmpfiles.rules = [
    "d /workspace 0755 boite boite -"
    "d /var/lib/boite 0755 root root -"
    "d /home/boite/.ssh 0700 boite boite -"
  ];

  systemd.services.boite-firstboot = {
    description = "Boite firstboot provisioning";
    wantedBy = [ "multi-user.target" ];
    before = [ "sshd.service" ];
    unitConfig = {
      ConditionPathExists = "!/var/lib/boite/firstboot.done";
    };
    serviceConfig = {
      Type = "oneshot";
      RemainAfterExit = false;
    };
    path = [ pkgs.util-linux pkgs.coreutils ];
    script = ''
      marker="/var/lib/boite/firstboot.done"
      if [ -f "$marker" ]; then
        exit 0
      fi

      mkdir -p /mnt/boitecfg
      blk=""
      label_dev="$(findfs LABEL=BOITECFG 2>/dev/null || true)"
      for attempt in 1 2 3 4 5 6; do
        for cand in "$label_dev" /dev/disk/by-label/BOITECFG /dev/vdb /dev/vdc /dev/sdb /dev/sdc /dev/sr0 /dev/sr1; do
          if [ -n "$cand" ] && [ -b "$cand" ] && mount -o ro "$cand" /mnt/boitecfg 2>/dev/null && [ -f /mnt/boitecfg/authorized_keys ]; then
            blk="$cand"
            break
          fi
          umount /mnt/boitecfg 2>/dev/null || true
        done
        if [ -n "$blk" ]; then
          break
        fi
        sleep 2
      done

      if [ -z "$blk" ] || [ ! -f /mnt/boitecfg/authorized_keys ]; then
        exit 1
      fi

      mkdir -p /home/boite/.ssh /var/lib/boite
      cp /mnt/boitecfg/authorized_keys /home/boite/.ssh/authorized_keys
      chown boite:boite /home/boite/.ssh/authorized_keys
      chmod 0600 /home/boite/.ssh/authorized_keys

      for f in /mnt/boitecfg/.tiroir*; do
        if [ -f "$f" ]; then
          fname="$(basename "$f")"
          cp "$f" "/home/boite/$fname"
          chown boite:boite "/home/boite/$fname"
          chmod 0600 "/home/boite/$fname"
        fi
      done

      umount /mnt/boitecfg
      touch "$marker"
    '';
  };

  virtualisation.diskSize = lib.mkDefault (20 * 1024);
  system.stateVersion = "24.11";
}
