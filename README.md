# seal-256-cfb

## seal-256-cfb is a CLI program which implements the [SeaLion Block Cipher](https://github.com/sid-sun/sealion) in CFB (cipher feedback) mode with 256-Bit length keys using SHA3-256 with files.

### The program uses buffered channels and seperate goroutines for reading chunks from file, processing them and writing them. The buffer size depends on the machine and is calculated based on processing speed of SeaLion.

## Usage:

```
To encrypt: seal-256-cfb (--encrypt / -e) <input file> <passphrase file> <output file (optional)>

To decrypt: seal-256-cfb (--decrypt / -d) <encrypted input> <passphrase file> <output file (optional)>

To get version number: seal-256-cfb (--version / -v)

To get help: seal-256-cfb (--help / -h)
```

## Installation:

### Compiled Binaries: 
> [Linux amd64](https://cdn.sidsun.com/seal-256-cfb/seal-256-cfb_linux-amd64)

> [Darwin amd64](https://cdn.sidsun.com/seal-256-cfb/seal-256-cfb_darwin-amd64)

> [Windows amd64](https://cdn.sidsun.com/seal-256-cfb/seal-256-cfb_windows-amd64.exe)

### Debian Packages:

> [amd64](https://cdn.sidsun.com/seal-256-cfb/seal-256-cfb_amd64.deb)

### Use YAPPA (Yet Another PPA) :

```bash
curl -s --compressed "https://sid-sun.github.io/yappa/KEY.gpg" | sudo apt-key add -
curl -s --compressed "https://sid-sun.github.io/yappa/yappa.list" | sudo tee /etc/apt/sources.list.d/yappa.list
sudo apt update
sudo apt install seal-256-cfb
```

## Versioning system:

The Versioning system follows a Trickle-down approach (i.e. the version part after the updated part is to be set to 0s)

The version number consists of three parts:

1. Major 

    Major version is to be updated when using the SAME input and key, the output generated differs (ex: bug fixes)

2. Minor

    Minor version is to be updated when features are added or change are made to the core system which don't affect how it behaves with the same inputs (ex: performance improvements)

3. Infant

    Infant version is to be changed when the change doesn't affect the core system (ex: UX updates)


Updating on major and minor version changes is highly recommended.

Cheers!